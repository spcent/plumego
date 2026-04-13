package frontend

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

const (
	defaultIndex = "index.html"
	methodAny    = "ANY"
)

// config defines how a built frontend bundle is served.
type config struct {
	Prefix              string
	IndexFile           string
	CacheControl        string
	IndexCacheControl   string
	Fallback            bool
	Headers             map[string]string
	EnablePrecompressed bool
	NotFoundPage        string
	ErrorPage           string
	MIMETypes           map[string]string
}

// Option mutates a config.
type Option func(*config)

// Mount describes an explicitly constructed frontend mount.
// It separates bundle construction from router registration.
type Mount struct {
	prefix  string
	handler http.Handler
}

type routeRegistrar interface {
	AddRoute(method, path string, handler http.Handler, opts ...router.RouteOption) error
}

// WithPrefix sets the mount prefix for the frontend bundle.
func WithPrefix(prefix string) Option {
	return func(cfg *config) {
		cfg.Prefix = prefix
	}
}

// WithIndex overrides the default index.html fallback filename.
func WithIndex(name string) Option {
	return func(cfg *config) {
		cfg.IndexFile = name
	}
}

// WithCacheControl sets a Cache-Control header for served assets.
func WithCacheControl(header string) Option {
	return func(cfg *config) {
		cfg.CacheControl = header
	}
}

// WithIndexCacheControl sets the Cache-Control header for index responses.
func WithIndexCacheControl(header string) Option {
	return func(cfg *config) {
		cfg.IndexCacheControl = header
	}
}

// WithFallback enables or disables SPA fallback to the index file.
func WithFallback(enabled bool) Option {
	return func(cfg *config) {
		cfg.Fallback = enabled
	}
}

// WithHeaders applies additional headers to every successful file response.
func WithHeaders(headers map[string]string) Option {
	return func(cfg *config) {
		cfg.Headers = copyHeaders(headers)
	}
}

// WithPrecompressed enables serving pre-compressed files (.gz, .br).
// When enabled, if a requested file has a .gz or .br variant and the client
// supports that encoding, the pre-compressed file is served directly.
func WithPrecompressed(enabled bool) Option {
	return func(cfg *config) {
		cfg.EnablePrecompressed = enabled
	}
}

// WithNotFoundPage sets a custom 404 error page.
// The page path is relative to the filesystem root.
func WithNotFoundPage(page string) Option {
	return func(cfg *config) {
		cfg.NotFoundPage = strings.TrimSpace(page)
	}
}

// WithErrorPage sets a custom 5xx error page.
// The page path is relative to the filesystem root.
func WithErrorPage(page string) Option {
	return func(cfg *config) {
		cfg.ErrorPage = strings.TrimSpace(page)
	}
}

// WithMIMETypes sets custom MIME type mappings for file extensions.
// Extensions may omit the leading dot; it is added automatically.
func WithMIMETypes(mimeTypes map[string]string) Option {
	return func(cfg *config) {
		cfg.MIMETypes = mimeTypes
	}
}

// RegisterFromDir mounts a built frontend directory (e.g. Next.js `out/`)
// at the given prefix. Returns an error if the directory is missing or unreadable.
func RegisterFromDir(r routeRegistrar, dir string, opts ...Option) error {
	mount, err := NewMountFromDir(dir, opts...)
	if err != nil {
		return err
	}
	return mount.Register(r)
}

// NewMountFromDir constructs a frontend mount from a filesystem directory.
func NewMountFromDir(dir string, opts ...Option) (*Mount, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("frontend directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("frontend path %q is not a directory", dir)
	}
	if _, err := os.ReadDir(dir); err != nil {
		return nil, fmt.Errorf("frontend directory %q not readable: %w", dir, err)
	}
	return NewMountFS(http.Dir(dir), opts...)
}

// RegisterFS mounts a frontend bundle served from the provided http.FileSystem.
// This is suitable for go:embed bundles using http.FS.
func RegisterFS(r routeRegistrar, fsys http.FileSystem, opts ...Option) error {
	mount, err := NewMountFS(fsys, opts...)
	if err != nil {
		return err
	}
	return mount.Register(r)
}

