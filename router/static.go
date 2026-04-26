package router

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spcent/plumego/contract"
)

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
	rc := contract.RequestContextFromContext(req.Context())
	relPath, ok := rc.Params["filepath"]
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

	if hasParentTraversal(cleanPath) {
		return "", false
	}

	if filepath.IsAbs(cleanPath) {
		return "", false
	}

	if strings.HasPrefix(cleanPath, "/") {
		return "", false
	}

	return cleanPath, true
}

func hasParentTraversal(path string) bool {
	if path == ".." {
		return true
	}
	for _, part := range strings.Split(path, string(filepath.Separator)) {
		if part == ".." {
			return true
		}
	}
	return false
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
//	err := r.Static("/static", "./public")
//	GET /static/js/app.js → ./public/js/app.js
func (r *Router) Static(prefix, dir string) error {
	return r.registerStaticRoute(normalizeStaticPrefix(prefix), http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		serveFromDirectory(w, req, dir)
	}))
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
//	err := r.StaticFS("/assets", http.FS(public))
//	GET /assets/index.html → served from embedded FS
func (r *Router) StaticFS(prefix string, fs http.FileSystem) error {
	return r.registerStaticRoute(normalizeStaticPrefix(prefix), http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		serveFromFileSystem(w, req, fs)
	}))
}

// serveFromDirectory serves files from a local directory
func serveFromDirectory(w http.ResponseWriter, req *http.Request, dir string) {
	cleanPath, ok := getFilePathFromRequest(req)
	if !ok {
		http.NotFound(w, req)
		return
	}

	// Construct the full path inside the given directory
	fullPath := filepath.Join(dir, cleanPath)

	// Security: verify the resolved path is within the root directory (prevents symlink escape)
	if !isPathWithinRoot(dir, fullPath) {
		http.NotFound(w, req)
		return
	}

	// Check if the file exists
	if handleStaticFileError(w, req, checkFileExists(fullPath)) {
		return
	}

	// Serve the file
	http.ServeFile(w, req, fullPath)
}

// serveFromFileSystem serves files from a custom http.FileSystem
func serveFromFileSystem(w http.ResponseWriter, req *http.Request, fs http.FileSystem) {
	cleanPath, ok := getFilePathFromRequest(req)
	if !ok {
		http.NotFound(w, req)
		return
	}

	// Open the file
	f, err := fs.Open(cleanPath)
	if handleStaticFileError(w, req, err) {
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if handleStaticFileError(w, req, err) {
		return
	}

	serveFileContent(w, req, f, info)
}

// checkFileExists checks if a file exists and returns an error if it doesn't
func checkFileExists(path string) error {
	_, err := os.Stat(path)
	return err
}

// isPathWithinRoot verifies that resolvedPath is within rootDir after resolving symlinks.
// This prevents symlink-based directory traversal attacks.
func isPathWithinRoot(rootDir, resolvedPath string) bool {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return false
	}
	// Resolve symlinks on root
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		// If root itself doesn't exist, deny
		return false
	}

	absPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return false
	}
	// Resolve symlinks on the target path
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// File doesn't exist yet (checked later), allow the check to pass
		// so the caller can return a proper 404
		return true
	}

	// Ensure the resolved path is within the root
	return strings.HasPrefix(realPath, realRoot+string(filepath.Separator)) || realPath == realRoot
}

// registerStaticRoute registers a GET route for static file primitives.
func (r *Router) registerStaticRoute(prefix string, handler http.Handler) error {
	routePath := prefix + "/*filepath"
	return r.AddRoute(http.MethodGet, routePath, handler)
}

// serveFileContent serves file content with seeking support, using ServeContent
// when possible, falling back to raw copy for non-seekable files.
func serveFileContent(w http.ResponseWriter, req *http.Request, f http.File, info os.FileInfo) {
	if seeker, ok := f.(io.ReadSeeker); ok {
		http.ServeContent(w, req, info.Name(), info.ModTime(), seeker)
	} else {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		_, _ = io.Copy(w, f)
	}
}
