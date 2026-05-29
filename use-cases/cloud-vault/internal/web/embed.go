package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:static
var staticFiles embed.FS

// NewHandler returns an http.Handler that serves the embedded SPA.
// API routes should be registered before this handler in the router.
// Any path not matching a static file falls back to index.html.
func NewHandler() http.Handler {
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic("web: failed to sub static embed: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(static))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip the wildcard path prefix if Plumego injects it.
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if the file exists; fall back to index.html for SPA routing.
		f, err := static.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Serve index.html so client-side routing works.
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	})
}
