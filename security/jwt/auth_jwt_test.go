package jwt

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

// --- mapJWTError tests (covers all branches at 0%) ---

func TestMapJWTError_AllBranches(t *testing.T) {
	tests := []struct {
		name  string
		input error
		want  error
	}{
		{"ErrTokenExpired", ErrTokenExpired, contract.ErrExpiredToken},
		{"ErrTokenNotYetValid", ErrTokenNotYetValid, contract.ErrInvalidToken},
		{"ErrTokenRevoked", ErrTokenRevoked, contract.ErrSessionRevoked},
		{"ErrVersionMismatch", ErrVersionMismatch, contract.ErrTokenVersionMismatch},
		{"ErrInvalidIssuer", ErrInvalidIssuer, contract.ErrInvalidToken},
		{"ErrInvalidAudience", ErrInvalidAudience, contract.ErrInvalidToken},
		{"ErrUnknownKey", ErrUnknownKey, contract.ErrInvalidToken},
		{"ErrMissingSubject", ErrMissingSubject, contract.ErrInvalidToken},
		{"ErrInvalidToken", ErrInvalidToken, contract.ErrInvalidToken},
		{"unknown error", errors.New("unexpected"), contract.ErrInvalidToken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapJWTError(tt.input)
			if !errors.Is(got, tt.want) {
				t.Errorf("mapJWTError(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- Authenticator struct (contract.Authenticator) ---

func TestAuthenticator_NilManager(t *testing.T) {
	a := Authenticator{Manager: nil, ExpectedType: TokenTypeAccess}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer sometoken")

	_, err := a.Authenticate(req)
	if !errors.Is(err, contract.ErrUnauthenticated) {
		t.Errorf("nil Manager: expected ErrUnauthenticated, got %v", err)
	}
}

func TestAuthenticator_NilRequest_Direct(t *testing.T) {
	store := newTestStore(t)
	mgr, err := NewJWTManager(store, DefaultJWTConfig())
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	a := Authenticator{Manager: mgr, ExpectedType: TokenTypeAccess}

	_, err = a.Authenticate(nil)
	if !errors.Is(err, contract.ErrUnauthenticated) {
		t.Errorf("nil request: expected ErrUnauthenticated, got %v", err)
	}
}

func TestAuthenticator_ExpiredToken_MapsToExpiredToken(t *testing.T) {
	store := newTestStore(t)
	mgr, err := NewJWTManager(store, DefaultJWTConfig())
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	a := mgr.Authenticator(TokenTypeAccess)

	// Present an invalid/garbage token — should map to ErrInvalidToken.
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer notavalidtoken.x.y")

	_, err = a.Authenticate(req)
	if !errors.Is(err, contract.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for garbage token, got %v", err)
	}
}

// --- PrincipalFromClaims edge cases (0% originally) ---

func TestPrincipalFromClaims_Nil(t *testing.T) {
	got := PrincipalFromClaims(nil)
	if got != nil {
		t.Fatalf("expected nil for nil claims, got %+v", got)
	}
}

func TestPrincipalFromClaims_WithAllFields(t *testing.T) {
	claims := &TokenClaims{
		TokenID:   "tok-1",
		TokenType: TokenTypeAccess,
		KeyID:     "key-abc",
		Identity:  IdentityClaims{Subject: "user-1"},
		Authorization: AuthorizationClaims{
			Roles:       []string{"admin"},
			Permissions: []string{"read:all"},
		},
	}
	p := PrincipalFromClaims(claims)
	if p == nil {
		t.Fatal("expected non-nil principal")
	}
	if p.Subject != "user-1" {
		t.Errorf("Subject = %q, want user-1", p.Subject)
	}
	if len(p.Roles) != 1 || p.Roles[0] != "admin" {
		t.Errorf("Roles = %v, want [admin]", p.Roles)
	}
	if len(p.Scopes) != 1 || p.Scopes[0] != "read:all" {
		t.Errorf("Scopes = %v, want [read:all]", p.Scopes)
	}
	if p.Claims["token_id"] != "tok-1" {
		t.Errorf("Claims[token_id] = %q, want tok-1", p.Claims["token_id"])
	}
	if p.Claims["key_id"] != "key-abc" {
		t.Errorf("Claims[key_id] = %q, want key-abc", p.Claims["key_id"])
	}
}

func TestPrincipalFromClaims_NoExtras(t *testing.T) {
	claims := &TokenClaims{
		Identity: IdentityClaims{Subject: "bare-user"},
	}
	p := PrincipalFromClaims(claims)
	if p == nil {
		t.Fatal("expected non-nil principal")
	}
	// No token_id, token_type, or key_id — Claims should be nil.
	if p.Claims != nil {
		t.Errorf("expected nil Claims for bare claims, got %v", p.Claims)
	}
}

// --- RotateKey coverage ---

func TestRotateKey(t *testing.T) {
	store := newTestStore(t)
	mgr, err := NewJWTManager(store, DefaultJWTConfig())
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}

	key, err := mgr.RotateKey()
	if err != nil {
		t.Fatalf("RotateKey: %v", err)
	}
	if key.ID == "" {
		t.Error("expected non-empty key ID after rotation")
	}

	// Token issued with new key should verify successfully.
	pair, err := mgr.GenerateTokenPair(context.Background(),
		IdentityClaims{Subject: "u"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair after rotation: %v", err)
	}

	_, err = mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("VerifyToken after rotation: %v", err)
	}
}

func TestRotateKey_EdDSA(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.Algorithm = AlgorithmEdDSA
	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("create EdDSA manager: %v", err)
	}

	_, err = mgr.RotateKey()
	if err != nil {
		t.Fatalf("RotateKey EdDSA: %v", err)
	}
}

// --- PolicyAuthorizer nil principal ---

func TestPolicyAuthorizer_NilPrincipal(t *testing.T) {
	pa := PolicyAuthorizer{Policy: AuthZPolicy{AllRoles: []string{"admin"}}}
	err := pa.Authorize(nil, "", "")
	if !errors.Is(err, contract.ErrUnauthenticated) {
		t.Errorf("nil principal: expected ErrUnauthenticated, got %v", err)
	}
}

// --- PermissionAuthorizer edge cases ---

func TestPermissionAuthorizer_NilPrincipal(t *testing.T) {
	pa := PermissionAuthorizer{}
	err := pa.Authorize(nil, "read", "resource")
	if !errors.Is(err, contract.ErrUnauthenticated) {
		t.Errorf("nil principal: expected ErrUnauthenticated, got %v", err)
	}
}

func TestPermissionAuthorizer_EmptyActionResource(t *testing.T) {
	pa := PermissionAuthorizer{}
	p := &contract.Principal{Scopes: []string{"read:docs"}}
	err := pa.Authorize(p, "", "")
	if err != nil {
		t.Errorf("empty action+resource should pass: %v", err)
	}
}

func TestPermissionAuthorizer_ActionOnly(t *testing.T) {
	pa := PermissionAuthorizer{Separator: ":"}
	p := &contract.Principal{Scopes: []string{"write"}}
	err := pa.Authorize(p, "write", "")
	if err != nil {
		t.Errorf("action only: expected nil, got %v", err)
	}
}

func TestPermissionAuthorizer_CustomSeparator(t *testing.T) {
	pa := PermissionAuthorizer{Separator: "/"}
	p := &contract.Principal{Scopes: []string{"read/docs"}}
	err := pa.Authorize(p, "read", "docs")
	if err != nil {
		t.Errorf("custom separator: expected nil, got %v", err)
	}
}
