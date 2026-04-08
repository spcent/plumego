package requestid

import (
	"net/http"

	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
)

type config struct {
	generate         func() string
	includeInRequest bool
}

type Option func(*config)

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
		generate:         internalobs.NewRequestID,
		includeInRequest: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := internalobs.RequestIDFromRequest(r)
			if id == "" {
				id = cfg.generate()
			}

			r = internalobs.AttachRequestID(w, r, id, cfg.includeInRequest)
			next.ServeHTTP(w, r)
		})
	}
}
