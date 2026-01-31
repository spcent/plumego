package router

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
)

// StaticConfig holds configuration for static file serving
type StaticConfig struct {
	// Prefix is the URL prefix for serving static files
	Prefix string

	// Root is the directory path or filesystem to serve from
	Root any // Can be string (directory) or http.FileSystem

	// CacheControl sets the Cache-Control header value
	// Example: "public, max-age=31536000" for 1 year
	CacheControl string

	// EnableETag enables ETag header generation for cache validation
	EnableETag bool

	// IndexFile is the default file to serve for directory requests
	// Default: "index.html"
	IndexFile string

	// SPAFallback enables SPA mode where missing files fall back to IndexFile
	SPAFallback bool

	// AllowedExtensions limits which file extensions can be served
	// If empty, all extensions are allowed
	AllowedExtensions []string
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
		Prefix:    normalizeStaticPrefix(prefix),
		Root:      dir,
		IndexFile: "index.html",
	}

	r.registerStaticRoute(config, serveFromDirectory)
}

// StaticWithConfig registers a route that serves files with custom configuration.
//
// Example:
//
//	r.StaticWithConfig(router.StaticConfig{
//	    Prefix:       "/static",
//	    Root:         "./public",
//	    CacheControl: "public, max-age=86400",
//	    EnableETag:   true,
//	})
func (r *Router) StaticWithConfig(config StaticConfig) {
	config.Prefix = normalizeStaticPrefix(config.Prefix)
	if config.IndexFile == "" {
		config.IndexFile = "index.html"
	}

	if _, ok := config.Root.(http.FileSystem); ok {
		r.registerStaticRouteWithConfig(config, serveFromFileSystemWithConfig)
	} else {
		r.registerStaticRouteWithConfig(config, serveFromDirectoryWithConfig)
	}
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
		Prefix:    normalizeStaticPrefix(prefix),
		Root:      fs,
		IndexFile: "index.html",
	}

	r.registerStaticRoute(config, serveFromFileSystem)
}

// StaticSPA registers a route for Single Page Applications.
// When a file is not found, it falls back to the index file.
//
// Example:
//
//	r.StaticSPA("/", "./dist", "index.html")
//	// GET /about -> serves ./dist/index.html (SPA handles routing)
func (r *Router) StaticSPA(prefix, dir, indexFile string) {
	config := StaticConfig{
		Prefix:      normalizeStaticPrefix(prefix),
		Root:        dir,
		IndexFile:   indexFile,
		SPAFallback: true,
	}

	r.registerStaticRouteWithConfig(config, serveFromDirectoryWithConfig)
}

// serveFromDirectory serves files from a local directory
func serveFromDirectory(w http.ResponseWriter, req *http.Request, root any) bool {
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
func serveFromFileSystem(w http.ResponseWriter, req *http.Request, root any) bool {
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
func (r *Router) registerStaticRoute(config StaticConfig, handler func(http.ResponseWriter, *http.Request, any) bool) {
	routePath := config.Prefix + "/*filepath"

	r.GetFunc(routePath, func(w http.ResponseWriter, req *http.Request) {
		handler(w, req, config.Root)
	})
}

// registerStaticRouteWithConfig registers static file routes with full configuration
func (r *Router) registerStaticRouteWithConfig(config StaticConfig, handler func(http.ResponseWriter, *http.Request, StaticConfig) bool) {
	routePath := config.Prefix + "/*filepath"

	r.GetFunc(routePath, func(w http.ResponseWriter, req *http.Request) {
		handler(w, req, config)
	})
}

// serveFromDirectoryWithConfig serves files from a local directory with configuration
func serveFromDirectoryWithConfig(w http.ResponseWriter, req *http.Request, config StaticConfig) bool {
	dir, ok := config.Root.(string)
	if !ok {
		return false
	}

	cleanPath, ok := getFilePathFromRequest(req)
	if !ok {
		if config.SPAFallback {
			serveStaticFile(w, req, filepath.Join(dir, config.IndexFile), config)
			return true
		}
		http.NotFound(w, req)
		return true
	}

	// Check allowed extensions
	if len(config.AllowedExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(cleanPath))
		allowed := false
		for _, allowedExt := range config.AllowedExtensions {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			http.NotFound(w, req)
			return true
		}
	}

	fullPath := filepath.Join(dir, cleanPath)

	// Check if file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if config.SPAFallback {
			serveStaticFile(w, req, filepath.Join(dir, config.IndexFile), config)
			return true
		}
		http.NotFound(w, req)
		return true
	}

	// If it's a directory, try to serve index file
	if info.IsDir() {
		indexPath := filepath.Join(fullPath, config.IndexFile)
		if _, err := os.Stat(indexPath); err == nil {
			fullPath = indexPath
		} else if config.SPAFallback {
			serveStaticFile(w, req, filepath.Join(dir, config.IndexFile), config)
			return true
		} else {
			http.NotFound(w, req)
			return true
		}
	}

	serveStaticFile(w, req, fullPath, config)
	return true
}

