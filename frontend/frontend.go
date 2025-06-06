package frontend

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
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
		panic(err)
	}
	distIndexHTML, err = fs.Sub(indexHTML, "dist")
	if err != nil {
		panic(err)
	}
}

func RegisterHandlers(mux *http.ServeMux) {
	if os.Getenv("ENV") == "dev" {
		log.Println("Running in dev mode")
		setupDevProxy(mux)
		return
	}

	mux.HandleFunc("/", spaHandler(distDirFS))
}

func setupDevProxy(mux *http.ServeMux) {
	target, err := url.Parse("http://localhost:5173")
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}
		proxy.ServeHTTP(w, r)
	})
}

func spaHandler(staticFS fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFS))
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}

		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" {
			p = "index.html"
		}

		if _, err := fs.Stat(staticFS, p); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		f, err := staticFS.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.Copy(w, f)
	}
}
