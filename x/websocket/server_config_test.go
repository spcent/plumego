package websocket

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeWSWithConfig_InvalidConfig(t *testing.T) {
	t.Run("nil hub", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/ws", nil)

		ServeWSWithConfig(w, r, ServerConfig{
			Auth: NewSimpleRoomAuth([]byte("secret")),
		})

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("nil auth", func(t *testing.T) {
		hub := NewHub(1, 4)
		defer hub.Stop()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/ws", nil)

		ServeWSWithConfig(w, r, ServerConfig{
			Hub: hub,
		})

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("negative queue size", func(t *testing.T) {
		hub := NewHub(1, 4)
		defer hub.Stop()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/ws", nil)

		ServeWSWithConfig(w, r, ServerConfig{
			Hub:       hub,
			Auth:      NewSimpleRoomAuth([]byte("secret")),
			QueueSize: -1,
		})

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestResolveValidationConfig(t *testing.T) {
	cfg := ServerConfig{
		ReadLimit: 1024,
		MessageValidation: MessageValidationConfig{
			MaxLength:               4096,
			AllowEmpty:              true,
			RejectControlCharacters: true,
			RequireValidUTF8:        true,
		},
	}

	got := resolveValidationConfig(cfg)
	if got.MaxLength != 1024 {
		t.Fatalf("expected MaxLength=1024, got %d", got.MaxLength)
	}
	if !got.AllowEmpty {
		t.Fatal("expected AllowEmpty to be preserved")
	}
}

func TestNormalizeServerConfig_ReadLimitFromAuth(t *testing.T) {
	secret := bytes.Repeat([]byte("a"), 32)
	auth, err := NewSecureRoomAuth(secret, SecurityConfig{
		JWTSecret:          secret,
		MinJWTSecretLength: 32,
		MaxMessageSize:     2048,
	})
	if err != nil {
		t.Fatalf("NewSecureRoomAuth error: %v", err)
	}

	hub := NewHub(1, 4)
	defer hub.Stop()

	cfg, err := normalizeServerConfig(ServerConfig{
		Hub:          hub,
		Auth:         auth,
		SendBehavior: SendBlock,
	})
	if err != nil {
		t.Fatalf("normalizeServerConfig error: %v", err)
	}
	if cfg.ReadLimit != 2048 {
		t.Fatalf("expected read limit from auth (2048), got %d", cfg.ReadLimit)
	}
}

func TestNewSecureRoomAuth_SecretMismatch(t *testing.T) {
	secretA := bytes.Repeat([]byte("a"), 32)
	secretB := bytes.Repeat([]byte("b"), 32)

	_, err := NewSecureRoomAuth(secretA, SecurityConfig{
		JWTSecret:          secretB,
		MinJWTSecretLength: 32,
	})
	if err == nil {
		t.Fatal("expected error for mismatched secrets")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}
