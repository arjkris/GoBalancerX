package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 3031, "Port to serve")
	flag.Parse()
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(helloHandler),
	}

	fmt.Printf("Server is running on port %s...\n", port)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from the backend!")
}
