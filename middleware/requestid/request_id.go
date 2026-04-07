package requestid

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
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
		generate:         contract.NewRequestID,
		includeInRequest: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			observePolicy := contract.NewObservabilityPolicy()
			id := observePolicy.RequestIDFromRequest(r)
			if id == "" {
				id = cfg.generate()
			}

			r = observePolicy.AttachRequestID(w, r, id, cfg.includeInRequest)
			next.ServeHTTP(w, r)
		})
	}
}
