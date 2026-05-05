package websocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

const validTestWSKey = "dGhlIHNhbXBsZSBub25jZQ=="

type failingTokenAuth struct{}

func (failingTokenAuth) AuthenticateToken(string) (map[string]any, error) {
	return nil, ErrInvalidToken
}

func TestServeWSWithConfig_HandshakeErrorContract(t *testing.T) {
	tests := []struct {
		name        string
		cfg         func(*testing.T) ServerConfig
		req         func() *http.Request
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name: "invalid config",
			cfg: func(t *testing.T) ServerConfig {
				return ServerConfig{
					RoomAuth:             NewSimpleRoomAuth(),
					AllowUnauthenticated: true,
					OnMessage:            noopMessageHandler,
				}
			},
			req:         newValidHandshakeRequest,
			wantStatus:  http.StatusInternalServerError,
			wantCode:    codeWebSocketInvalidConfig,
			wantMessage: "websocket server misconfigured",
		},
		{
			name: "method not allowed",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/ws", nil)
			},
			wantStatus:  http.StatusMethodNotAllowed,
			wantCode:    contract.CodeMethodNotAllowed,
			wantMessage: "method not allowed",
		},
		{
			name: "bad upgrade",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/ws", nil)
			},
			wantStatus:  http.StatusBadRequest,
			wantCode:    codeWebSocketBadUpgrade,
			wantMessage: "websocket upgrade required",
		},
		{
			name: "missing websocket key",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/ws", nil)
				r.Header.Set("Connection", "Upgrade")
				r.Header.Set("Upgrade", "websocket")
				return r
			},
			wantStatus:  http.StatusBadRequest,
			wantCode:    codeWebSocketKeyMissing,
			wantMessage: "websocket key required",
		},
		{
			name: "invalid websocket key",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.Header.Set("Sec-WebSocket-Key", "bad-key")
				return r
			},
			wantStatus:  http.StatusBadRequest,
			wantCode:    codeWebSocketKeyInvalid,
			wantMessage: "invalid websocket key",
		},
		{
			name: "missing websocket version",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.Header.Del("Sec-WebSocket-Version")
				return r
			},
			wantStatus:  http.StatusBadRequest,
			wantCode:    codeWebSocketVersionUnsupported,
			wantMessage: "unsupported websocket version",
		},
		{
			name: "unsupported websocket version",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.Header.Set("Sec-WebSocket-Version", "12")
				return r
			},
			wantStatus:  http.StatusBadRequest,
			wantCode:    codeWebSocketVersionUnsupported,
			wantMessage: "unsupported websocket version",
		},
		{
			name: "forbidden origin",
			cfg: func(t *testing.T) ServerConfig {
				cfg := defaultHandshakeConfig(t)
				cfg.AllowedOrigins = []string{"https://allowed.example"}
				return cfg
			},
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.Header.Set("Origin", "https://blocked.example")
				return r
			},
			wantStatus:  http.StatusForbidden,
			wantCode:    codeWebSocketForbiddenOrigin,
			wantMessage: "forbidden origin",
		},
		{
			name: "origin requires explicit allowlist",
			cfg: func(t *testing.T) ServerConfig {
				cfg := defaultHandshakeConfig(t)
				cfg.AllowedOrigins = nil
				return cfg
			},
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.Header.Set("Origin", "https://app.example")
				return r
			},
			wantStatus:  http.StatusForbidden,
			wantCode:    codeWebSocketForbiddenOrigin,
			wantMessage: "forbidden origin",
		},
		{
			name: "room password denied",
			cfg: func(t *testing.T) ServerConfig {
				auth := NewSimpleRoomAuth()
				if err := auth.SetRoomPassword("private", "correct"); err != nil {
					t.Fatalf("SetRoomPassword: %v", err)
				}
				return ServerConfig{
					Hub:                  mustHub(t, 1, 4),
					RoomAuth:             auth,
					AllowUnauthenticated: true,
					SendBehavior:         SendBlock,
					AllowedOrigins:       []string{"*"},
					OnMessage:            noopMessageHandler,
				}
			},
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.URL.RawQuery = "room=private"
				r.Header.Set(RoomPasswordHeader, "wrong")
				return r
			},
			wantStatus:  http.StatusForbidden,
			wantCode:    codeWebSocketRoomForbidden,
			wantMessage: "websocket room access denied",
		},
		{
			name: "invalid room",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.URL.RawQuery = "room=bad/room"
				return r
			},
			wantStatus:  http.StatusBadRequest,
			wantCode:    codeWebSocketRoomInvalid,
			wantMessage: "invalid websocket room",
		},
		{
			name: "query room password rejected",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.URL.RawQuery = "room=private&room_password=wrong"
				return r
			},
			wantStatus:  http.StatusForbidden,
			wantCode:    codeWebSocketRoomForbidden,
			wantMessage: "websocket room access denied",
		},
		{
			name: "join denied",
			cfg: func(t *testing.T) ServerConfig {
				cfg := defaultHandshakeConfig(t)
				cfg.Hub.Stop()
				return cfg
			},
			req:         newValidHandshakeRequest,
			wantStatus:  http.StatusServiceUnavailable,
			wantCode:    codeWebSocketJoinDenied,
			wantMessage: "websocket room join denied",
		},
		{
			name: "invalid token before join denied",
			cfg: func(t *testing.T) ServerConfig {
				cfg := defaultHandshakeConfig(t)
				cfg.Hub.Stop()
				cfg.AllowUnauthenticated = false
				cfg.TokenAuth = failingTokenAuth{}
				return cfg
			},
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.Header.Set("Authorization", "Bearer bad-token")
				return r
			},
			wantStatus:  http.StatusForbidden,
			wantCode:    codeWebSocketInvalidToken,
			wantMessage: "invalid websocket token",
		},
		{
			name: "bearer token with nil token auth",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.Header.Set("Authorization", "Bearer bad-token")
				return r
			},
			wantStatus:  http.StatusForbidden,
			wantCode:    codeWebSocketInvalidToken,
			wantMessage: "invalid websocket token",
		},
		{
			name: "invalid token",
			cfg:  defaultHandshakeConfig,
			req: func() *http.Request {
				r := newValidHandshakeRequest()
				r.URL.RawQuery = "token=not-a-token"
				return r
			},
			wantStatus:  http.StatusForbidden,
			wantCode:    codeWebSocketInvalidToken,
			wantMessage: "invalid websocket token",
		},
		{
			name:        "hijack unsupported",
			cfg:         defaultHandshakeConfig,
			req:         newValidHandshakeRequest,
			wantStatus:  http.StatusInternalServerError,
			wantCode:    codeWebSocketHijackUnsupported,
			wantMessage: "websocket hijack unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg(t)
			if cfg.Hub != nil {
				defer cfg.Hub.Stop()
			}

			w := httptest.NewRecorder()
			ServeWSWithConfig(w, tt.req(), cfg)

			assertWebSocketError(t, w, tt.wantStatus, tt.wantCode, tt.wantMessage)
		})
	}
}

