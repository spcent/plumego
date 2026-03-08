package conformance_test

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
)

func fixtureNoOpMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func fixtureDoubleNextMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			next.ServeHTTP(w, r)
		})
	}
}

func fixtureSuccessWritingMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"middleware":true}`))
			next.ServeHTTP(w, r)
		})
	}
}

func fixtureCanonicalErrorMiddleware(code string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contract.WriteError(w, r, contract.APIError{
				Code:     code,
				Message:  "conformance fixture error",
				Category: contract.CategoryServer,
				Status:   http.StatusInternalServerError,
			})
		})
	}
}
