package authn

import (
	"context"
	"errors"
	"net/http"
)

// Principal is the authenticated identity attached to a request context.
type Principal struct {
	Subject  string
	TenantID string
	Roles    []string
	Scopes   []string
	Claims   map[string]string
}

// Authenticator validates a request and returns the authenticated principal.
type Authenticator interface {
	Authenticate(r *http.Request) (*Principal, error)
}

// Authorizer checks whether a principal may perform action on resource.
type Authorizer interface {
	Authorize(p *Principal, action string, resource string) error
}

var (
	ErrUnauthenticated      = errors.New("unauthenticated")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("expired token")
	ErrSessionRevoked       = errors.New("session revoked")
	ErrSessionExpired       = errors.New("session expired")
	ErrRefreshReused        = errors.New("refresh token reuse detected")
	ErrTokenVersionMismatch = errors.New("token version mismatch")
)

type principalContextKey struct{}

// WithPrincipal attaches a principal to a context.
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, p)
}

// PrincipalFromContext extracts a principal from a context.
func PrincipalFromContext(ctx context.Context) *Principal {
	if v := ctx.Value(principalContextKey{}); v != nil {
		if p, ok := v.(*Principal); ok {
			return p
		}
	}
	return nil
}

// PrincipalFromRequest extracts a principal from a request context.
func PrincipalFromRequest(r *http.Request) *Principal {
	if r == nil {
		return nil
	}
	return PrincipalFromContext(r.Context())
}

// RequestWithPrincipal returns a shallow copy of r with the principal attached.
func RequestWithPrincipal(r *http.Request, p *Principal) *http.Request {
	if r == nil {
		return nil
	}
	return r.WithContext(WithPrincipal(r.Context(), p))
}