func defaultHandshakeConfig(t *testing.T) ServerConfig {
	t.Helper()
	return ServerConfig{
		Hub:                  mustHub(t, 1, 4),
		RoomAuth:             NewSimpleRoomAuth(),
		AllowUnauthenticated: true,
		SendBehavior:         SendBlock,
		AllowedOrigins:       []string{"*"},
		OnMessage:            noopMessageHandler,
	}
}

func noopMessageHandler(*Conn, Message) {}

func newValidHandshakeRequest() *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/ws", nil)
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Sec-WebSocket-Key", validTestWSKey)
	r.Header.Set("Sec-WebSocket-Version", "13")
	return r
}

func assertWebSocketError(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantCode, wantMessage string) {
	t.Helper()
	if w.Code != wantStatus {
		t.Fatalf("status = %d, want %d", w.Code, wantStatus)
	}

	var resp contract.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != wantCode {
		t.Fatalf("error code = %q, want %q", resp.Error.Code, wantCode)
	}
	if resp.Error.Message != wantMessage {
		t.Fatalf("error message = %q, want %q", resp.Error.Message, wantMessage)
	}
}

func TestServeWSWithConfig_InvalidConfig(t *testing.T) {
	t.Run("nil hub", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/ws", nil)

		ServeWSWithConfig(w, r, ServerConfig{
			RoomAuth:             NewSimpleRoomAuth(),
			AllowUnauthenticated: true,
			OnMessage:            noopMessageHandler,
		})

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("nil auth", func(t *testing.T) {
		hub := mustHub(t, 1, 4)
		defer hub.Stop()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/ws", nil)

		ServeWSWithConfig(w, r, ServerConfig{
			Hub:                  hub,
			AllowUnauthenticated: true,
			OnMessage:            noopMessageHandler,
		})

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("negative queue size", func(t *testing.T) {
		hub := mustHub(t, 1, 4)
		defer hub.Stop()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/ws", nil)

		ServeWSWithConfig(w, r, ServerConfig{
			Hub:                  hub,
			RoomAuth:             NewSimpleRoomAuth(),
			AllowUnauthenticated: true,
			OnMessage:            noopMessageHandler,
			QueueSize:            -1,
		})

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestServeWSWithConfig_CapacityRejectionMetrics(t *testing.T) {
	cfg := defaultHandshakeConfig(t)
	cfg.Hub.Stop()

	w := httptest.NewRecorder()
	ServeWSWithConfig(w, newValidHandshakeRequest(), cfg)

	assertWebSocketError(t, w, http.StatusServiceUnavailable, codeWebSocketJoinDenied, "websocket room join denied")
	if got := cfg.Hub.Metrics().RejectedTotal; got != 1 {
		t.Fatalf("RejectedTotal = %d, want 1", got)
	}
}

func TestResolveValidationConfig(t *testing.T) {
	cfg := ServerConfig{
		ReadLimit: 1024,
		OnMessage: noopMessageHandler,
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

	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	cfg, err := normalizeServerConfig(ServerConfig{
		Hub:                  hub,
		RoomAuth:             auth,
		TokenAuth:            auth,
		AllowUnauthenticated: true,
		OnMessage:            noopMessageHandler,
		SendBehavior:         SendBlock,
	})
	if err != nil {
		t.Fatalf("normalizeServerConfig error: %v", err)
	}
	if cfg.ReadLimit != 2048 {
		t.Fatalf("expected read limit from auth (2048), got %d", cfg.ReadLimit)
	}
}

func TestServeRoomFanoutWSRejectsConflictingHandler(t *testing.T) {
	hub := mustHub(t, 1, 4)
	defer hub.Stop()

	w := httptest.NewRecorder()
	r := newValidHandshakeRequest()

	ServeRoomFanoutWS(w, r, ServerConfig{
		Hub:                  hub,
		RoomAuth:             NewSimpleRoomAuth(),
		AllowUnauthenticated: true,
		AllowedOrigins:       []string{"*"},
		OnMessage:            noopMessageHandler,
	})

	assertWebSocketError(t, w, http.StatusInternalServerError, codeWebSocketInvalidConfig, "websocket server misconfigured")
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

func TestNewSecureRoomAuthClonesSecret(t *testing.T) {
	secret := bytes.Repeat([]byte("a"), 32)
	auth, err := NewSecureRoomAuth(secret, SecurityConfig{
		JWTSecret:          secret,
		MinJWTSecretLength: 32,
	})
	if err != nil {
		t.Fatalf("NewSecureRoomAuth error: %v", err)
	}

	secret[0] = 'b'
	if auth.securityConfig.JWTSecret[0] != 'a' {
		t.Fatal("security config retained caller secret slice")
	}
	if auth.tokenAuth.secret[0] != 'a' {
		t.Fatal("token authenticator retained caller secret slice")
	}
}