// NewMountFS constructs a frontend mount from an http.FileSystem.
func NewMountFS(fsys http.FileSystem, opts ...Option) (*Mount, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}
	h, err := NewHandlerFS(fsys, opts...)
	if err != nil {
		return nil, err
	}
	return &Mount{
		prefix:  cfg.Prefix,
		handler: h,
	}, nil
}

// NewHandlerFS constructs a frontend handler without registering routes.
func NewHandlerFS(fsys http.FileSystem, opts ...Option) (http.Handler, error) {
	if fsys == nil {
		return nil, errors.New("filesystem cannot be nil")
	}
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}
	return &handler{cfg: *cfg, fs: fsys}, nil
}

// Prefix returns the normalized mount prefix.
func (m *Mount) Prefix() string {
	if m == nil {
		return ""
	}
	return m.prefix
}

// Handler returns the mounted frontend handler.
func (m *Mount) Handler() http.Handler {
	if m == nil {
		return nil
	}
	return m.handler
}

// Register attaches the mount to the provided router.
func (m *Mount) Register(r routeRegistrar) error {
	if m == nil {
		return errors.New("mount cannot be nil")
	}
	if r == nil {
		return errors.New("router cannot be nil")
	}
	if m.handler == nil {
		return errors.New("mount handler cannot be nil")
	}

	if m.prefix == "/" {
		if err := r.AddRoute(methodAny, "/", m.handler); err != nil {
			return err
		}
		return r.AddRoute(methodAny, "/*filepath", m.handler)
	}

	if err := r.AddRoute(methodAny, m.prefix+"/*filepath", m.handler); err != nil {
		return err
	}
	return r.AddRoute(methodAny, m.prefix, m.handler)
}

func newConfig(opts ...Option) (*config, error) {
	cfg := &config{
		Prefix:    "/",
		IndexFile: defaultIndex,
		Fallback:  true,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	cleanedPrefix, err := normalizePrefix(cfg.Prefix)
	if err != nil {
		return nil, fmt.Errorf("invalid prefix %q: %w", cfg.Prefix, err)
	}
	cfg.Prefix = cleanedPrefix

	indexFile := strings.TrimSpace(cfg.IndexFile)
	if indexFile == "" {
		indexFile = defaultIndex
	}
	if strings.Contains(indexFile, "/") || strings.Contains(indexFile, "\\") {
		return nil, fmt.Errorf("index file %q cannot contain path separators", cfg.IndexFile)
	}
	cfg.IndexFile = indexFile
	cfg.MIMETypes = normalizeMIMETypes(cfg.MIMETypes)
	return cfg, nil
}

// normalizePrefix validates and normalizes the URL prefix.
func normalizePrefix(prefix string) (string, error) {
	if prefix == "" {
		return "/", nil
	}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "/", nil
	}

	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	if strings.Contains(prefix, "..") || strings.Contains(prefix, "/./") || strings.HasSuffix(prefix, "/.") {
		return "", fmt.Errorf("prefix contains path traversal elements")
	}

	cleaned := path.Clean(prefix)
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("prefix contains path traversal elements")
	}

	cleaned = strings.TrimSuffix(cleaned, "/")
	if cleaned == "" {
		cleaned = "/"
	}

	return cleaned, nil
}

