package handler

import (
	"crypto/subtle"
	"net/http"

	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"
)

// RequireMetricsToken gates an endpoint behind a static bearer token compared
// in constant time. An empty configured token disables the guard (the app
// logs a startup warning in that case).
func RequireMetricsToken(token string, logger plumelog.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if token == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			presented := authn.ExtractBearerToken(r)
			if presented == "" || subtle.ConstantTimeCompare([]byte(presented), []byte(token)) != 1 {
				writeUnauthorized(w, r, logger, "metrics.invalid_token", "missing or invalid metrics token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
