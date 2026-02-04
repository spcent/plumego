package observability

import (
	"context"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
)

type requestIDConfig struct {
	headerName       string
	fallbackHeader   string
	generate         func() string
	includeInRequest bool
}

// RequestIDOption configures the RequestID middleware.
type RequestIDOption func(*requestIDConfig)

// WithRequestIDHeader sets the primary header name for request ids.
func WithRequestIDHeader(name string) RequestIDOption {
	return func(cfg *requestIDConfig) {
		if name != "" {
			cfg.headerName = name
		}
	}
}

// WithRequestIDFallbackHeader sets a secondary header to read when the primary is missing.
func WithRequestIDFallbackHeader(name string) RequestIDOption {
	return func(cfg *requestIDConfig) {
		if name != "" {
			cfg.fallbackHeader = name
		}
	}
}

// WithRequestIDGenerator sets the request id generator.
func WithRequestIDGenerator(fn func() string) RequestIDOption {
	return func(cfg *requestIDConfig) {
		if fn != nil {
			cfg.generate = fn
		}
	}
}

// WithRequestIDRequestHeader controls whether to write the id into the request header.
func WithRequestIDRequestHeader(enabled bool) RequestIDOption {
	return func(cfg *requestIDConfig) {
		cfg.includeInRequest = enabled
	}
}

// RequestID ensures each request has a trace/request id in context and response headers.
func RequestID(opts ...RequestIDOption) middleware.Middleware {
	cfg := requestIDConfig{
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

			ctx := context.WithValue(r.Context(), contract.TraceIDKey{}, id)
			if cfg.includeInRequest {
				r.Header.Set(cfg.headerName, id)
			}
			w.Header().Set(cfg.headerName, id)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func traceIDFromRequest(r *http.Request, headerName, fallback string) string {
	if r == nil {
		return ""
	}
	if id := contract.TraceIDFromContext(r.Context()); id != "" {
		return id
	}
	if headerName != "" {
		if id := strings.TrimSpace(r.Header.Get(headerName)); id != "" {
			return id
		}
	}
	if fallback != "" {
		if id := strings.TrimSpace(r.Header.Get(fallback)); id != "" {
			return id
		}
	}
	return ""
}
