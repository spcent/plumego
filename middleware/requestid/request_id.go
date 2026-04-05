package requestid

import (
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
)

type config struct {
	headerName       string
	fallbackHeader   string
	generate         func() string
	includeInRequest bool
}

type Option func(*config)

func WithHeader(name string) Option {
	return func(cfg *config) {
		if name != "" {
			cfg.headerName = name
		}
	}
}

func WithFallbackHeader(name string) Option {
	return func(cfg *config) {
		if name != "" {
			cfg.fallbackHeader = name
		}
	}
}

func WithGenerator(fn func() string) Option {
	return func(cfg *config) {
		if fn != nil {
			cfg.generate = fn
		}
	}
}

func WithRequestHeader(enabled bool) Option {
	return func(cfg *config) {
		cfg.includeInRequest = enabled
	}
}

func Middleware(opts ...Option) middleware.Middleware {
	cfg := config{
		headerName:       "X-Request-ID",
		fallbackHeader:   "X-Trace-ID",
		generate:         log.NewTraceID,
		includeInRequest: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := traceIDFromRequest(r, cfg.headerName, cfg.fallbackHeader)
			if id == "" {
				id = cfg.generate()
			}

			r = contract.DefaultObservabilityPolicy.AttachRequestID(w, r, id, cfg.includeInRequest)
			r = r.WithContext(log.WithTraceID(r.Context(), id))
			if cfg.headerName != "" && cfg.headerName != contract.RequestIDHeader {
				if cfg.includeInRequest {
					r.Header.Set(cfg.headerName, id)
				}
				w.Header().Set(cfg.headerName, id)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func traceIDFromRequest(r *http.Request, headerName, fallback string) string {
	if id := contract.DefaultObservabilityPolicy.RequestIDFromRequest(r); id != "" {
		return id
	}
	if r == nil {
		return ""
	}
	if headerName != "" && headerName != contract.RequestIDHeader {
		if id := strings.TrimSpace(r.Header.Get(headerName)); id != "" {
			return id
		}
	}
	if fallback != "" && fallback != contract.FallbackRequestIDHeader {
		if id := strings.TrimSpace(r.Header.Get(fallback)); id != "" {
			return id
		}
	}
	return ""
}
