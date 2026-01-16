package middleware

import (
	"net/http"
	"os"
)

// AuthMiddleware defines an interface for authentication middleware.
//
// Implementations should authenticate the request and call the next handler if successful,
// or return an appropriate error response if authentication fails.
type AuthMiddleware interface {
	// Authenticate authenticates the request and calls the next handler if successful
	Authenticate(next http.Handler) http.Handler
}

// SimpleAuthMiddleware is a simple authentication middleware that uses X-Token header,
// query parameter, or AUTH_TOKEN environment variable for authentication.
//
// This middleware is suitable for simple API authentication scenarios.
// For production use, consider using JWT or OAuth2-based authentication.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	// Use environment variable
//	handler := middleware.NewSimpleAuthMiddleware("").Authenticate(myHandler)
//
//	// Or provide token directly
//	handler := middleware.NewSimpleAuthMiddleware("my-secret-token").Authenticate(myHandler)
//
// The middleware checks for the token in the following order:
//  1. X-Token header
//  2. URL query parameter "token"
//  3. Cookie named "auth_token"
//
// If authentication fails, it returns a 401 Unauthorized response with
// a WWW-Authenticate header and a JSON error message.
type SimpleAuthMiddleware struct {
	authToken string // Expected authentication token
	realm     string // Authentication realm for WWW-Authenticate header
}

// NewSimpleAuthMiddleware creates a new SimpleAuthMiddleware with the given authToken.
// If authToken is empty, it will be read from the AUTH_TOKEN environment variable.
func NewSimpleAuthMiddleware(authToken string) *SimpleAuthMiddleware {
	if authToken == "" {
		authToken = os.Getenv("AUTH_TOKEN")
	}
	return &SimpleAuthMiddleware{
		authToken: authToken,
		realm:     "Protected Area",
	}
}

// Authenticate implements the AuthMiddleware interface.
func (am *SimpleAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if no auth token is configured
		if am.authToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Get token from request headers, query parameters, or cookies
		token := r.Header.Get("X-Token")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token == "" {
			cookie, err := r.Cookie("auth_token")
			if err == nil {
				token = cookie.Value
			}
		}

		// Check if token matches expected token
		if token != am.authToken {
			w.Header().Set("WWW-Authenticate", "Bearer realm=\""+am.realm+"\"")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized","message":"Invalid or missing authentication token"}`))
			return
		}

		// Authentication successful, call next handler
		next.ServeHTTP(w, r)
	})
}

// Auth is a middleware function that validates the X-Token header against the AUTH_TOKEN environment variable when it is set.
// This is a convenience function for backwards compatibility.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	// Set AUTH_TOKEN environment variable before using
//	os.Setenv("AUTH_TOKEN", "my-secret-token")
//	handler := middleware.Auth(myHandlerFunc)
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
//	import "github.com/spcent/plumego/middleware"
//
//	auth := middleware.NewSimpleAuthMiddleware("my-token")
//	handler := middleware.Apply(myHandler, middleware.FromAuthMiddleware(auth))
func FromAuthMiddleware(am AuthMiddleware) Middleware {
	return func(h http.Handler) http.Handler {
		return am.Authenticate(h)
	}
}
