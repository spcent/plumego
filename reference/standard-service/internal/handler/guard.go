package handler

import (
	"crypto/subtle"
	"net/http"

	"github.com/spcent/plumego/contract"
)

// WriteKeyHeader is the request header that carries the write-operation key.
const WriteKeyHeader = "X-Write-Key"

const codeAuthKeyInvalid = "auth.key.invalid"

// RequireWriteKey returns a middleware that gates mutating operations behind a
// static bearer key supplied via the X-Write-Key header.
//
// When key is empty the middleware is a no-op, allowing unauthenticated access
// during local development. Set APP_WRITE_KEY in production to enforce the guard.
//
// This demonstrates the per-route middleware wrapping pattern: callers wrap only
// the handlers that need protection rather than adding the check globally.
// Timing-safe comparison (crypto/subtle.ConstantTimeCompare) prevents attackers
// from using response time to guess the key byte-by-byte.
//
//	mux.Post("/api/v1/items", RequireWriteKey(cfg.WriteKey)(http.HandlerFunc(items.Create)))
func RequireWriteKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if key == "" {
			// Guard disabled — pass through without any header check.
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerVal := r.Header.Get(WriteKeyHeader)
			if subtle.ConstantTimeCompare([]byte(headerVal), []byte(key)) != 1 {
				_ = contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeUnauthorized).
					Code(codeAuthKeyInvalid).
					Message("valid X-Write-Key header required").
					Build())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
