package frontend

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

var (
	//go:embed dist/*
	dist embed.FS

	//go:embed dist/index.html
	indexHTML embed.FS

	distDirFS     fs.FS
	distIndexHTML fs.FS
)

func init() {
	var err error
	distDirFS, err = fs.Sub(dist, "dist")
	if err != nil {
		log.Fatal("Failed to create dist sub filesystem:", err)
	}

	distIndexHTML, err = fs.Sub(indexHTML, "dist")
	if err != nil {
		log.Fatal("Failed to create index.html sub filesystem:", err)
	}
}

func RegisterHandlers(mux *http.ServeMux) {
	if os.Getenv("ENV") == "dev" {
		log.Println("Running in dev mode")
		setupDevProxy(mux)
		// In dev mode, we still need to handle API routes
		mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
			// Let the main server handle API routes
			return
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
