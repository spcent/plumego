package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CORSOptions configures Cross-Origin Resource Sharing (CORS) behavior.
//
// CORS allows web applications to control which origins can access their resources.
// This is essential for web APIs that are consumed by frontend applications hosted
// on different domains.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/cors"
//
//	opts := cors.CORSOptions{
//		AllowedOrigins:   []string{"https://example.com", "https://app.example.com"},
//		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
//		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID"},
//		AllowCredentials: true,
//		MaxAge:           24 * time.Hour, // Cache preflight responses for 24 hours
//	}
//	handler := cors.CORSWithOptions(opts, myHandlerFunc)
type CORSOptions struct {
	// AllowedOrigins specifies which origins are allowed to access the resource.
	// Use []string{"*"} to allow all origins (not recommended for production).
	// Example: []string{"https://example.com", "https://app.example.com"}
	AllowedOrigins []string

	// AllowedMethods specifies which HTTP methods are allowed.
	// Example: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	AllowedMethods []string

	// AllowedHeaders specifies which HTTP headers are allowed in requests.
	// Example: []string{"Content-Type", "Authorization", "X-Request-ID"}
	AllowedHeaders []string

	// AllowCredentials indicates whether credentials (cookies, HTTP authentication)
	// can be included in cross-origin requests.
	AllowCredentials bool

	// ExposeHeaders specifies which response headers are exposed to the client.
	// Example: []string{"X-Request-ID", "X-RateLimit-Remaining"}
	ExposeHeaders []string

	// MaxAge specifies how long the browser should cache preflight responses.
	// Example: 24 * time.Hour
	MaxAge time.Duration
}

// contains helper (case-sensitive for origins)
func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// joinOrDefault joins slice by comma or returns default if empty
func joinOrDefault(slice []string, def string) string {
	if len(slice) == 0 {
		return def
	}
	return strings.Join(slice, ", ")
}

// CORS provides basic CORS support with wildcard origin.
//
// WARNING: This middleware allows ALL origins ("*") and is intended for
// development and testing only. Do NOT use in production — it exposes your
// API to cross-origin requests from any website.
// For production, use [CORSWithOptions] with explicit allowed origins.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/cors"
//
//	handler := cors.CORS(myHandler)
//
// The middleware sets the following headers:
//   - Access-Control-Allow-Origin: *
//   - Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
//   - Access-Control-Allow-Headers: Content-Type, Authorization
//
// For OPTIONS requests, it returns a 204 No Content response.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CORSWithOptions provides configurable CORS support.
//
// This middleware allows you to specify which origins, methods, and headers are allowed.
// It properly handles preflight OPTIONS requests and respects credentials.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/cors"
//
//	opts := cors.CORSOptions{
//		AllowedOrigins:   []string{"https://example.com"},
//		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
//		AllowedHeaders:   []string{"Content-Type", "Authorization"},
//		AllowCredentials: true,
//		MaxAge:           24 * time.Hour,
//	}
//	handler := cors.CORSWithOptions(opts, myHandlerFunc)
//
// The middleware handles both simple requests and preflight requests:
//   - Simple requests: Adds CORS headers and forwards to next handler
//   - Preflight requests (OPTIONS): Returns 204 No Content with CORS headers
//
// Security considerations:
//   - When AllowCredentials is true, Access-Control-Allow-Origin cannot be "*"
//     and will echo the request's Origin header instead
//   - Use specific origins instead of "*" in production
func CORSWithOptions(opts CORSOptions, next http.Handler) http.Handler {
	// provide sensible defaults
	if len(opts.AllowedMethods) == 0 {
		opts.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(opts.AllowedHeaders) == 0 {
		opts.AllowedHeaders = []string{"Accept", "Accept-Language", "Content-Language", "Content-Type", "Authorization"}
	}

	allowAllOrigins := contains(opts.AllowedOrigins, "*")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// If no Origin header -> not a cross-origin request; just pass through
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Determine allowed origin value to put into header.
		var allowOriginValue string
		if allowAllOrigins {
			// If credentials allowed, you must echo the request origin instead of "*"
			if opts.AllowCredentials {
				allowOriginValue = origin
			} else {
				allowOriginValue = "*"
			}
		} else if contains(opts.AllowedOrigins, origin) {
			allowOriginValue = origin
		} else {
			// origin not allowed -> no CORS headers, proceed normally (browser will block)
			next.ServeHTTP(w, r)
			return
		}

		// Always vary on Origin (caching proxies)
		w.Header().Add("Vary", "Origin")

		// Common CORS headers for non-preflight responses
		w.Header().Set("Access-Control-Allow-Origin", allowOriginValue)
		if opts.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if len(opts.ExposeHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", joinOrDefault(opts.ExposeHeaders, ""))
		}

		// Handle preflight request
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			// Access-Control-Allow-Methods
			w.Header().Set("Access-Control-Allow-Methods", joinOrDefault(opts.AllowedMethods, "GET, POST, OPTIONS"))

			// Access-Control-Allow-Headers:
			// If client requested specific headers, echo them back if allowed; otherwise use configured allowed headers
			reqHeaders := r.Header.Get("Access-Control-Request-Headers")
			if reqHeaders != "" {
				// If AllowedHeaders contains "*", just echo requested headers.
				if contains(opts.AllowedHeaders, "*") {
					w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
				} else {
					// validate each requested header is allowed (case-insensitive)
					reqs := strings.Split(reqHeaders, ",")
					allowed := []string{}
					for _, h := range reqs {
						h = strings.TrimSpace(h)
						hi := strings.ToLower(h)
						for _, ah := range opts.AllowedHeaders {
							if strings.ToLower(ah) == hi {
								allowed = append(allowed, h)
								break
							}
						}
					}
					if len(allowed) > 0 {
						w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowed, ", "))
					}
				}
			} else {
				// no requested headers -> set configured allowed headers
				w.Header().Set("Access-Control-Allow-Headers", joinOrDefault(opts.AllowedHeaders, ""))
			}

			// Max-Age
			if opts.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.FormatInt(int64(opts.MaxAge/time.Second), 10))
			}

			// Successful preflight - no body
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Not preflight -> call next handler normally
		next.ServeHTTP(w, r)
	})
}

// CORSWithOptionsFunc is a convenience wrapper around CORSWithOptions
// for http.HandlerFunc callers.
func CORSWithOptionsFunc(opts CORSOptions, next http.HandlerFunc) http.HandlerFunc {
	return CORSWithOptions(opts, next).ServeHTTP
}
