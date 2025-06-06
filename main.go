package main

import (
	"fmt"
	"net/http"

	"github.com/g-h-miles/go-vite-react-example-std-mux/frontend"
)

func main() {
	// Create a new echo server
	// e := echo.New()
	m := http.NewServeMux()

	// Setup the frontend handlers to service vite static assets
	frontend.RegisterHandlers(m)

	// Add standard middleware
	// handler := logger(m)

	// Setup the API Group
	// api := m.Group("/api")

	// Basic APi endpoint
	m.HandleFunc("/api/message", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello, from the golang World! - updated- update again!! i donnnnoooo"}`))
	})

	fmt.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", m); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}

	err := http.ListenAndServe(":3000", logger(m))
	if err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request received", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
