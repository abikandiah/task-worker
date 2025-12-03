package server

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

const staticFolder = "./dist"

// UI and static file server
func (server *Server) setupStaticFileServer() func(r chi.Router) {
	// Get true mod time of index.html for caching
	indexFilePath := filepath.Join(staticFolder, "index.html")
	info, err := os.Stat(indexFilePath)
	if err != nil {
		slog.Error(
			"failed to read index.html stat",
			slog.String("path", indexFilePath),
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	indexModTime := info.ModTime()

	staticDir := http.Dir(staticFolder)
	staticFSHandler := http.StripPrefix("/", http.FileServer(staticDir))

	return func(r chi.Router) {
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Strip and clean path to prevent traversal attacks
			path := strings.TrimPrefix(r.URL.Path, "/")
			path = filepath.Clean(path)

			// Check if the requested path corresponds to a file that exists
			if _, err := staticDir.Open(path); err == nil {
				// Set short cache life for files without filename hashes
				if strings.HasPrefix(path, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else if path != "index.html" {
					w.Header().Set("Cache-Control", "public, max-age=600")
				}

				staticFSHandler.ServeHTTP(w, r)
				return
			}

			// Catch-All Handler (SPA Fallback)
			// Serve index.html so React Router can take over
			indexFile, err := staticDir.Open("index.html")
			if err != nil {
				http.Error(w, "Could not open index.html", http.StatusInternalServerError)
				return
			}

			// Set the content type and serve the index.html content
			// Cache-Control: no-cache ensures the browser checks the server (using If-Modified-Since)
			w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(w, r, "index.html", indexModTime, indexFile)
		})
	}
}
