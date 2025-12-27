package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CORSOptions struct {
	AllowedOrigins   []string // e.g. []string{"https://example.com", "https://foo.com"} or []string{"*"}
	AllowedMethods   []string // e.g. []string{"GET","POST","PUT","DELETE","OPTIONS"}
	AllowedHeaders   []string // e.g. []string{"Content-Type","Authorization"}
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           time.Duration // e.g. time.Hour * 24
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

func CORSWithOptions(opts CORSOptions, next http.HandlerFunc) http.HandlerFunc {
	// provide sensible defaults
	if len(opts.AllowedMethods) == 0 {
		opts.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(opts.AllowedHeaders) == 0 {
		opts.AllowedHeaders = []string{"Accept", "Accept-Language", "Content-Language", "Content-Type", "Authorization"}
	}

	allowAllOrigins := contains(opts.AllowedOrigins, "*")

	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// If no Origin header -> not a cross-origin request; just pass through
		if origin == "" {
			next(w, r)
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
			next(w, r)
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
		next(w, r)
	}
}
