package frontend

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spcent/plumego/router"
)

const defaultIndex = "index.html"

// Config defines how a built Node/Next.js frontend should be served.
// It is intentionally minimal to keep compatibility with static exports
// (e.g. `next export`, Vite/React build) and embedded assets via go:embed.
type Config struct {
	// Prefix is the URL prefix where the frontend is mounted.
	// Examples: "/" or "/app".
	Prefix string

	// IndexFile is returned as a fallback when the requested file
	// does not exist (SPA-style routing).
	IndexFile string

	// CacheControl optionally sets a Cache-Control header on successful responses.
	CacheControl string
}

// Option mutates a Config.
type Option func(*Config)

// WithPrefix sets the mount prefix for the frontend bundle.
func WithPrefix(prefix string) Option {
	return func(cfg *Config) {
		cfg.Prefix = prefix
	}
}

// WithIndex overrides the default index.html fallback filename.
func WithIndex(name string) Option {
	return func(cfg *Config) {
		cfg.IndexFile = name
	}
}

// WithCacheControl sets a Cache-Control header for served assets.
func WithCacheControl(header string) Option {
	return func(cfg *Config) {
		cfg.CacheControl = header
	}
}

// RegisterFromDir mounts a built frontend directory (e.g. Next.js `out/`)
// at the given prefix. If the directory is missing no routes are registered
// and an error is returned.
func RegisterFromDir(r *router.Router, dir string, opts ...Option) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("frontend directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("frontend path %q is not a directory", dir)
	}

	// Verify directory is readable
	if f, err := os.Open(dir); err != nil {
		return fmt.Errorf("frontend directory %q not readable: %w", dir, err)
	} else {
		f.Close()
	}

	return RegisterFS(r, http.Dir(dir), opts...)
}

// RegisterFS mounts a frontend bundle served from the provided http.FileSystem.
// This is suitable for go:embed bundles using http.FS.
func RegisterFS(r *router.Router, fsys http.FileSystem, opts ...Option) error {
	if fsys == nil {
		return errors.New("filesystem cannot be nil")
	}

	cfg := &Config{Prefix: "/", IndexFile: defaultIndex}
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate and normalize prefix
	cleanedPrefix, err := normalizePrefix(cfg.Prefix)
	if err != nil {
		return fmt.Errorf("invalid prefix %q: %w", cfg.Prefix, err)
	}

	// Validate index file
	indexFile := strings.TrimSpace(cfg.IndexFile)
	if indexFile == "" {
		indexFile = defaultIndex
	}
	if strings.Contains(indexFile, "/") || strings.Contains(indexFile, "\\") {
		return fmt.Errorf("index file %q cannot contain path separators", cfg.IndexFile)
	}

	// Create handler
	h := &handler{
		fs:           fsys,
		prefix:       cleanedPrefix,
		indexFile:    indexFile,
		cacheControl: strings.TrimSpace(cfg.CacheControl),
	}

	// Register routes based on prefix
	if cleanedPrefix == "/" {
		// Root prefix - register root first, then catch-all
		// This ensures root path "/" gets handled correctly
		r.Any("/", h)
		r.Any("/*filepath", h)
	} else {
		// Non-root prefix - register prefix catch-all and prefix itself
		pattern := cleanedPrefix + "/*filepath"
		r.AnyFunc(pattern, h.ServeHTTP)
		r.AnyFunc(cleanedPrefix, h.ServeHTTP)
	}

	return nil
}

// normalizePrefix validates and normalizes the URL prefix
func normalizePrefix(prefix string) (string, error) {
	if prefix == "" {
		return "/", nil
	}

	// Trim whitespace
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "/", nil
	}

	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	// Check for path traversal and current directory references BEFORE cleaning
	if strings.Contains(prefix, "..") || strings.Contains(prefix, "/./") || strings.HasSuffix(prefix, "/.") {
		return "", fmt.Errorf("prefix contains path traversal elements")
	}

	// Clean the path using path.Clean (not filepath.Clean for URL paths)
	cleaned := path.Clean(prefix)

	// Check again after cleaning (path.Clean might introduce ..)
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("prefix contains path traversal elements")
	}

	// Remove trailing slash except for root
	cleaned = strings.TrimSuffix(cleaned, "/")
	if cleaned == "" {
		cleaned = "/"
	}

	return cleaned, nil
}

type handler struct {
	fs           http.FileSystem
	prefix       string
	indexFile    string
	cacheControl string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestPath := r.URL.Path

	// Handle root path
	if requestPath == "/" {
		if h.prefix == "/" {
			h.serveFile(w, r, h.indexFile)
			return
		}
		// If we have a non-root prefix, root path should not be handled here
		http.NotFound(w, r)
		return
	}

	// Handle exact prefix match (e.g. /app)
	if requestPath == h.prefix {
		h.serveFile(w, r, h.indexFile)
		return
	}

	// For non-root prefix, ensure path starts with prefix
	if h.prefix != "/" && !strings.HasPrefix(requestPath, h.prefix+"/") {
		http.NotFound(w, r)
		return
	}

	// Extract relative path
	var relativePath string
	if h.prefix == "/" {
		relativePath = strings.TrimPrefix(requestPath, "/")
	} else {
		relativePath = strings.TrimPrefix(requestPath, h.prefix+"/")
	}

	// If relative path is empty after trimming, serve index
	if relativePath == "" {
		h.serveFile(w, r, h.indexFile)
		return
	}

	// Clean the path to prevent directory traversal
	cleanedPath := path.Clean(relativePath)
	if cleanedPath == "." || cleanedPath == ".." || strings.Contains(cleanedPath, "..") {
		h.serveFile(w, r, h.indexFile)
		return
	}

	// Convert to slash-separated path for consistency
	filePath := filepath.ToSlash(cleanedPath)

	// Try to serve the requested file, fall back to index
	if !h.serveFile(w, r, filePath) {
		h.serveFile(w, r, h.indexFile)
	}
}

func (h *handler) serveFile(w http.ResponseWriter, r *http.Request, filePath string) bool {
	f, err := h.fs.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return false
	}

	if stat.IsDir() {
		// Try to serve index file from directory
		indexPath := path.Join(filePath, h.indexFile)
		return h.serveFile(w, r, indexPath)
	}

	// Apply cache control if configured and not index file
	if h.cacheControl != "" && !isIndexFile(filePath, h.indexFile) {
		w.Header().Set("Cache-Control", h.cacheControl)
	}

	// Use http.ServeContent for proper content serving
	http.ServeContent(w, r, path.Base(filePath), stat.ModTime(), f)
	return true
}

// isIndexFile checks if the given path is the index file
func isIndexFile(filePath, indexFile string) bool {
	// Normalize both paths for comparison
	cleanPath := path.Clean(filePath)
	cleanIndex := path.Clean(indexFile)

	// Check if path ends with index file
	return cleanPath == cleanIndex || strings.HasSuffix(cleanPath, "/"+cleanIndex)
}
