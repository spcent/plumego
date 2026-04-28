package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/middleware"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
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

func (o CORSOptions) withDefaults() CORSOptions {
	if len(o.AllowedOrigins) == 0 {
		o.AllowedOrigins = []string{"*"}
	}
	if len(o.AllowedMethods) == 0 {
		o.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(o.AllowedHeaders) == 0 {
		o.AllowedHeaders = []string{"Accept", "Accept-Language", "Content-Language", "Content-Type", "Authorization"}
	}
	return o
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

func containsFold(slice []string, s string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}

// Middleware provides configurable CORS support.
func Middleware(opts CORSOptions) middleware.Middleware {
	opts = opts.withDefaults()
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

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				requestMethod := strings.TrimSpace(r.Header.Get("Access-Control-Request-Method"))
				if !containsFold(opts.AllowedMethods, requestMethod) {
					next.ServeHTTP(w, r)
					return
				}

				var allowHeadersValue string
				reqHeaders := r.Header.Get("Access-Control-Request-Headers")
				if reqHeaders != "" {
					if contains(opts.AllowedHeaders, "*") {
						allowHeadersValue = reqHeaders
					} else {
						allowed, ok := allowedRequestedHeaders(reqHeaders, opts.AllowedHeaders)
						if !ok {
							next.ServeHTTP(w, r)
							return
						}
						allowHeadersValue = strings.Join(allowed, ", ")
					}
				} else {
					allowHeadersValue = joinOrDefault(opts.AllowedHeaders, "")
				}

				internaltransport.AddVary(w.Header(), "Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers")
				w.Header().Set("Access-Control-Allow-Origin", allowOriginValue)
				if opts.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Set("Access-Control-Allow-Methods", joinOrDefault(opts.AllowedMethods, "GET, POST, OPTIONS"))
				if allowHeadersValue != "" {
					w.Header().Set("Access-Control-Allow-Headers", allowHeadersValue)
				}

				if opts.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.FormatInt(int64(opts.MaxAge/time.Second), 10))
				}

				w.WriteHeader(http.StatusNoContent)
				return
			}

			internaltransport.AddVary(w.Header(), "Origin")
			w.Header().Set("Access-Control-Allow-Origin", allowOriginValue)
			if opts.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if len(opts.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", joinOrDefault(opts.ExposeHeaders, ""))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func allowedRequestedHeaders(requestHeaders string, allowedHeaders []string) ([]string, bool) {
	reqs := strings.Split(requestHeaders, ",")
	allowed := make([]string, 0, len(reqs))
	for _, h := range reqs {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if !containsFold(allowedHeaders, h) {
			return nil, false
		}
		allowed = append(allowed, h)
	}
	if len(allowed) == 0 {
		return nil, false
	}
	return allowed, true
}
