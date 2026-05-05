package frontend

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"unicode"
)

const (
	defaultIndex = "index.html"
	methodAny    = "ANY"
	allowMethods = "GET, HEAD"
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
		cfg.Headers = cloneHeaders(headers)
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
	if cfg.NotFoundPage != "" {
		page, ok := cleanAssetPath(cfg.NotFoundPage)
		if !ok {
			return nil, fmt.Errorf("not found page %q must be a relative asset path", cfg.NotFoundPage)
		}
		cfg.NotFoundPage = page
	}
	if cfg.ErrorPage != "" {
		page, ok := cleanAssetPath(cfg.ErrorPage)
		if !ok {
			return nil, fmt.Errorf("error page %q must be a relative asset path", cfg.ErrorPage)
		}
		cfg.ErrorPage = page
	}
	cacheControl, err := normalizeHeaderOptionValue("cache control", cfg.CacheControl)
	if err != nil {
		return nil, err
	}
	cfg.CacheControl = cacheControl
	indexCacheControl, err := normalizeHeaderOptionValue("index cache control", cfg.IndexCacheControl)
	if err != nil {
		return nil, err
	}
	cfg.IndexCacheControl = indexCacheControl
	headers, err := normalizeHeaders(cfg.Headers)
	if err != nil {
		return nil, err
	}
	cfg.Headers = headers
	mimeTypes, err := normalizeMIMETypes(cfg.MIMETypes)
	if err != nil {
		return nil, err
	}
	cfg.MIMETypes = mimeTypes
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

	if strings.Contains(prefix, "\x00") || strings.Contains(prefix, "\\") {
		return "", fmt.Errorf("prefix contains path traversal elements")
	}
	if hasUnsafePathSegment(prefix) {
		return "", fmt.Errorf("prefix contains path traversal elements")
	}

	cleaned := path.Clean(prefix)
	if hasUnsafePathSegment(cleaned) {
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
func normalizeMIMETypes(raw map[string]string) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
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
		if !isSafeHeaderValue(mime) {
			return nil, fmt.Errorf("frontend MIME type %q contains invalid value", ext)
		}
		ext = strings.ToLower(ext)
		normalized[ext] = mime
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, nil
}

func normalizeHeaderOptionValue(name, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !isSafeHeaderValue(value) {
		return "", fmt.Errorf("frontend %s contains invalid value", name)
	}
	return value, nil
}

func cloneHeaders(headers map[string]string) map[string]string {
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

func normalizeHeaders(headers map[string]string) (map[string]string, error) {
	if len(headers) == 0 {
		return nil, nil
	}
	normalized := make(map[string]string, len(headers))
	for key, value := range headers {
		cleanKey := http.CanonicalHeaderKey(strings.TrimSpace(key))
		if cleanKey == "" {
			continue
		}
		if isDisallowedCustomHeader(cleanKey) {
			return nil, fmt.Errorf("frontend header %q cannot be set with WithHeaders", cleanKey)
		}
		cleanValue := strings.TrimSpace(value)
		if !isSafeHeaderValue(cleanValue) {
			return nil, fmt.Errorf("frontend header %q contains invalid value", cleanKey)
		}
		normalized[cleanKey] = cleanValue
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, nil
}

func isDisallowedCustomHeader(key string) bool {
	switch http.CanonicalHeaderKey(key) {
	case "Accept-Ranges",
		"Cache-Control",
		"Connection",
		"Content-Encoding",
		"Content-Length",
		"Content-Range",
		"Content-Type",
		"Etag",
		"Keep-Alive",
		"Last-Modified",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
		"Vary":
		return true
	default:
		return false
	}
}

func isSafeHeaderValue(value string) bool {
	for _, r := range value {
		if r == '\t' {
			continue
		}
		if r == '\r' || r == '\n' || r == 0 || unicode.IsControl(r) {
			return false
		}
	}
	return true
}