// normalizeMIMETypes normalizes MIME type extension keys: trims whitespace, ensures
// a leading dot, and filters entries with empty extensions or values.
func normalizeMIMETypes(raw map[string]string) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	normalized := make(map[string]string, len(raw))
	for ext, mime := range raw {
		ext = strings.TrimSpace(ext)
		mime = strings.TrimSpace(mime)
		if ext == "" || mime == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		normalized[ext] = mime
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

type handler struct {
	cfg config
	fs  http.FileSystem
}

// statusCodeWriter wraps http.ResponseWriter to enforce a specific HTTP status
// code when serving custom error or not-found pages. http.ServeContent always
// writes 200 (or 206/304), so this wrapper intercepts WriteHeader and replaces
// any 2xx code with the desired error status while passing through 3xx and
// other codes unchanged.
type statusCodeWriter struct {
	http.ResponseWriter
	code        int
	wroteHeader bool
}

func (s *statusCodeWriter) WriteHeader(code int) {
	if !s.wroteHeader {
		s.wroteHeader = true
		if code >= 200 && code < 300 {
			code = s.code
		}
		s.ResponseWriter.WriteHeader(code)
	}
}

func (s *statusCodeWriter) Write(b []byte) (int, error) {
	if !s.wroteHeader {
		s.WriteHeader(s.code)
	}
	return s.ResponseWriter.Write(b)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusMethodNotAllowed).
			Code(contract.CodeMethodNotAllowed).
			Message("method not allowed").
			Category(contract.CategoryClient).
			Build())
		return
	}

	rel, ok := h.stripPrefix(r.URL.Path)
	if !ok {
		h.serveNotFound(w, r)
		return
	}

	if rel == "" {
		h.serveIndex(w, r)
		return
	}

	cleaned := path.Clean(rel)
	if cleaned == "." || cleaned == ".." || strings.Contains(cleaned, "..") {
		if h.cfg.Fallback {
			h.serveIndex(w, r)
		} else {
			h.serveNotFound(w, r)
		}
		return
	}

	filePath := filepath.ToSlash(cleaned)
	served, err := h.serveFile(w, r, filePath)
	if err != nil {
		h.serveError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}
	if !served {
		if h.cfg.Fallback {
			h.serveIndex(w, r)
		} else {
			h.serveNotFound(w, r)
		}
	}
}

// stripPrefix extracts the relative path from requestPath by stripping h.cfg.Prefix.
// Returns ("", false) if the path does not belong to this handler's prefix.
// Returns (rel, true) on success; rel may be "" meaning: serve the index.
func (h *handler) stripPrefix(requestPath string) (string, bool) {
	if h.cfg.Prefix == "/" {
		if requestPath == "/" {
			return "", true
		}
		return strings.TrimPrefix(requestPath, "/"), true
	}
	if requestPath == h.cfg.Prefix {
		return "", true
	}
	if !strings.HasPrefix(requestPath, h.cfg.Prefix+"/") {
		return "", false
	}
	return strings.TrimPrefix(requestPath, h.cfg.Prefix+"/"), true
}

// serveIndex serves the configured index file, falling back to 404 if not found.
func (h *handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	served, err := h.serveFile(w, r, h.cfg.IndexFile)
	if err != nil {
		h.serveError(w, r, "internal server error", http.StatusInternalServerError)
	} else if !served {
		h.serveNotFound(w, r)
	}
}

// serveFile attempts to serve the specified file. Return values:
//   - (true, nil):  file was served, response written
//   - (false, nil): file not found, no response written; caller should fall back or serve 404
//   - (false, err): server-side IO error, no response written; caller should call serveError
func (h *handler) serveFile(w http.ResponseWriter, r *http.Request, filePath string) (bool, error) {
	if preFile, preStat, encoding := h.tryPrecompressed(r, filePath); preFile != nil {
		defer preFile.Close()
		h.applyFileHeaders(w, filePath)
		w.Header().Set("Content-Encoding", encoding)
		w.Header().Add("Vary", "Accept-Encoding")
		http.ServeContent(w, r, path.Base(filePath), preStat.ModTime(), preFile)
		return true, nil
	}

	f, err := h.fs.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("open %q: %w", filePath, err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return false, fmt.Errorf("stat %q: %w", filePath, err)
	}

	if stat.IsDir() {
		return h.serveFile(w, r, path.Join(filePath, h.cfg.IndexFile))
	}

	h.applyFileHeaders(w, filePath)
	http.ServeContent(w, r, path.Base(filePath), stat.ModTime(), f)
	return true, nil
}

