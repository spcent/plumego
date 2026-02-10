// Package versioning provides API version negotiation middleware
//
// This package implements multiple strategies for API versioning:
//   - Accept header (content negotiation): application/vnd.myapi.v2+json
//   - URL path: /v2/users/123
//   - Query parameter: /users/123?version=2
//   - Custom header: X-API-Version: 2
//
// Example usage:
//
//	import (
//		"github.com/spcent/plumego/middleware/versioning"
//		"github.com/spcent/plumego/core"
//	)
//
//	app := core.New()
//
//	// Accept header strategy
//	app.Use(versioning.Middleware(versioning.Config{
//		Strategy:          versioning.StrategyAcceptHeader,
//		VendorPrefix:      "application/vnd.myapi",
//		DefaultVersion:    1,
//		SupportedVersions: []int{1, 2, 3},
//	}))
//
//	// URL path strategy
//	app.Use(versioning.Middleware(versioning.Config{
//		Strategy:          versioning.StrategyURLPath,
//		PathPrefix:        "/v",
//		DefaultVersion:    1,
//		SupportedVersions: []int{1, 2, 3},
//	}))
//
//	// Access version in handler
//	app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
//		version := versioning.GetVersion(r.Context())
//		// Handle request based on version
//	})
package versioning

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/spcent/plumego/contract"
)

// Strategy defines the version extraction strategy
type Strategy int

const (
	// StrategyAcceptHeader extracts version from Accept header
	// Example: Accept: application/vnd.myapi.v2+json
	StrategyAcceptHeader Strategy = iota

	// StrategyURLPath extracts version from URL path
	// Example: /v2/users/123
	StrategyURLPath

	// StrategyQueryParam extracts version from query parameter
	// Example: /users/123?version=2
	StrategyQueryParam

	// StrategyCustomHeader extracts version from custom header
	// Example: X-API-Version: 2
	StrategyCustomHeader
)

// String returns the strategy name
func (s Strategy) String() string {
	switch s {
	case StrategyAcceptHeader:
		return "accept-header"
	case StrategyURLPath:
		return "url-path"
	case StrategyQueryParam:
		return "query-param"
	case StrategyCustomHeader:
		return "custom-header"
	default:
		return "unknown"
	}
}

// Config holds version negotiation configuration
type Config struct {
	// Strategy is the version extraction strategy
	// Default: StrategyAcceptHeader
	Strategy Strategy

	// DefaultVersion is the version to use if none specified
	// Default: 1
	DefaultVersion int

	// SupportedVersions is the list of supported API versions
	// If empty, all versions are accepted
	// Default: nil (accept all)
	SupportedVersions []int

	// For StrategyAcceptHeader:
	// VendorPrefix is the vendor media type prefix
	// Example: "application/vnd.myapi" matches "application/vnd.myapi.v2+json"
	// Default: "application/vnd.api"
	VendorPrefix string

	// For StrategyURLPath:
	// PathPrefix is the version path prefix
	// Example: "/v" matches "/v2/users/123"
	// Default: "/v"
	PathPrefix string

	// For StrategyQueryParam:
	// ParamName is the query parameter name
	// Example: "version" matches "?version=2"
	// Default: "version"
	ParamName string

	// For StrategyCustomHeader:
	// HeaderName is the custom header name
	// Example: "X-API-Version"
	// Default: "X-API-Version"
	HeaderName string

	// StripVersionFromPath removes version from URL path (for StrategyURLPath)
	// Example: /v2/users/123 becomes /users/123
	// Default: true
	StripVersionFromPath bool

	// OnVersionMismatch is called when unsupported version is requested
	// If nil, returns 406 Not Acceptable
	OnVersionMismatch func(w http.ResponseWriter, r *http.Request, requested int)

	// OnVersionExtracted is called after version is extracted (optional)
	OnVersionExtracted func(r *http.Request, version int)
}

// WithDefaults returns a Config with default values applied
func (c *Config) WithDefaults() *Config {
	config := *c

	if config.DefaultVersion == 0 {
		config.DefaultVersion = 1
	}

	if config.VendorPrefix == "" {
		config.VendorPrefix = "application/vnd.api"
	}

	if config.PathPrefix == "" {
		config.PathPrefix = "/v"
	}

	if config.ParamName == "" {
		config.ParamName = "version"
	}

	if config.HeaderName == "" {
		config.HeaderName = "X-API-Version"
	}

	if !config.StripVersionFromPath {
		config.StripVersionFromPath = true
	}

	return &config
}

// contextKey is the type for context keys
type contextKey string

const (
	versionContextKey contextKey = "api_version"
)

