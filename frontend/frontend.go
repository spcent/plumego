package frontend

import (
	"errors"
	"net/http"
	"os"
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
		return err
	}
	if !info.IsDir() {
		return errors.New("frontend path is not a directory")
	}

	return RegisterFS(r, http.Dir(dir), opts...)
}

// RegisterFS mounts a frontend bundle served from the provided http.FileSystem.
// This is suitable for go:embed bundles using http.FS.
func RegisterFS(r *router.Router, fsys http.FileSystem, opts ...Option) error {
	cfg := &Config{Prefix: "/", IndexFile: defaultIndex}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.Prefix == "" {
		cfg.Prefix = "/"
	}

	cleaned := cfg.Prefix
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}
	cleaned = strings.TrimRight(cleaned, "/")
	if cleaned == "" {
		cleaned = "/"
	}

	h := &handler{
		fs:           fsys,
		prefix:       cleaned,
		indexFile:    cfg.IndexFile,
		cacheControl: cfg.CacheControl,
	}

	pattern := cleaned + "/*filepath"
	if cleaned == "/" {
		pattern = "/*filepath"
	}

	r.AnyFunc(pattern, h.ServeHTTP)
	if cleaned == "/" {
		r.AnyFunc("/", h.ServeHTTP)
	}
	if cleaned != "/" {
		// Ensure the bare prefix also routes to the handler (e.g. /app → /app/index.html).
		r.AnyFunc(cleaned, h.ServeHTTP)
	}

	return nil
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

	path := strings.TrimPrefix(r.URL.Path, h.prefix)
	if path == "" || path == "/" {
		path = h.indexFile
	} else {
		path = strings.TrimPrefix(path, "/")
		path = filepath.Clean(path)
		if path == "." {
			path = h.indexFile
		}
	}

	if !h.serveIfExists(w, r, path) {
		h.serveIfExists(w, r, h.indexFile)
	}
}

func (h *handler) serveIfExists(w http.ResponseWriter, r *http.Request, path string) bool {
	f, err := h.fs.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return false
	}

	if stat.IsDir() {
		indexPath := filepath.Join(path, h.indexFile)
		return h.serveIfExists(w, r, indexPath)
	}

	if h.cacheControl != "" && filepath.Base(path) != h.indexFile {
		w.Header().Set("Cache-Control", h.cacheControl)
	}

	http.ServeContent(w, r, path, stat.ModTime(), f)
	return true
}
