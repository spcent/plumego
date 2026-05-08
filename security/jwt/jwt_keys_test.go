package jwt

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewJWTManagerRecoversMissingPersistedActiveKey(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-stale-active"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}
	staleKid := mustExtractKid(t, pair.AccessToken)
	if err := store.Delete(keyPrefix + staleKid); err != nil {
		t.Fatalf("delete active key: %v", err)
	}
	if err := store.Set(activeKeyKey, []byte(staleKid), 0); err != nil {
		t.Fatalf("set stale active key: %v", err)
	}

	recovered, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("NewJWTManager with stale active key: %v", err)
	}
	if recovered.active == "" || recovered.active == staleKid {
		t.Fatalf("active key = %q, want new key different from %q", recovered.active, staleKid)
	}
	if _, ok := recovered.keyCache[recovered.active]; !ok {
		t.Fatalf("recovered active key %q missing from cache", recovered.active)
	}

	newPair, err := recovered.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-recovered"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair after recovery: %v", err)
	}
	if _, err := recovered.VerifyToken(t.Context(), newPair.AccessToken, TokenTypeAccess); err != nil {
		t.Fatalf("VerifyToken after recovery: %v", err)
	}
}

func TestNewJWTManagerRecoversAbsentActiveKeyWithExistingSigningKeys(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("NewJWTManager: %v", err)
	}
	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-active-missing"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}
	originalKid := mustExtractKid(t, pair.AccessToken)
	if err := store.Delete(activeKeyKey); err != nil {
		t.Fatalf("delete active key marker: %v", err)
	}

	recovered, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("NewJWTManager with absent active key: %v", err)
	}
	if recovered.active == "" {
		t.Fatal("expected recovered active key")
	}
	if recovered.active == originalKid {
		t.Fatalf("active key = %q, want new key after absent active marker", recovered.active)
	}
	if _, ok := recovered.keyCache[originalKid]; !ok {
		t.Fatalf("existing signing key %q should remain cached", originalKid)
	}
}

func TestNewJWTManagerFailsOnPersistedActiveKeyReadError(t *testing.T) {
	store := newRecordingContextKeyStore()
	store.data[activeKeyKey] = []byte("kid")
	store.getContextErr = errors.New("active read failed")

	_, err := NewJWTManagerContext(t.Context(), store, DefaultJWTConfig())
	if err == nil {
		t.Fatal("expected active key read error")
	}
	if !strings.Contains(err.Error(), "failed to read active signing key") {
		t.Fatalf("expected active key read context, got %v", err)
	}
	if !errors.Is(err, store.getContextErr) {
		t.Fatalf("expected active read error in chain, got %v", err)
	}
}

func TestEdDSAAlgorithm(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.Algorithm = AlgorithmEdDSA

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-eddsa"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair with EdDSA: %v", err)
	}

	// verify header algorithm is EdDSA
	header := mustExtractHeader(t, pair.AccessToken)
	if header["alg"] != string(AlgorithmEdDSA) {
		t.Errorf("expected alg=EdDSA, got %v", header["alg"])
	}

	// verify token claims
	claims, err := mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("failed to verify EdDSA token: %v", err)
	}
	if claims.Identity.Subject != "user-eddsa" {
		t.Errorf("unexpected subject: %s", claims.Identity.Subject)
	}
}

func TestNewJWTManagerDefaultsEmptyAlgorithmToHS256(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.Algorithm = ""

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("NewJWTManager with empty algorithm: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-default-alg"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}
	header := mustExtractHeader(t, pair.AccessToken)
	if header["alg"] != string(AlgorithmHS256) {
		t.Fatalf("alg = %v, want %s", header["alg"], AlgorithmHS256)
	}
}

func TestJWTConfigRejectsNegativeDurations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*JWTConfig)
	}{
		{
			name: "negative rotation interval",
			mutate: func(cfg *JWTConfig) {
				cfg.RotationInterval = -time.Second
			},
		},
		{
			name: "negative clock skew",
			mutate: func(cfg *JWTConfig) {
				cfg.ClockSkew = -time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultJWTConfig()
			tt.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestNewJWTManagerContextUsesContextKeyStore(t *testing.T) {
	store := newRecordingContextKeyStore()
	want := errors.New("keys canceled")
	store.keysContextErr = want

	if _, err := NewJWTManagerContext(t.Context(), store, DefaultJWTConfig()); !errors.Is(err, want) {
		t.Fatalf("NewJWTManagerContext error = %v, want %v", err, want)
	}
	if store.keysContextCalls != 1 {
		t.Fatalf("KeysContext calls = %d, want 1", store.keysContextCalls)
	}
	if store.legacyKeysCalls != 0 || store.legacyGetCalls != 0 || store.legacySetCalls != 0 {
		t.Fatalf("legacy store methods used: keys=%d get=%d set=%d", store.legacyKeysCalls, store.legacyGetCalls, store.legacySetCalls)
	}
}

func TestNewJWTManagerRejectsMalformedPersistedSigningKey(t *testing.T) {
	store := newTestStore(t)
	raw, err := json.Marshal(JWTSigningKey{
		ID:        "bad-key",
		Algorithm: AlgorithmHS256,
		Secret:    []byte("short"),
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	if err := store.Set(keyPrefix+"bad-key", raw, 0); err != nil {
		t.Fatalf("set bad key: %v", err)
	}

	if _, err := NewJWTManager(store, DefaultJWTConfig()); err == nil {
		t.Fatalf("expected malformed persisted signing key to fail manager startup")
	}
}
