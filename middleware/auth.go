package middleware

import (
	"net/http"
	"os"
)

// AuthMiddleware defines an interface for authentication middleware
type AuthMiddleware interface {
	// Authenticate authenticates the request and calls the next handler if successful
	Authenticate(next http.Handler) http.Handler
}

// SimpleAuthMiddleware is a simple authentication middleware that uses X-Token header
// or AUTH_TOKEN environment variable for authentication
type SimpleAuthMiddleware struct {
	authToken string // Expected authentication token
	realm     string // Authentication realm for WWW-Authenticate header
}

// NewSimpleAuthMiddleware creates a new SimpleAuthMiddleware with the given authToken
// If authToken is empty, it will be read from the AUTH_TOKEN environment variable
func NewSimpleAuthMiddleware(authToken string) *SimpleAuthMiddleware {
	if authToken == "" {
		authToken = os.Getenv("AUTH_TOKEN")
	}
	return &SimpleAuthMiddleware{
		authToken: authToken,
		realm:     "Protected Area",
	}
}

// Authenticate implements the AuthMiddleware interface
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
// This is a convenience function for backwards compatibility
func Auth(next http.HandlerFunc) http.HandlerFunc {
	middleware := NewSimpleAuthMiddleware("")
	return func(w http.ResponseWriter, r *http.Request) {
		middleware.Authenticate(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

// FromAuthMiddleware converts an AuthMiddleware to a Middleware
func FromAuthMiddleware(am AuthMiddleware) Middleware {
	return func(h Handler) Handler {
		return Handler(am.Authenticate(http.Handler(h)))
	}
}
