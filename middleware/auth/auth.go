package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
)

// AuthMiddleware defines an interface for authentication middleware.
//
// Implementations should authenticate the request and call the next handler if successful,
// or return an appropriate error response if authentication fails.
type AuthMiddleware interface {
	// Authenticate authenticates the request and calls the next handler if successful
	Authenticate(next http.Handler) http.Handler
}

// SimpleAuthMiddleware is a simple authentication middleware that validates
// explicit header-based credentials.
//
// This middleware is suitable for simple API authentication scenarios.
// For production use, consider using JWT or OAuth2-based authentication.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/auth"
//
//	// Or provide token directly
//	handler := auth.NewSimpleAuthMiddleware("my-secret-token").Authenticate(myHandler)
//
// The middleware checks for credentials in the following order:
//  1. Authorization header with Bearer token
//  2. X-Token header
//
// If authentication fails, it returns a 401 Unauthorized response with
// a WWW-Authenticate header and a JSON error message.
type SimpleAuthMiddleware struct {
	authToken string // Expected authentication token
	realm     string // Authentication realm for WWW-Authenticate header
}

// NewSimpleAuthMiddleware creates a new SimpleAuthMiddleware with the given authToken.
func NewSimpleAuthMiddleware(authToken string) *SimpleAuthMiddleware {
	return &SimpleAuthMiddleware{
		authToken: strings.TrimSpace(authToken),
		realm:     "Protected Area",
	}
}

// Authenticate implements the AuthMiddleware interface.
func (am *SimpleAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Fail closed if auth token is not configured.
		if am.authToken == "" {
			writeUnauthorized(w, r, am.realm)
			return
		}

		token := extractTokenFromHeaders(r)

		// Check if token matches expected token
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(am.authToken)) != 1 {
			writeUnauthorized(w, r, am.realm)
			return
		}

		// Authentication successful, call next handler
		next.ServeHTTP(w, r)
	})
}

func extractTokenFromHeaders(r *http.Request) string {
	if r == nil {
		return ""
	}

	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authz) >= len("Bearer ") && strings.EqualFold(authz[:len("Bearer")], "Bearer") {
		bearerToken := strings.TrimSpace(authz[len("Bearer"):])
		if bearerToken != "" {
			return bearerToken
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
		contract.CategoryAuthentication,
		nil,
	)
}

// Auth is a compatibility helper that applies a middleware with empty configuration.
//
// Deprecated: inject auth configuration from bootstrap using NewSimpleAuthMiddleware.
// This is a convenience function for backwards compatibility.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/auth"
//
//	// Prefer NewSimpleAuthMiddleware("my-secret-token") for explicit configuration.
//	handler := auth.Auth(myHandlerFunc)
func Auth(next http.HandlerFunc) http.HandlerFunc {
	middleware := NewSimpleAuthMiddleware("")
	return func(w http.ResponseWriter, r *http.Request) {
		middleware.Authenticate(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

// FromAuthMiddleware converts an AuthMiddleware to a Middleware.
//
// Example:
//
//	import (
//		"github.com/spcent/plumego/middleware"
//		"github.com/spcent/plumego/middleware/auth"
//	)
//
//	authMw := auth.NewSimpleAuthMiddleware("my-token")
//	handler := middleware.Apply(myHandler, auth.FromAuthMiddleware(authMw))
func FromAuthMiddleware(am AuthMiddleware) middleware.Middleware {
	return func(h http.Handler) http.Handler {
		return am.Authenticate(h)
	}
}
