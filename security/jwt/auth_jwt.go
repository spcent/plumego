package jwt

import (
	"errors"
	"net/http"

	"github.com/spcent/plumego/security/authn"
)

// Authenticator adapts JWT verification to the authn.Authenticator interface.
type Authenticator struct {
	Manager      *JWTManager
	ExpectedType TokenType
}

// Authenticate verifies the request token and returns a principal.
func (a Authenticator) Authenticate(r *http.Request) (*authn.Principal, error) {
	principal, _, err := a.AuthenticateRequest(r)
	return principal, err
}

// AuthenticateRequest verifies the request token and enriches the request context with token claims.
func (a Authenticator) AuthenticateRequest(r *http.Request) (*authn.Principal, *http.Request, error) {
	if a.Manager == nil {
		return nil, r, authn.ErrUnauthenticated
	}
	if r == nil {
		return nil, r, authn.ErrUnauthenticated
	}

	token := extractBearerToken(r)
	if token == "" {
		return nil, r, authn.ErrUnauthenticated
	}

	claims, err := a.Manager.VerifyToken(r.Context(), token, a.ExpectedType)
	if err != nil {
		return nil, r, mapJWTError(err)
	}

	return PrincipalFromClaims(claims), r.WithContext(WithTokenClaims(r.Context(), claims)), nil
}

// Authenticator returns an authn.Authenticator for the given token type.
func (m *JWTManager) Authenticator(tokenType TokenType) authn.Authenticator {
	return Authenticator{
		Manager:      m,
		ExpectedType: tokenType,
	}
}

// PolicyAuthorizer enforces an AuthZPolicy against principal roles/scopes.
type PolicyAuthorizer struct {
	Policy AuthZPolicy
}

// Authorize checks the principal against the configured policy.
func (p PolicyAuthorizer) Authorize(principal *authn.Principal, _, _ string) error {
	if principal == nil {
		return authn.ErrUnauthenticated
	}

	authz := AuthorizationClaims{
		Roles:       principal.Roles,
		Permissions: principal.Scopes,
	}
	if !checkPolicy(p.Policy, authz) {
		return authn.ErrUnauthorized
	}
	return nil
}

// PermissionAuthorizer matches "action:resource" against principal scopes.
type PermissionAuthorizer struct {
	Separator string
}

// Authorize checks the action/resource permission in principal scopes.
func (p PermissionAuthorizer) Authorize(principal *authn.Principal, action string, resource string) error {
	if principal == nil {
		return authn.ErrUnauthenticated
	}
	if action == "" && resource == "" {
		return nil
	}

	sep := p.Separator
	if sep == "" {
		sep = ":"
	}
	perm := action
	if resource != "" {
		perm = action + sep + resource
	}

	if contains(principal.Scopes, perm) {
		return nil
	}
	return authn.ErrUnauthorized
}

// PrincipalFromClaims converts JWT claims to an authn.Principal.
func PrincipalFromClaims(claims *TokenClaims) *authn.Principal {
	if claims == nil {
		return nil
	}
	principal := &authn.Principal{
		Subject: claims.Identity.Subject,
		Roles:   claims.Authorization.Roles,
		Scopes:  claims.Authorization.Permissions,
		Claims:  map[string]string{},
	}
	if claims.TokenID != "" {
		principal.Claims["token_id"] = claims.TokenID
	}
	if claims.TokenType != "" {
		principal.Claims["token_type"] = string(claims.TokenType)
	}
	if claims.KeyID != "" {
		principal.Claims["key_id"] = claims.KeyID
	}
	if len(principal.Claims) == 0 {
		principal.Claims = nil
	}
	return principal
}

func mapJWTError(err error) error {
	switch {
	case errors.Is(err, ErrTokenExpired):
		return authn.ErrExpiredToken
	case errors.Is(err, ErrTokenNotYetValid):
		return authn.ErrInvalidToken
	case errors.Is(err, ErrInvalidIssuer):
		return authn.ErrInvalidToken
	case errors.Is(err, ErrInvalidAudience):
		return authn.ErrInvalidToken
	case errors.Is(err, ErrUnknownKey):
		return authn.ErrInvalidToken
	case errors.Is(err, ErrMissingSubject):
		return authn.ErrInvalidToken
	case errors.Is(err, ErrInvalidToken):
		return authn.ErrInvalidToken
	default:
		return authn.ErrInvalidToken
	}
}
