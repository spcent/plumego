package router

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/spcent/plumego/contract"
)

// Static registers a route that serves files from a local directory
// under the given URL prefix.
//
// Example:
//
//	r.Static("/static", "./public")
//	GET /static/js/app.js → ./public/js/app.js
func (r *Router) Static(prefix, dir string) {
	// Ensure prefix always starts with "/"
	if prefix == "" || prefix[0] != '/' {
		prefix = "/" + prefix
	}

	// Register a GET route with wildcard *filepath
	r.GetFunc(prefix+"/*filepath", func(w http.ResponseWriter, req *http.Request) {
		relPath, _ := contract.Param(req, "filepath")

		// Clean the relative path to avoid directory traversal (e.g., "../../etc/passwd")
		cleanPath := filepath.Clean(relPath)

		// Construct the full path inside the given directory
		fullPath := filepath.Join(dir, cleanPath)

		// Check if the file exists
		if _, err := os.Stat(fullPath); err != nil {
			http.NotFound(w, req)
			return
		}

		// Serve the file
		http.ServeFile(w, req, fullPath)
	})
}

// StaticFS registers a route that serves files from a custom http.FileSystem.
//
// This allows serving from embedded filesystems (e.g., embed.FS) or other backends.
//
// Example:
//
//	//go:embed public/*
//	var public embed.FS
//
//	r.StaticFS("/assets", http.FS(public))
//	GET /assets/index.html → served from embedded FS
func (r *Router) StaticFS(prefix string, fs http.FileSystem) {
	// Ensure prefix always starts with "/"
	if prefix == "" || prefix[0] != '/' {
		prefix = "/" + prefix
	}

	// Register a GET route with wildcard *filepath
	r.GetFunc(prefix+"/*filepath", func(w http.ResponseWriter, req *http.Request) {
		relPath, _ := contract.Param(req, "filepath")

		// Clean the relative path to avoid directory traversal
		cleanPath := filepath.Clean(relPath)

		f, err := fs.Open(cleanPath)
		if err != nil {
			http.NotFound(w, req)
			return
		}
		defer f.Close()

		// Use http.FileServer logic by stripping the prefix
		http.FileServer(fs).ServeHTTP(w, req)
	})
}
