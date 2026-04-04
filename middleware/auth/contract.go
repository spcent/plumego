package auth

import (
	"errors"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
)

// AuthErrorHandler handles authentication/authorization errors.
type AuthErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

type authOptions struct {
	errorHandler AuthErrorHandler
	realm        string
}

// AuthOption configures authentication middleware behavior.
type AuthOption func(*authOptions)

// WithAuthErrorHandler overrides the default error handling.
func WithAuthErrorHandler(handler AuthErrorHandler) AuthOption {
	return func(o *authOptions) {
		if handler != nil {
			o.errorHandler = handler
		}
	}
}

// WithAuthRealm sets the WWW-Authenticate realm header value for 401 responses.
func WithAuthRealm(realm string) AuthOption {
	return func(o *authOptions) {
		o.realm = realm
	}
}

// Authenticate validates the request and stores the principal in context.
func Authenticate(authenticator contract.Authenticator, opts ...AuthOption) middleware.Middleware {
	cfg := applyAuthOptions(opts...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authenticator == nil {
				writeAuthInternal(w, r, "authenticator is nil")
				return
			}

			principal, err := authenticator.Authenticate(r)
			if err != nil {
				cfg.errorHandler(w, r, err)
				return
			}
			if principal == nil {
				cfg.errorHandler(w, r, contract.ErrUnauthenticated)
				return
			}

			ctx := contract.WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Authorize enforces access for the provided action/resource pair.
func Authorize(authorizer contract.Authorizer, action, resource string, opts ...AuthOption) middleware.Middleware {
	return AuthorizeFunc(authorizer, func(*http.Request) (string, string) {
		return action, resource
	}, opts...)
}

// AuthorizeFunc enforces access using a resolver for action/resource values.
func AuthorizeFunc(authorizer contract.Authorizer, resolver func(*http.Request) (string, string), opts ...AuthOption) middleware.Middleware {
	cfg := applyAuthOptions(opts...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authorizer == nil {
				writeAuthInternal(w, r, "authorizer is nil")
				return
			}
			if resolver == nil {
				writeAuthInternal(w, r, "authorize resolver is nil")
				return
			}

			principal := contract.PrincipalFromRequest(r)
			if principal == nil {
				cfg.errorHandler(w, r, contract.ErrUnauthenticated)
				return
			}

			action, resource := resolver(r)
			if err := authorizer.Authorize(principal, action, resource); err != nil {
				cfg.errorHandler(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func applyAuthOptions(opts ...AuthOption) authOptions {
	cfg := authOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.errorHandler == nil {
		cfg.errorHandler = defaultAuthErrorHandler(cfg.realm)
	}
	return cfg
}

func writeAuthInternal(w http.ResponseWriter, r *http.Request, message string) {
	contract.WriteError(w, r, contract.NewInternalError(message))
}

func defaultAuthErrorHandler(realm string) AuthErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		if err == nil {
			err = contract.ErrUnauthenticated
		}

		var apiErr contract.APIError
		if errors.As(err, &apiErr) {
			contract.WriteError(w, r, apiErr)
			return
		}

		apiErr = authErrorToAPIError(err)
		if apiErr.Status == http.StatusUnauthorized && realm != "" {
			w.Header().Set("WWW-Authenticate", `Bearer realm="`+realm+`"`)
		}
		contract.WriteError(w, r, apiErr)
	}
}

func authErrorToAPIError(err error) contract.APIError {
	switch {
	case errors.Is(err, contract.ErrUnauthorized):
		return contract.NewForbiddenError("access forbidden")
	case errors.Is(err, contract.ErrInvalidToken):
		return contract.NewUnauthorizedError("invalid token")
	case errors.Is(err, contract.ErrExpiredToken):
		return contract.NewUnauthorizedError("token expired")
	case errors.Is(err, contract.ErrSessionRevoked):
		return contract.NewUnauthorizedError("session revoked")
	case errors.Is(err, contract.ErrSessionExpired):
		return contract.NewUnauthorizedError("session expired")
	case errors.Is(err, contract.ErrRefreshReused):
		return contract.NewUnauthorizedError("refresh token reuse detected")
	case errors.Is(err, contract.ErrTokenVersionMismatch):
		return contract.NewUnauthorizedError("token version mismatch")
	case errors.Is(err, contract.ErrUnauthenticated):
		fallthrough
	default:
		return contract.NewUnauthorizedError("authentication required")
	}
}
