package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/health"
	ws "github.com/spcent/plumego/net/websocket"
	"github.com/spcent/plumego/router"
)

// validSecret returns a secret that meets the minimum length requirement.
func validSecret() []byte {
	return []byte("this-is-a-secret-key-that-is-at-least-32-bytes-long!!")
}

func TestDefaultWebSocketConfig(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	if cfg.WorkerCount != 16 {
		t.Fatalf("expected WorkerCount 16, got %d", cfg.WorkerCount)
	}
	if cfg.JobQueueSize != 4096 {
		t.Fatalf("expected JobQueueSize 4096, got %d", cfg.JobQueueSize)
	}
	if cfg.SendQueueSize != DefaultSendQueueSize {
		t.Fatalf("expected SendQueueSize %d, got %d", DefaultSendQueueSize, cfg.SendQueueSize)
	}
	if cfg.SendTimeout != 200*time.Millisecond {
		t.Fatalf("expected SendTimeout 200ms, got %v", cfg.SendTimeout)
	}
	if cfg.SendBehavior != ws.SendBlock {
		t.Fatalf("expected SendBehavior SendBlock, got %v", cfg.SendBehavior)
	}
	if cfg.WSRoutePath != "/ws" {
		t.Fatalf("expected WSRoutePath /ws, got %q", cfg.WSRoutePath)
	}
	if cfg.BroadcastPath != "/_admin/broadcast" {
		t.Fatalf("expected BroadcastPath /_admin/broadcast, got %q", cfg.BroadcastPath)
	}
	if !cfg.BroadcastEnabled {
		t.Fatal("expected BroadcastEnabled true")
	}
}

func TestNewComponentSecretTooShort(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = []byte("short")

	_, err := NewComponent(cfg, false, nil)
	if err == nil {
		t.Fatal("expected error for short secret")
	}
	if !strings.Contains(err.Error(), "at least") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewComponentEmptySecret(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = nil

	_, err := NewComponent(cfg, false, nil)
	if err == nil {
		t.Fatal("expected error for nil secret")
	}
}

func TestNewComponentValidSecret(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp == nil {
		t.Fatal("expected non-nil component")
	}
	if comp.Hub() == nil {
		t.Fatal("expected non-nil hub")
	}
}

func TestHealthHealthy(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	name, status := comp.Health()
	if name != "websocket" {
		t.Fatalf("expected name websocket, got %q", name)
	}
	if status.Status != health.StatusHealthy {
		t.Fatalf("expected healthy, got %s", status.Status)
	}
	if status.Details == nil {
		t.Fatal("expected non-nil details")
	}
	if val, ok := status.Details["broadcastEnabled"]; !ok || val != true {
		t.Fatalf("expected broadcastEnabled=true in details, got %v", status.Details)
	}
}

func TestHealthUnhealthyAfterStop(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := comp.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}

	_, status := comp.Health()
	if status.Status != health.StatusUnhealthy {
		t.Fatalf("expected unhealthy after stop, got %s", status.Status)
	}
	if status.Message == "" {
		t.Fatal("expected non-empty message")
	}
}

func TestStartReturnsNil(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, _ := NewComponent(cfg, false, nil)
	if err := comp.Start(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestStopNilHub(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, _ := NewComponent(cfg, false, nil)
	// Stop once to set hub to nil
	_ = comp.Stop(context.Background())
	// Stop again should not panic
	if err := comp.Stop(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDependencies(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, _ := NewComponent(cfg, false, nil)
	if deps := comp.Dependencies(); deps != nil {
		t.Fatalf("expected nil dependencies, got %v", deps)
	}
}

func TestRegisterMiddlewareNoop(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, _ := NewComponent(cfg, false, nil)
	comp.RegisterMiddleware(nil) // should not panic
}

func TestBroadcastEndpointNoAuth(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader(`{"msg":"hi"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBroadcastEndpointWrongToken(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader(`{"msg":"hi"}`))
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBroadcastEndpointValidToken(t *testing.T) {
	secret := validSecret()
	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader(`{"msg":"hi"}`))
	req.Header.Set("Authorization", "Bearer "+string(secret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		body := rec.Body.String()
		t.Fatalf("expected 204, got %d; body: %s", rec.Code, body)
	}
}

func TestBroadcastDisabled(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.BroadcastEnabled = false

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+string(cfg.Secret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should get 404 or 405 since route is not registered
	if rec.Code == http.StatusNoContent {
		t.Fatal("broadcast should be disabled")
	}
}

func TestRegisterRoutesIdempotent(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	// Calling RegisterRoutes twice should not panic (sync.Once)
	comp.RegisterRoutes(r)
	comp.RegisterRoutes(r)
}

func TestHealthBroadcastEnabledInDetails(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.BroadcastEnabled = false

	comp, _ := NewComponent(cfg, false, nil)
	_, status := comp.Health()

	val, ok := status.Details["broadcastEnabled"]
	if !ok {
		t.Fatal("expected broadcastEnabled in details")
	}
	if val != false {
		t.Fatal("expected broadcastEnabled=false")
	}
}

func TestBroadcastEndpointEmptyBody(t *testing.T) {
	secret := validSecret()
	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+string(secret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for empty body, got %d", rec.Code)
	}
}

func TestBroadcastAuthCaseInsensitive(t *testing.T) {
	secret := validSecret()
	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret

	comp, err := NewComponent(cfg, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	// lowercase "bearer" should also work
	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader("test"))
	req.Header.Set("Authorization", "bearer "+string(secret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		var body map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&body)
		t.Fatalf("expected 204, got %d; body: %v", rec.Code, body)
	}
}

func TestNewComponentCustomConfig(t *testing.T) {
	cfg := WebSocketConfig{
		WorkerCount:        4,
		JobQueueSize:       128,
		SendQueueSize:      64,
		SendTimeout:        100 * time.Millisecond,
		SendBehavior:       ws.SendDrop,
		Secret:             validSecret(),
		WSRoutePath:        "/custom-ws",
		BroadcastPath:      "/custom-broadcast",
		BroadcastEnabled:   true,
		MaxConnections:     100,
		MaxRoomConnections: 10,
	}

	comp, err := NewComponent(cfg, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if comp.Hub() == nil {
		t.Fatal("expected hub")
	}
}
