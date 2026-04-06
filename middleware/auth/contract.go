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

type requestAuthenticator interface {
	AuthenticateRequest(r *http.Request) (*contract.Principal, *http.Request, error)
}

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

			req := r
			var (
				principal *contract.Principal
				err       error
			)
			if enriched, ok := authenticator.(requestAuthenticator); ok {
				principal, req, err = enriched.AuthenticateRequest(r)
				if req == nil {
					req = r
				}
			} else {
				principal, err = authenticator.Authenticate(r)
			}
			if err != nil {
				cfg.errorHandler(w, r, err)
				return
			}
			if principal == nil {
				cfg.errorHandler(w, r, contract.ErrUnauthenticated)
				return
			}

			ctx := contract.WithPrincipal(req.Context(), principal)
			next.ServeHTTP(w, req.WithContext(ctx))
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
	contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).
		Message(message).
		Build())
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
	unauthorized := func(msg string) contract.APIError {
		return contract.NewErrorBuilder().
			Status(http.StatusUnauthorized).
			Category(contract.CategoryAuth).
			Type(contract.TypeUnauthorized).
			Code(contract.CodeUnauthorized).
			Message(msg).
			Build()
	}
	switch {
	case errors.Is(err, contract.ErrUnauthorized):
		return contract.NewErrorBuilder().
			Status(http.StatusForbidden).
			Category(contract.CategoryAuth).
			Type(contract.TypeForbidden).
			Code(contract.CodeForbidden).
			Message("access forbidden").
			Build()
	case errors.Is(err, contract.ErrInvalidToken):
		return unauthorized("invalid token")
	case errors.Is(err, contract.ErrExpiredToken):
		return unauthorized("token expired")
	case errors.Is(err, contract.ErrSessionRevoked):
		return unauthorized("session revoked")
	case errors.Is(err, contract.ErrSessionExpired):
		return unauthorized("session expired")
	case errors.Is(err, contract.ErrRefreshReused):
		return unauthorized("refresh token reuse detected")
	case errors.Is(err, contract.ErrTokenVersionMismatch):
		return unauthorized("token version mismatch")
	case errors.Is(err, contract.ErrUnauthenticated):
		fallthrough
	default:
		return unauthorized("authentication required")
	}
}
