package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spcent/plumego/contract"
)

// StaticConfig holds configuration for static file serving
type StaticConfig struct {
	// Prefix is the URL prefix for serving static files
	Prefix string

	// Root is the directory path or filesystem to serve from
	Root interface{} // Can be string (directory) or http.FileSystem
}

// normalizeStaticPrefix ensures the prefix always starts with "/"
func normalizeStaticPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "/"
	}
	if prefix[0] != '/' {
		return "/" + prefix
	}
	return prefix
}

// getFilePathFromRequest extracts and cleans the file path from request
func getFilePathFromRequest(req *http.Request) (string, bool) {
	relPath, ok := contract.Param(req, "filepath")
	if !ok {
		return "", false
	}

	// Security: reject empty paths
	if relPath == "" {
		return "", false
	}

	// Security: reject paths with null bytes
	if strings.Contains(relPath, "\x00") {
		return "", false
	}

	// Clean the relative path to avoid directory traversal (e.g., "../../etc/passwd")
	cleanPath := filepath.Clean(relPath)

	// Additional security checks
	// 1. Ensure the cleaned path doesn't contain ".."
	if strings.Contains(cleanPath, "..") {
		return "", false
	}

	// 2. Reject absolute paths
	if filepath.IsAbs(cleanPath) {
		return "", false
	}

	// 3. Reject paths starting with "/" (shouldn't happen after Clean, but be safe)
	if strings.HasPrefix(cleanPath, "/") {
		return "", false
	}

	return cleanPath, true
}

// handleStaticFileError provides consistent error handling for static file operations
func handleStaticFileError(w http.ResponseWriter, req *http.Request, err error) bool {
	if err != nil {
		http.NotFound(w, req)
		return true
	}
	return false
}

// Static registers a route that serves files from a local directory
// under the given URL prefix.
//
// Example:
//
//	r.Static("/static", "./public")
//	GET /static/js/app.js → ./public/js/app.js
func (r *Router) Static(prefix, dir string) {
	config := StaticConfig{
		Prefix: normalizeStaticPrefix(prefix),
		Root:   dir,
	}

	r.registerStaticRoute(config, serveFromDirectory)
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
	config := StaticConfig{
		Prefix: normalizeStaticPrefix(prefix),
		Root:   fs,
	}

	r.registerStaticRoute(config, serveFromFileSystem)
}

// serveFromDirectory serves files from a local directory
func serveFromDirectory(w http.ResponseWriter, req *http.Request, root interface{}) bool {
	dir, ok := root.(string)
	if !ok {
		return false
	}

	cleanPath, ok := getFilePathFromRequest(req)
	if !ok {
		http.NotFound(w, req)
		return true
	}

	// Construct the full path inside the given directory
	fullPath := filepath.Join(dir, cleanPath)

	// Check if the file exists
	if handleStaticFileError(w, req, checkFileExists(fullPath)) {
		return true
	}

	// Serve the file
	http.ServeFile(w, req, fullPath)
	return true
}

// serveFromFileSystem serves files from a custom http.FileSystem
func serveFromFileSystem(w http.ResponseWriter, req *http.Request, root interface{}) bool {
	fs, ok := root.(http.FileSystem)
	if !ok {
		return false
	}

	cleanPath, ok := getFilePathFromRequest(req)
	if !ok {
		http.NotFound(w, req)
		return true
	}

	// Try to open the file
	f, err := fs.Open(cleanPath)
	if handleStaticFileError(w, req, err) {
		return true
	}
	defer f.Close()

	// Use http.FileServer logic by stripping the prefix
	http.FileServer(fs).ServeHTTP(w, req)
	return true
}

// checkFileExists checks if a file exists and returns an error if it doesn't
func checkFileExists(path string) error {
	_, err := os.Stat(path)
	return err
}

// registerStaticRoute is a generic function that registers static file routes
func (r *Router) registerStaticRoute(config StaticConfig, handler func(http.ResponseWriter, *http.Request, interface{}) bool) {
	routePath := config.Prefix + "/*filepath"

	r.GetFunc(routePath, func(w http.ResponseWriter, req *http.Request) {
		handler(w, req, config.Root)
	})
}
