package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
)

// SimpleAuth returns a Middleware that validates the request token against the
// configured token using constant-time comparison.
//
// Tokens are accepted from two sources, checked in order:
//  1. Authorization header with Bearer prefix: "Authorization: Bearer <token>"
//  2. X-Token header: "X-Token: <token>"
//
// token must be non-empty; if empty, all requests are rejected (fail-closed).
//
// Example:
//
//	app.Use(auth.SimpleAuth(os.Getenv("API_TOKEN")))
func SimpleAuth(token string) middleware.Middleware {
	token = strings.TrimSpace(token)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" || subtle.ConstantTimeCompare([]byte(extractToken(r)), []byte(token)) != 1 {
				writeUnauthorized(w, r, "Protected Area")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) string {
	if r == nil {
		return ""
	}

	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authz) > len("Bearer ") && strings.EqualFold(authz[:len("Bearer")], "Bearer") {
		if tok := strings.TrimSpace(authz[len("Bearer"):]); tok != "" {
			return tok
		}
	}

	return strings.TrimSpace(r.Header.Get("X-Token"))
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request, realm string) {
	w.Header().Set("WWW-Authenticate", "Bearer realm=\""+realm+"\"")
	middleware.WriteTransportError(
		w,
		r,
		http.StatusUnauthorized,
		middleware.CodeAuthUnauthenticated,
		"invalid or missing authentication token",
		contract.CategoryAuth,
		nil,
	)
}
