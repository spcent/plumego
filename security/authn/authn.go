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

// RequestAuthenticator authenticates requests that may be enriched before
// reaching downstream handlers.
type RequestAuthenticator interface {
	Authenticator
	AuthenticateRequest(r *http.Request) (*Principal, *http.Request, error)
}

// Authorizer checks whether a principal may perform action on resource.
type Authorizer interface {
	Authorize(p *Principal, action string, resource string) error
}

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidToken    = errors.New("invalid token")
	ErrExpiredToken    = errors.New("expired token")
)

type principalContextKey struct{}

// WithPrincipal attaches a principal to a context.
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, principalContextKey{}, clonePrincipal(p))
}

// PrincipalFromContext extracts a principal from a context.
func PrincipalFromContext(ctx context.Context) *Principal {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(principalContextKey{}); v != nil {
		if p, ok := v.(*Principal); ok {
			return clonePrincipal(p)
		}
	}
	return nil
}

func clonePrincipal(p *Principal) *Principal {
	if p == nil {
		return nil
	}
	copied := *p
	copied.Roles = append([]string(nil), p.Roles...)
	copied.Scopes = append([]string(nil), p.Scopes...)
	if p.Claims != nil {
		copied.Claims = make(map[string]string, len(p.Claims))
		for k, v := range p.Claims {
			copied.Claims[k] = v
		}
	}
	return &copied
}
