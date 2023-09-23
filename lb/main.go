package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) GetNextInd() int {
	ind := int(atomic.AddUint64(&s.current, 1) % uint64(len(s.backends)))
	return ind
}

func (b *Backend) IsAlive() bool {
	var alive bool
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return alive
}

func (b *Backend) SetAlive() bool {
	var alive bool
	b.mux.Lock()
	alive = b.Alive
	b.mux.Unlock()
	return alive
}

func (s *ServerPool) GetNextPeer() *Backend {
	next := s.GetNextInd()
	for i := next; i < len(s.backends)+next; i++ {
		indx := i % len(s.backends)

		if s.backends[indx].IsAlive() {
			atomic.StoreUint64(&s.current, uint64(indx))
		}
		return s.backends[indx]
	}
	return nil
}

func loadbalance(w http.ResponseWriter, r *http.Request) {
	peer := serverPool.GetNextPeer()
	if peer != nil {
		log.Printf(peer.URL.String())
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

var serverPool ServerPool

func main() {
	port := 3030
	backends := "http://localhost:3031,http://localhost:3032,http://localhost:3033,http://localhost:3034"
	b := strings.Split(backends, ",")
	for _, cur := range b {
		serverUrl, _ := url.Parse(cur)
		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		newB := &Backend{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		}
		serverPool.AddBackend(newB)
		log.Printf("Configured server: %s\n", serverUrl)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(loadbalance),
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error:", err)
	}
}
