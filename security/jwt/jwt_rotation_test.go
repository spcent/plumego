package jwt

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestKeyRotationAndVerification(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.RotationInterval = time.Second

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// generate first token pair
	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-2"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate pair: %v", err)
	}
	originalKid := mustExtractKid(t, pair.AccessToken)

	// force rotation of keys
	rotatedKey, err := mgr.RotateKey()
	if err != nil {
		t.Fatalf("rotate key: %v", err)
	}

	// old token should still be valid (old key is retained)
	_, err = mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("old token failed verification after rotation: %v", err)
	}

	// new token should use new key
	newPair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-2"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("generate new pair: %v", err)
	}
	newKid := mustExtractKid(t, newPair.AccessToken)

	if newKid == originalKid {
		t.Errorf("expected new key id after rotation, got same id: %s", newKid)
	}
	if newKid != rotatedKey.ID {
		t.Errorf("expected kid=%s, got %s", rotatedKey.ID, newKid)
	}
}

func TestAutomaticRotation(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.RotationInterval = 100 * time.Millisecond

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// generate first token pair
	pair1, _ := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-auto"}, AuthorizationClaims{})
	kid1 := mustExtractKid(t, pair1.AccessToken)

	// wait for rotation interval
	time.Sleep(150 * time.Millisecond)

	// generate new token pair should trigger automatic rotation
	pair2, _ := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-auto"}, AuthorizationClaims{})
	kid2 := mustExtractKid(t, pair2.AccessToken)

	if kid1 == kid2 {
		t.Errorf("expected automatic rotation, but kid remained the same")
	}
}

func TestRotateKeyContextUsesContextKeyStore(t *testing.T) {
	store := newRecordingContextKeyStore()
	mgr, err := NewJWTManagerContext(t.Context(), store, DefaultJWTConfig())
	if err != nil {
		t.Fatalf("NewJWTManagerContext: %v", err)
	}
	if store.keysContextCalls != 1 || store.setContextCalls == 0 {
		t.Fatalf("context store not used during startup: keys=%d sets=%d", store.keysContextCalls, store.setContextCalls)
	}

	want := errors.New("set canceled")
	store.setContextErr = want
	if _, err := mgr.RotateKeyContext(t.Context()); !errors.Is(err, want) {
		t.Fatalf("RotateKeyContext error = %v, want %v", err, want)
	}
}

func TestJWTManagerClockControlsRotationBoundary(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.RotationInterval = time.Hour
	fixed := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("NewJWTManager: %v", err)
	}
	mgr.now = func() time.Time { return fixed.Add(time.Hour - time.Second) }

	mgr.mu.Lock()
	initialActive := mgr.active
	activeKey := mgr.keyCache[initialActive]
	activeKey.CreatedAt = fixed
	mgr.keyCache[initialActive] = activeKey
	mgr.mu.Unlock()

	if _, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "rotation-user"}, AuthorizationClaims{}); err != nil {
		t.Fatalf("GenerateTokenPair before boundary: %v", err)
	}
	if mgr.active != initialActive {
		t.Fatalf("active key rotated before boundary: got %q want %q", mgr.active, initialActive)
	}

	mgr.now = func() time.Time { return fixed.Add(time.Hour) }
	if _, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "rotation-user"}, AuthorizationClaims{}); err != nil {
		t.Fatalf("GenerateTokenPair at boundary: %v", err)
	}
	if mgr.active == initialActive {
		t.Fatalf("active key did not rotate at configured boundary")
	}
}

func TestVerifyTokenRotationIsConcurrentSafe(t *testing.T) {
	store := newTestStore(t)
	cfg := DefaultJWTConfig()
	cfg.RotationInterval = time.Nanosecond
	mgr, err := NewJWTManager(store, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	pair, err := mgr.GenerateTokenPair(t.Context(), IdentityClaims{Subject: "user-race"}, AuthorizationClaims{})
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				claims, err := mgr.VerifyToken(t.Context(), pair.AccessToken, TokenTypeAccess)
				if err != nil {
					t.Errorf("VerifyToken: %v", err)
					return
				}
				if claims.Identity.Subject != "user-race" {
					t.Errorf("subject = %q, want user-race", claims.Identity.Subject)
					return
				}
			}
		}()
	}
	wg.Wait()
}
