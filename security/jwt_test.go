package security

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

func newTestStore(t *testing.T) *kvstore.KVStore {
	t.Helper()
	store, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("failed to create kv store: %v", err)
	}
	return store
}

func TestGenerateAndVerifyTokenPair(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.AccessExpiration = time.Minute
	cfg.RefreshExpiration = time.Hour

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-1"}, AuthorizationClaims{Roles: []string{"user"}, Permissions: []string{"read"}})
	if err != nil {
		t.Fatalf("failed to generate pair: %v", err)
	}

	accessClaims, err := mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("failed to verify access token: %v", err)
	}
	if accessClaims.Identity.Subject != "user-1" || accessClaims.TokenType != TokenTypeAccess {
		t.Fatalf("unexpected claims: %+v", accessClaims)
	}

	refreshClaims, err := mgr.VerifyToken(context.Background(), pair.RefreshToken, TokenTypeRefresh)
	if err != nil {
		t.Fatalf("failed to verify refresh token: %v", err)
	}
	if refreshClaims.TokenType != TokenTypeRefresh {
		t.Fatalf("expected refresh token, got %s", refreshClaims.TokenType)
	}
}

func TestKeyRotationAndVerification(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.RotationInterval = time.Second

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-2"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}
	originalKid := mustExtractKid(t, pair.AccessToken)

	// Force rotation
	if _, err := mgr.RotateKey(); err != nil {
		t.Fatalf("rotate key: %v", err)
	}

	// Old token should still verify because old key is kept
	if _, err := mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess); err != nil {
		t.Fatalf("old token failed verification after rotation: %v", err)
	}

	// New token should be signed with rotated key
	newPair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-2"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate new pair: %v", err)
	}
	newKid := mustExtractKid(t, newPair.AccessToken)
	if newKid == originalKid {
		t.Fatalf("expected new key id after rotation, got %s", newKid)
	}
}

func TestBlacklistAndVersioning(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig(nil)
	cfg.AccessExpiration = time.Minute

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-3"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}

	if err := mgr.RevokeToken(pair.AccessToken); err != nil {
		t.Fatalf("revoke token: %v", err)
	}
	if _, err := mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess); !errorsIs(err, ErrTokenRevoked) {
		t.Fatalf("expected ErrTokenRevoked, got %v", err)
	}

	// Issue again and bump version
	pair, err = mgr.GenerateTokenPair(context.Background(), IdentityClaims{Subject: "user-3"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}
	if err := mgr.IncrementIdentityVersion("user-3"); err != nil {
		t.Fatalf("increment version: %v", err)
	}
	if _, err := mgr.VerifyToken(context.Background(), pair.AccessToken, TokenTypeAccess); !errorsIs(err, ErrVersionMismatch) {
		t.Fatalf("expected ErrVersionMismatch, got %v", err)
	}
}

func TestCheckPolicy(t *testing.T) {
	policy := AuthZPolicy{
		AllRoles:       []string{"admin"},
		AnyPermission:  []string{"write"},
		AllPermissions: []string{"read"},
	}
	auth := AuthorizationClaims{Roles: []string{"admin", "user"}, Permissions: []string{"read", "write"}}
	if !checkPolicy(policy, auth) {
		t.Fatalf("policy should pass")
	}

	if checkPolicy(policy, AuthorizationClaims{Roles: []string{"user"}, Permissions: []string{"read"}}) {
		t.Fatalf("policy should fail without required role")
	}
}

func mustExtractKid(t *testing.T, token string) string {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid token format")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	var hdr map[string]any
	if err := json.Unmarshal(headerBytes, &hdr); err != nil {
		t.Fatalf("unmarshal header: %v", err)
	}
	kid, _ := hdr["kid"].(string)
	return kid
}

func errorsIs(err, target error) bool {
	return err == target
}
