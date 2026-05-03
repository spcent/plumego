package router

import (
	"errors"
	"fmt"
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
	if prefix == "" || prefix == "/" {
		return "/"
	}
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}
	prefix = strings.TrimRight(prefix, "/")
	if prefix == "" {
		return "/"
	}
	prefix = strings.TrimRight(prefix, "/")
	if prefix == "" {
		return "/"
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
	handler, err := newDirectoryHandler(prefix, dir)
	if err != nil {
		return err
	}
	return r.registerStaticRoute(normalizeStaticPrefix(prefix), handler)
}

// StaticFS registers a route that serves files from a custom http.FileSystem.
//
// This allows serving from embedded filesystems (e.g., embed.FS) or other backends.
//
// Example:
//
//	//go:embed public/*
//	var public embed.FS
//	sub, err := fs.Sub(public, "public")
//	if err != nil {
//	    return err
//	}
//
//	err := r.StaticFS("/assets", http.FS(sub))
//	GET /assets/index.html → served from embedded FS
func (r *Router) StaticFS(prefix string, fs http.FileSystem) error {
	if fs == nil {
		return fmt.Errorf("router static_fs %s: %w", prefix, errors.New("nil file system"))
	}
	return r.registerStaticRoute(normalizeStaticPrefix(prefix), http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		serveFromFileSystem(w, req, fs)
	}))
}

func newDirectoryHandler(prefix, dir string) (http.Handler, error) {
	root, err := resolveStaticRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("router static %s: %w", prefix, err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		serveFromDirectory(w, req, root)
	}), nil
}

func resolveStaticRoot(dir string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", errors.New("empty directory")
	}
	absRoot, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve directory: %w", err)
	}
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", fmt.Errorf("resolve directory symlinks: %w", err)
	}
	info, err := os.Stat(realRoot)
	if err != nil {
		return "", fmt.Errorf("stat directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", dir)
	}
	return realRoot, nil
}

// serveFromDirectory serves files from a local directory
func serveFromDirectory(w http.ResponseWriter, req *http.Request, root string) {
	cleanPath, ok := getFilePathFromRequest(req)
	if !ok {
		http.NotFound(w, req)
		return
	}

	fullPath := filepath.Join(root, cleanPath)

	// Security: verify the resolved path is within the root directory (prevents symlink escape)
	if !isPathWithinRoot(root, fullPath) {
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
	return strings.HasPrefix(realPath, rootDir+string(filepath.Separator)) || realPath == rootDir
}

// registerStaticRoute registers a GET route for static file primitives.
func (r *Router) registerStaticRoute(prefix string, handler http.Handler) error {
	if prefix == "/" {
		return r.AddRoute(http.MethodGet, "/*filepath", handler)
	}
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
