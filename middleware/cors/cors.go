package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/middleware"
)

// CORSOptions configures Cross-Origin Resource Sharing (CORS) behavior.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           time.Duration
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func joinOrDefault(slice []string, def string) string {
	if len(slice) == 0 {
		return def
	}
	return strings.Join(slice, ", ")
}

// Middleware provides configurable CORS support.
func Middleware(opts CORSOptions) middleware.Middleware {
	if len(opts.AllowedOrigins) == 0 {
		opts.AllowedOrigins = []string{"*"}
	}
	if len(opts.AllowedMethods) == 0 {
		opts.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(opts.AllowedHeaders) == 0 {
		opts.AllowedHeaders = []string{"Accept", "Accept-Language", "Content-Language", "Content-Type", "Authorization"}
	}

	allowAllOrigins := contains(opts.AllowedOrigins, "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			var allowOriginValue string
			if allowAllOrigins {
				if opts.AllowCredentials {
					allowOriginValue = origin
				} else {
					allowOriginValue = "*"
				}
			} else if contains(opts.AllowedOrigins, origin) {
				allowOriginValue = origin
			} else {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Origin", allowOriginValue)
			if opts.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if len(opts.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", joinOrDefault(opts.ExposeHeaders, ""))
			}

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", joinOrDefault(opts.AllowedMethods, "GET, POST, OPTIONS"))

				reqHeaders := r.Header.Get("Access-Control-Request-Headers")
				if reqHeaders != "" {
					if contains(opts.AllowedHeaders, "*") {
						w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
					} else {
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
					w.Header().Set("Access-Control-Allow-Headers", joinOrDefault(opts.AllowedHeaders, ""))
				}

				if opts.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.FormatInt(int64(opts.MaxAge/time.Second), 10))
				}

				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