// serveStaticFile serves a single static file with cache headers
func serveStaticFile(w http.ResponseWriter, req *http.Request, path string, config StaticConfig) {
	f, err := os.Open(path)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		http.NotFound(w, req)
		return
	}

	// Set cache headers
	if config.CacheControl != "" {
		w.Header().Set("Cache-Control", config.CacheControl)
	}

	// Generate and set ETag if enabled
	if config.EnableETag {
		etag := generateETag(info)
		w.Header().Set("ETag", etag)

		// Check If-None-Match
		if match := req.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	http.ServeContent(w, req, info.Name(), info.ModTime(), f)
}

// serveFromFileSystemWithConfig serves files from http.FileSystem with configuration
func serveFromFileSystemWithConfig(w http.ResponseWriter, req *http.Request, config StaticConfig) bool {
	fs, ok := config.Root.(http.FileSystem)
	if !ok {
		return false
	}

	cleanPath, ok := getFilePathFromRequest(req)
	if !ok {
		if config.SPAFallback {
			serveFromFS(w, req, fs, config.IndexFile, config)
			return true
		}
		http.NotFound(w, req)
		return true
	}

	// Check allowed extensions
	if len(config.AllowedExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(cleanPath))
		allowed := false
		for _, allowedExt := range config.AllowedExtensions {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			http.NotFound(w, req)
			return true
		}
	}

	// Try to open the file
	f, err := fs.Open(cleanPath)
	if err != nil {
		if config.SPAFallback {
			serveFromFS(w, req, fs, config.IndexFile, config)
			return true
		}
		http.NotFound(w, req)
		return true
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		http.NotFound(w, req)
		return true
	}

	// If it's a directory, try index file
	if info.IsDir() {
		indexPath := cleanPath + "/" + config.IndexFile
		if indexF, err := fs.Open(indexPath); err == nil {
			indexF.Close()
			serveFromFS(w, req, fs, indexPath, config)
			return true
		} else if config.SPAFallback {
			serveFromFS(w, req, fs, config.IndexFile, config)
			return true
		}
		http.NotFound(w, req)
		return true
	}

	serveFromFS(w, req, fs, cleanPath, config)
	return true
}

// serveFromFS serves a file from http.FileSystem with cache headers
func serveFromFS(w http.ResponseWriter, req *http.Request, fs http.FileSystem, path string, config StaticConfig) {
	f, err := fs.Open(path)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		http.NotFound(w, req)
		return
	}

	// Set cache headers
	if config.CacheControl != "" {
		w.Header().Set("Cache-Control", config.CacheControl)
	}

	// Generate and set ETag if enabled
	if config.EnableETag {
		etag := generateETag(info)
		w.Header().Set("ETag", etag)

		// Check If-None-Match
		if match := req.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// For files that support seeking
	if seeker, ok := f.(io.ReadSeeker); ok {
		http.ServeContent(w, req, info.Name(), info.ModTime(), seeker)
	} else {
		// Fallback for non-seekable files
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		io.Copy(w, f)
	}
}

// generateETag generates an ETag based on file info
func generateETag(info os.FileInfo) string {
	// Use file size and modification time for ETag
	data := info.Name() + strconv.FormatInt(info.Size(), 10) + info.ModTime().Format(time.RFC3339Nano)
	hash := md5.Sum([]byte(data))
	return `"` + hex.EncodeToString(hash[:]) + `"`
}