// applyFileHeaders sets response headers common to all file responses:
// custom security/app headers, cache-control, and custom MIME type.
func (h *handler) applyFileHeaders(w http.ResponseWriter, filePath string) {
	for key, value := range h.cfg.Headers {
		if key != "" {
			w.Header().Set(key, value)
		}
	}
	if isIndexFile(filePath, h.cfg.IndexFile) {
		if h.cfg.IndexCacheControl != "" {
			w.Header().Set("Cache-Control", h.cfg.IndexCacheControl)
		}
	} else if h.cfg.CacheControl != "" {
		w.Header().Set("Cache-Control", h.cfg.CacheControl)
	}
	if ext := path.Ext(filePath); ext != "" {
		if customType := h.cfg.MIMETypes[ext]; customType != "" {
			w.Header().Set("Content-Type", customType)
		}
	}
}

// isIndexFile checks if the given path is the index file.
func isIndexFile(filePath, indexFile string) bool {
	cleanPath := path.Clean(filePath)
	cleanIndex := path.Clean(indexFile)
	return cleanPath == cleanIndex || strings.HasSuffix(cleanPath, "/"+cleanIndex)
}

func copyHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	copied := make(map[string]string, len(headers))
	for key, value := range headers {
		cleanKey := http.CanonicalHeaderKey(strings.TrimSpace(key))
		if cleanKey == "" {
			continue
		}
		copied[cleanKey] = strings.TrimSpace(value)
	}
	if len(copied) == 0 {
		return nil
	}
	return copied
}

// serveNotFound serves a custom 404 page or falls back to http.NotFound.
// When a custom page is configured it is served with a 404 status code
// (not 200) via statusCodeWriter.
func (h *handler) serveNotFound(w http.ResponseWriter, r *http.Request) {
	if h.cfg.NotFoundPage != "" {
		sw := &statusCodeWriter{ResponseWriter: w, code: http.StatusNotFound}
		if served, _ := h.serveFile(sw, r, h.cfg.NotFoundPage); served {
			return
		}
	}
	http.NotFound(w, r)
}

// serveError serves a custom 5xx error page or falls back to a JSON error response.
// When a custom page is configured it is served with the given error status code
// (not 200) via statusCodeWriter. Errors from loading the error page itself are
// ignored to avoid recursion; the JSON fallback is used instead.
func (h *handler) serveError(w http.ResponseWriter, r *http.Request, message string, code int) {
	if h.cfg.ErrorPage != "" && code >= 500 {
		sw := &statusCodeWriter{ResponseWriter: w, code: code}
		if served, _ := h.serveFile(sw, r, h.cfg.ErrorPage); served {
			return
		}
	}
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Status(code).
		Code(contract.CodeInternalError).
		Message(message).
		Category(contract.CategoryServer).
		Build())
}

// acceptsToken reports whether the client's Accept-Encoding header contains the
// named encoding token. It compares tokens case-insensitively and ignores quality
// factor parameters (e.g. ";q=0.5").
func acceptsToken(r *http.Request, token string) bool {
	header := r.Header.Get("Accept-Encoding")
	if header == "" {
		return false
	}
	for _, part := range strings.Split(header, ",") {
		t, _, _ := strings.Cut(strings.TrimSpace(part), ";")
		if strings.EqualFold(strings.TrimSpace(t), token) {
			return true
		}
	}
	return false
}

// tryPrecompressed attempts to serve a pre-compressed version of the file.
// Returns (file, stat, encoding) if successful, or (nil, nil, "") if not found.
func (h *handler) tryPrecompressed(r *http.Request, filePath string) (http.File, os.FileInfo, string) {
	if !h.cfg.EnablePrecompressed {
		return nil, nil, ""
	}

	// Try Brotli first (better compression)
	if acceptsToken(r, "br") {
		if f, stat := h.tryOpenFile(filePath + ".br"); f != nil {
			return f, stat, "br"
		}
	}

	// Try Gzip
	if acceptsToken(r, "gzip") {
		if f, stat := h.tryOpenFile(filePath + ".gz"); f != nil {
			return f, stat, "gzip"
		}
	}

	return nil, nil, ""
}

// tryOpenFile attempts to open a file and return it with its stat.
func (h *handler) tryOpenFile(filePath string) (http.File, os.FileInfo) {
	f, err := h.fs.Open(filePath)
	if err != nil {
		return nil, nil
	}
	stat, err := f.Stat()
	if err != nil || stat.IsDir() {
		f.Close()
		return nil, nil
	}
	return f, stat
}
