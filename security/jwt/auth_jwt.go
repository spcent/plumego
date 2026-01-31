package jwt

import (
	"errors"
	"net/http"

	"github.com/spcent/plumego/contract"
)

// Authenticator adapts JWT verification to the contract.Authenticator interface.
type Authenticator struct {
	Manager      *JWTManager
	ExpectedType TokenType
}

// Authenticate verifies the request token and returns a principal.
func (a Authenticator) Authenticate(r *http.Request) (*contract.Principal, error) {
	if a.Manager == nil {
		return nil, contract.ErrUnauthenticated
	}
	if r == nil {
		return nil, contract.ErrUnauthenticated
	}

	token := extractBearerToken(r)
	if token == "" {
		return nil, contract.ErrUnauthenticated
	}

	claims, err := a.Manager.VerifyToken(r.Context(), token, a.ExpectedType)
	if err != nil {
		return nil, mapJWTError(err)
	}

	return PrincipalFromClaims(claims), nil
}

// Authenticator returns a contract.Authenticator for the given token type.
func (m *JWTManager) Authenticator(tokenType TokenType) contract.Authenticator {
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
func (p PolicyAuthorizer) Authorize(principal *contract.Principal, _, _ string) error {
	if principal == nil {
		return contract.ErrUnauthenticated
	}

	authz := AuthorizationClaims{
		Roles:       principal.Roles,
		Permissions: principal.Scopes,
	}
	if !checkPolicy(p.Policy, authz) {
		return contract.ErrUnauthorized
	}
	return nil
}

// PermissionAuthorizer matches "action:resource" against principal scopes.
type PermissionAuthorizer struct {
	Separator string
}

// Authorize checks the action/resource permission in principal scopes.
func (p PermissionAuthorizer) Authorize(principal *contract.Principal, action string, resource string) error {
	if principal == nil {
		return contract.ErrUnauthenticated
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
	return contract.ErrUnauthorized
}

// PrincipalFromClaims converts JWT claims to a contract.Principal.
func PrincipalFromClaims(claims *TokenClaims) *contract.Principal {
	if claims == nil {
		return nil
	}
	principal := &contract.Principal{
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
		return contract.ErrExpiredToken
	case errors.Is(err, ErrTokenNotYetValid):
		return contract.ErrInvalidToken
	case errors.Is(err, ErrTokenRevoked):
		return contract.ErrSessionRevoked
	case errors.Is(err, ErrVersionMismatch):
		return contract.ErrTokenVersionMismatch
	case errors.Is(err, ErrInvalidIssuer):
		return contract.ErrInvalidToken
	case errors.Is(err, ErrInvalidAudience):
		return contract.ErrInvalidToken
	case errors.Is(err, ErrUnknownKey):
		return contract.ErrInvalidToken
	case errors.Is(err, ErrMissingSubject):
		return contract.ErrInvalidToken
	case errors.Is(err, ErrInvalidToken):
		return contract.ErrInvalidToken
	default:
		return contract.ErrInvalidToken
	}
}
