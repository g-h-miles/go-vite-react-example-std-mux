// middleware/spa.go
package middleware

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path" // Use 'path' for virtual FS paths, not 'filepath'
	"strings"
)

type SPAConfig struct {
	// The filesystem containing your built frontend assets.
	// Must be populated with a //go:embed directive.
	DistFS embed.FS

	// The path within the embed.FS to the built assets.
	// Example: "frontend/dist"
	DistPath string

	// The name of the index file, defaults to "index.html".
	IndexFile string

	// --- Optional fields with sensible defaults ---
	DevProxyURL string
	APIPrefix   string
	DevEnvVar   string
	DevEnvValue string
}

// NewSPAHandler creates a new handler for a Single Page Application.
// It returns an error if the configuration is invalid for production mode.
func NewSPAHandler(config SPAConfig) (http.HandlerFunc, error) {
	// --- Set defaults for optional fields ---
	if config.DistPath == "" {
		config.DistPath = "frontend/dist"
	}
	if config.IndexFile == "" {
		config.IndexFile = "index.html"
	}
	if config.DevProxyURL == "" {
		config.DevProxyURL = "http://localhost:5173"
	}
	if config.APIPrefix == "" {
		config.APIPrefix = "/api"
	}
	if config.DevEnvVar == "" {
		config.DevEnvVar = "ENV"
	}
	if config.DevEnvValue == "" {
		config.DevEnvValue = "dev"
	}

	// --- THE CRUCIAL VALIDATION STEP ---
	// Only validate if we are in production mode. This allows developers
	// to run the backend in dev mode without needing to build the frontend first.
	isDevMode := os.Getenv(config.DevEnvVar) == config.DevEnvValue
	// if !isDevMode {
	// 	// Check that the DistFS is not empty and contains the index file.
	// 	// This is our guardrail against a missing //go:embed directive.
	// 	indexPath := path.Join(config.DistPath, config.IndexFile)
	// 	f, err := config.DistFS.Open(indexPath)
	// 	if err != nil {
	// 		// Return a clear, actionable error.
	// 		return nil, fmt.Errorf(
	// 			"SPA assets misconfigured: could not find '%s' in the embedded filesystem. "+
	// 				"Ensure your //go:embed directive is correct and includes the frontend build output.",
	// 			indexPath,
	// 		)
	// 	}
	// 	f.Close() // Don't forget to close the file.
	// }

	// --- MODIFIED VALIDATION LOGIC ---
	if isDevMode {
		// In DEV mode, we CHECK but only WARN if files are missing.
		indexPath := path.Join(config.DistPath, config.IndexFile)
		if _, err := config.DistFS.Open(indexPath); err != nil {
			log.Printf(
				"WARN: SPA assets not found at '%s'. This is okay in dev mode, as requests will be proxied to '%s'. "+
					"However, this will be a FATAL ERROR in production builds.",
				indexPath,
				config.DevProxyURL,
			)
		}
	} else {
		// In PROD mode, we CHECK and return a FATAL ERROR if files are missing.
		indexPath := path.Join(config.DistPath, config.IndexFile)
		f, err := config.DistFS.Open(indexPath)
		if err != nil {
			return nil, fmt.Errorf(
				"SPA assets misconfigured: could not find '%s' in the embedded filesystem. "+
					"Ensure your //go:embed directive is correct and includes the frontend build output.",
				indexPath,
			)
		}
		f.Close()
	}

	// --- Create the handler using the now-validated config ---
	// We pre-calculate the sub-filesystem for production mode here.
	distDirFS, err := fs.Sub(config.DistFS, config.DistPath)
	if err != nil {
		// This error is less likely but still possible if DistPath is wrong.
		return nil, fmt.Errorf("failed to create sub-filesystem for dist path '%s': %w", config.DistPath, err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, config.APIPrefix) {
			// This should be handled by your API router, but as a fallback,
			// we can just let it pass through. A real router is better.
			// For this example, we assume another handler will take care of it.
			return
		}

		if isDevMode {
			handleDevMode(w, r, config.DevProxyURL)
		} else {
			handleProdMode(w, r, distDirFS)
		}
	}

	return handler, nil
}

func handleDevMode(w http.ResponseWriter, r *http.Request, proxyURL string) {
	target, err := url.Parse(proxyURL)
	if err != nil {
		log.Printf("Failed to parse dev proxy URL: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	httputil.NewSingleHostReverseProxy(target).ServeHTTP(w, r)
}

func handleProdMode(w http.ResponseWriter, r *http.Request, distFS fs.FS) {
	// The path must be relative to the root of the sub-filesystem.
	reqPath := strings.TrimPrefix(r.URL.Path, "/")

	// Try to serve a static file from the filesystem.
	f, err := distFS.Open(reqPath)
	if err == nil { // File exists
		defer f.Close()
		// Use the default file server to handle content types, etc.
		http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
		return
	}

	// If the file does not exist, fall back to serving the index file.
	// This is the key behavior for SPAs.
	r.URL.Path = "/" // Serve the root, which will be index.html
	http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
}