// Middleware creates a version negotiation middleware
func Middleware(config Config) func(http.Handler) http.Handler {
	cfg := config.WithDefaults()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract version based on strategy
			version := extractVersion(r, cfg)

			// Validate version
			if len(cfg.SupportedVersions) > 0 && !isVersionSupported(version, cfg.SupportedVersions) {
				if cfg.OnVersionMismatch != nil {
					cfg.OnVersionMismatch(w, r, version)
				} else {
					contract.WriteError(w, r, contract.APIError{
						Status:   http.StatusNotAcceptable,
						Code:     "UNSUPPORTED_VERSION",
						Message:  fmt.Sprintf("Unsupported API version: %d", version),
						Category: contract.CategoryClient,
					})
				}
				return
			}

			// Strip version from path if needed
			if cfg.Strategy == StrategyURLPath && cfg.StripVersionFromPath {
				r.URL.Path = stripVersionFromPath(r.URL.Path, cfg.PathPrefix, version)
			}

			// Call hook
			if cfg.OnVersionExtracted != nil {
				cfg.OnVersionExtracted(r, version)
			}

			// Add version to context
			ctx := context.WithValue(r.Context(), versionContextKey, version)

			// Add version header to response
			w.Header().Set("X-API-Version", strconv.Itoa(version))

			// Continue with version in context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetVersion returns the API version from context
// Returns 0 if version not found in context
func GetVersion(ctx context.Context) int {
	if v, ok := ctx.Value(versionContextKey).(int); ok {
		return v
	}
	return 0
}

// extractVersion extracts version from request based on strategy
func extractVersion(r *http.Request, config *Config) int {
	var version int

	switch config.Strategy {
	case StrategyAcceptHeader:
		version = extractFromAccept(r, config.VendorPrefix)
	case StrategyURLPath:
		version = extractFromPath(r, config.PathPrefix)
	case StrategyQueryParam:
		version = extractFromQuery(r, config.ParamName)
	case StrategyCustomHeader:
		version = extractFromHeader(r, config.HeaderName)
	}

	// Use default if no version found
	if version <= 0 {
		version = config.DefaultVersion
	}

	return version
}

// extractFromAccept extracts version from Accept header
// Example: application/vnd.myapi.v2+json -> 2
func extractFromAccept(r *http.Request, vendorPrefix string) int {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return 0
	}

	// Match pattern: vendorPrefix.vN or vendorPrefix.vN+format
	// Example: application/vnd.myapi.v2+json
	pattern := regexp.QuoteMeta(vendorPrefix) + `\.v(\d+)(?:\+|;|$)`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(accept)
	if len(matches) >= 2 {
		if v, err := strconv.Atoi(matches[1]); err == nil {
			return v
		}
	}

	return 0
}

// extractFromPath extracts version from URL path
// Example: /v2/users/123 -> 2
func extractFromPath(r *http.Request, pathPrefix string) int {
	path := r.URL.Path

	// Match pattern: /vN/ or /vN (at start of path)
	if !strings.HasPrefix(path, pathPrefix) {
		return 0
	}

	// Remove prefix
	remaining := strings.TrimPrefix(path, pathPrefix)

	// Extract version number
	var versionStr string
	for i, ch := range remaining {
		if ch >= '0' && ch <= '9' {
			versionStr += string(ch)
		} else if i > 0 {
			break
		} else {
			return 0
		}
	}

	if v, err := strconv.Atoi(versionStr); err == nil {
		return v
	}

	return 0
}

// extractFromQuery extracts version from query parameter
// Example: ?version=2 -> 2
func extractFromQuery(r *http.Request, paramName string) int {
	versionStr := r.URL.Query().Get(paramName)
	if versionStr == "" {
		return 0
	}

	if v, err := strconv.Atoi(versionStr); err == nil {
		return v
	}

	return 0
}

// extractFromHeader extracts version from custom header
// Example: X-API-Version: 2 -> 2
func extractFromHeader(r *http.Request, headerName string) int {
	versionStr := r.Header.Get(headerName)
	if versionStr == "" {
		return 0
	}

	if v, err := strconv.Atoi(versionStr); err == nil {
		return v
	}

	return 0
}

// isVersionSupported checks if version is in supported list
func isVersionSupported(version int, supported []int) bool {
	for _, v := range supported {
		if version == v {
			return true
		}
	}
	return false
}

// stripVersionFromPath removes version prefix from path
// Example: /v2/users/123 -> /users/123
func stripVersionFromPath(path, pathPrefix string, version int) string {
	versionPrefix := fmt.Sprintf("%s%d", pathPrefix, version)

	if strings.HasPrefix(path, versionPrefix+"/") {
		return strings.TrimPrefix(path, versionPrefix)
	}

	if path == versionPrefix {
		return "/"
	}

	return path
}

// Extractor is a custom version extractor function
type Extractor func(r *http.Request) int

// CustomExtractor creates a middleware with custom version extraction
func CustomExtractor(extractor Extractor, defaultVersion int, supportedVersions []int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			version := extractor(r)

			if version <= 0 {
				version = defaultVersion
			}

			if len(supportedVersions) > 0 && !isVersionSupported(version, supportedVersions) {
				contract.WriteError(w, r, contract.APIError{
					Status:   http.StatusNotAcceptable,
					Code:     "UNSUPPORTED_VERSION",
					Message:  fmt.Sprintf("Unsupported API version: %d", version),
					Category: contract.CategoryClient,
				})
				return
			}

			ctx := context.WithValue(r.Context(), versionContextKey, version)
			w.Header().Set("X-API-Version", strconv.Itoa(version))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
