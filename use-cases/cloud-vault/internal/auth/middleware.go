package auth

import (
	"context"
	"net/http"
	"time"
)

// Middleware creates authentication middleware.
func Middleware(svc *Service, cookieName string, requireAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract session token from cookie
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				if requireAuth {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			token := cookie.Value
			if token == "" {
				if requireAuth {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Validate session
			user, session, err := svc.ValidateSession(r.Context(), token)
			if err != nil {
				if requireAuth {
					// Clear invalid cookie
					http.SetCookie(w, &http.Cookie{
						Name:     cookieName,
						Value:    "",
						Path:     "/",
						Expires:  time.Unix(0, 0),
						HttpOnly: true,
						Secure:   true,
					})
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Add user and session to context
			ctx := context.WithValue(r.Context(), userContextKey, user)
			ctx = context.WithValue(ctx, sessionContextKey, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth creates middleware that optionally authenticates users.
// If a valid session exists, the user is added to context.
// If no session or invalid session, the request continues without authentication.
func OptionalAuth(svc *Service, cookieName string) func(http.Handler) http.Handler {
	return Middleware(svc, cookieName, false)
}

// RequireAuth creates middleware that requires authentication.
// Requests without valid sessions are rejected with 401 Unauthorized.
func RequireAuth(svc *Service, cookieName string) func(http.Handler) http.Handler {
	return Middleware(svc, cookieName, true)
}
