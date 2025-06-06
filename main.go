package main

import (
	"fmt"
	"net/http"

	"github.com/danhawkins/go-vite-react-example/frontend"
)

func main() {
	mux := http.NewServeMux()

	// Setup the frontend handlers to service vite static assets
	frontend.RegisterHandlers(mux)

	// Basic API endpoint
	mux.HandleFunc("/api/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"Hello, from the golang World! - updated"}`)
	})

	fmt.Println("Listening on :3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		fmt.Println(err)
	}
}
