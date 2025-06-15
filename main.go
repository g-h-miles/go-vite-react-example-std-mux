package main

import (
	"fmt"

	"github.com/g-h-miles/go-vite-react-example-std-mux/middleware"
	// "github.com/g-h-miles/go-vite-react-example-std-mux/frontend"

	"embed"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/g-h-miles/httpmux"
	_ "github.com/joho/godotenv/autoload"
)

var (
	//go:embed frontend/dist/*
	dist embed.FS

	distDirFS     fs.FS
	distIndexHTML fs.FS
)

func init() {
	// var err error
	// distDirFS, err = fs.Sub(dist, "frontend/dist")
	// if err != nil {
	// 	log.Fatal("Failed to create dist sub filesystem:", err)
	// }

	// distIndexHTML, err = fs.Sub(indexHTML, "frontend/dist")
	// if err != nil {
	// 	log.Fatal("Failed to create index.html sub filesystem:", err)
	// }
}

// func ensureFrontendBuild() {
// 	var err error
// 	distDirFS, err = fs.Sub(dist, "frontend/dist")
// 	if err != nil {
// 		log.Fatal("Failed to create dist sub filesystem:", err)
// 	}

// 	distIndexHTML, err = fs.Sub(indexHTML, "frontend/dist")
// 	if err != nil {
// 		log.Fatal("Failed to create index.html sub filesystem:", err)
// 	}
// }

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"message": "Hello, from the golang World! - updated- update again!! YESS!!"}`))
}

func main() {
	// ensureFrontendBuild()

	// Create a new echo server
	// e := echo.New()
	// m := http.NewServeMux()
	// m := httpmux.New()
	multi := httpmux.NewMultiRouter()

	// Setup the frontend handlers to service vite static assets
	// frontend.RegisterHandlers(m)

	// Add standard middleware
	// handler := logger(m)

	// Setup the API Group
	// api := m.Group("/api")

	spaHandler := middleware.SPA(middleware.SPAConfig{
		DistFS:    dist,
		DistPath:  "frontend/dist",
		IsDevMode: os.Getenv("ENV") == "dev",
	})

	spaRouter := httpmux.NewServeMux()
	spaRouter.Handle("GET", "/{theRest...}", spaHandler((nil)))

	multi.Default(spaRouter)

	// Basic APi endpoint
	apiRouter := httpmux.NewServeMux()

	apiRouter.HandleFunc("GET", "/message", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello, from the golang World! - updated- update again!! YESS!!"}`))
	})

	multi.Group("/api", apiRouter)

	fmt.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", logger(multi)); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request received", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func RegisterHandlers(mux *http.ServeMux) {
	if os.Getenv("ENV") == "dev" {
		log.Println("Running in dev mode")
		setupDevProxy(mux)
		// In dev mode, we still need to handle API routes
		mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
			// Let the main server handle API routes

		})
		return
	}
	// Use the static assets from the dist directory
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Skip if path starts with /api
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			return
		}
		// Try to serve static file first
		if _, err := fs.Stat(distDirFS, strings.TrimPrefix(r.URL.Path, "/")); err == nil {
			http.FileServer(http.FS(distDirFS)).ServeHTTP(w, r)
			return
		}
		// Fallback to index.html for SPA routing
		http.FileServer(http.FS(distIndexHTML)).ServeHTTP(w, r)
	})
}

func setupDevProxy(mux *http.ServeMux) {
	target, err := url.Parse("http://localhost:5173")
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Handle all requests
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If it's an API request, don't proxy to Vite
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			return
		}
		proxy.ServeHTTP(w, r)
	})
}
