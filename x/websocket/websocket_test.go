package websocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/router"
)

// validSecret returns a secret that meets the minimum length requirement.
func validSecret() []byte {
	return []byte("this-is-a-secret-key-that-is-at-least-32-bytes-long!!")
}

func mustHub(t *testing.T, workerCount, jobQueueSize int) *Hub {
	t.Helper()
	hub, err := NewHubE(workerCount, jobQueueSize)
	if err != nil {
		t.Fatalf("NewHubE: %v", err)
	}
	return hub
}

func mustHubWithConfig(t *testing.T, cfg HubConfig) *Hub {
	t.Helper()
	hub, err := NewHubWithConfigE(cfg)
	if err != nil {
		t.Fatalf("NewHubWithConfigE: %v", err)
	}
	return hub
}

func mustJoin(t *testing.T, hub *Hub, room string, conn *Conn) {
	t.Helper()
	if err := hub.TryJoin(room, conn); err != nil {
		t.Fatalf("TryJoin(%q): %v", room, err)
	}
}

type websocketErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type countingRouteRegistrar struct {
	count int
}

func (r *countingRouteRegistrar) AddRoute(method, path string, handler http.Handler, opts ...router.RouteOption) error {
	r.count++
	return nil
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
	if cfg.SendBehavior != SendBlock {
		t.Fatalf("expected SendBehavior SendBlock, got %v", cfg.SendBehavior)
	}
	if cfg.WSRoutePath != "/ws" {
		t.Fatalf("expected WSRoutePath /ws, got %q", cfg.WSRoutePath)
	}
	if cfg.BroadcastPath != "/_admin/broadcast" {
		t.Fatalf("expected BroadcastPath /_admin/broadcast, got %q", cfg.BroadcastPath)
	}
	if cfg.BroadcastEnabled {
		t.Fatal("expected BroadcastEnabled false")
	}
	if cfg.BroadcastMaxBodyBytes != defaultBroadcastMaxBodyBytes {
		t.Fatalf("expected BroadcastMaxBodyBytes %d, got %d", defaultBroadcastMaxBodyBytes, cfg.BroadcastMaxBodyBytes)
	}
	if len(cfg.Secret) != 0 {
		t.Fatal("expected DefaultWebSocketConfig to leave Secret empty")
	}
}

func TestNewRejectsShortSecret(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = []byte("short")

	_, err := New(cfg)
	if !errors.Is(err, ErrWeakJWTSecret) {
		t.Fatalf("New error = %v, want ErrWeakJWTSecret", err)
	}
}

func TestNewEmptySecret(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = nil

	_, err := New(cfg)
	assertErrorIsOrContains(t, err, ErrNilTokenAuthorizer, "token authenticator")
}

func TestNewAllowsAnonymousWithoutSecret(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = nil
	cfg.AllowUnauthenticated = true

	comp, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp == nil {
		t.Fatal("expected component")
	}
}

func TestNewValidSecret(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := New(cfg)
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

func TestNewRejectsEmptyRoutePath(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.WSRoutePath = ""

	comp, err := New(cfg)
	if !errors.Is(err, ErrEmptyRoutePath) {
		t.Fatalf("New error = %v, want ErrEmptyRoutePath", err)
	}
	if comp != nil {
		t.Fatal("New returned component for invalid route config")
	}
}

func TestNewRejectsEmptyBroadcastPath(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = []byte("this-is-a-broadcast-token-at-least-32-bytes-long")
	cfg.BroadcastPath = ""

	comp, err := New(cfg)
	if !errors.Is(err, ErrEmptyRoutePath) {
		t.Fatalf("New error = %v, want ErrEmptyRoutePath", err)
	}
	if comp != nil {
		t.Fatal("New returned component for invalid broadcast route config")
	}
}

func TestNewClonesSecretAndAllowedOrigins(t *testing.T) {
	secret := validSecret()
	origSecret := append([]byte(nil), secret...)
	origins := []string{"https://app.example.com"}

	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret
	cfg.AllowedOrigins = origins

	comp, err := New(cfg)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	secret[0] ^= 0xff
	origins[0] = "https://mutated.example.com"

	if !bytes.Equal(comp.config.Secret, origSecret) {
		t.Fatal("server retained caller-owned Secret slice")
	}
	if got := comp.config.AllowedOrigins[0]; got != "https://app.example.com" {
		t.Fatalf("AllowedOrigins[0] = %q, want original origin", got)
	}
}

func TestBroadcastEndpointUsesClonedBroadcastSecret(t *testing.T) {
	secret := validSecret()
	broadcastSecret := []byte("this-is-a-broadcast-token-at-least-32-bytes-long")
	originalBroadcastSecret := string(append([]byte(nil), broadcastSecret...))

	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = broadcastSecret

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	broadcastSecret[0] ^= 0xff

	r := router.NewRouter()
	if err := comp.RegisterRoutes(r); err != nil {
		t.Fatalf("RegisterRoutes error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader("hello"))
	req.Header.Set("Authorization", "Bearer "+originalBroadcastSecret)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected cloned broadcast secret to authorize, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestNewPropagatesHubConfig(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)
	events := make(chan SecurityEvent, 1)

	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.EnableDebugLogging = true
	cfg.Logger = logger
	cfg.RejectOnQueueFull = true
	cfg.MaxConnectionRate = 10
	cfg.EnableSecurityEvents = true
	cfg.SecurityEventHandler = func(event SecurityEvent) {
		events <- event
	}

	comp, err := New(cfg)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer comp.Hub().Stop()

	hubCfg := comp.Hub().config
	if !hubCfg.EnableDebugLogging || hubCfg.Logger != logger || !hubCfg.RejectOnQueueFull || hubCfg.MaxConnectionRate != 10 || !hubCfg.EnableSecurityEvents {
		t.Fatalf("hub config was not propagated: %#v", hubCfg)
	}
	if hubCfg.SecurityEventHandler == nil {
		t.Fatal("expected SecurityEventHandler to be propagated")
	}
	if comp.Hub().rateLimiter == nil {
		t.Fatal("expected MaxConnectionRate to initialize rate limiter")
	}
}

func TestRegisterRoutesUsesTopLevelMessageValidation(t *testing.T) {
	delivered := make(chan Message, 1)

	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.AllowUnauthenticated = true
	cfg.MessageValidation = MessageValidationConfig{
		MaxLength:        4,
		AllowEmpty:       true,
		RequireValidUTF8: true,
	}
	cfg.OnMessage = func(_ *Conn, msg Message) {
		delivered <- msg
	}

	comp, err := New(cfg)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer comp.Shutdown(t.Context())

	r := router.NewRouter()
	if err := comp.RegisterRoutes(r); err != nil {
		t.Fatalf("RegisterRoutes error: %v", err)
	}
	server := httptest.NewServer(r)
	defer server.Close()

	client := newTestWSClient(t, server.URL, "room1", "", "")
	defer client.conn.Close()
	if err := client.sendFrame(OpcodeText, true, []byte("too long")); err != nil {
		t.Fatalf("sendFrame error: %v", err)
	}

	select {
	case msg := <-delivered:
		t.Fatalf("message validation did not drop oversized text message: %#v", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestNewRejectsNegativeReadLimit(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.ReadLimit = -1

	if _, err := New(cfg); !errors.Is(err, ErrNegativeReadLimit) {
		t.Fatalf("New error = %v, want ErrNegativeReadLimit", err)
	}
}

func TestNewRejectsReadLimitAboveHardCap(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.ReadLimit = maxReadLimit + 1

	if _, err := New(cfg); !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("New error = %v, want ErrPayloadTooLarge", err)
	}
}

func TestHealthHealthy(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := New(cfg)
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
	if val, ok := status.Details["broadcastEnabled"]; !ok || val != false {
		t.Fatalf("expected broadcastEnabled=false in details, got %v", status.Details)
	}
}

func TestHealthUnhealthyAfterStop(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = []byte("this-is-a-broadcast-token-at-least-32-bytes-long")

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := comp.Shutdown(t.Context()); err != nil {
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

func TestStopNilHub(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = []byte("this-is-a-broadcast-token-at-least-32-bytes-long")

	comp, _ := New(cfg)
	// Stop once to set hub to nil
	_ = comp.Shutdown(t.Context())
	// Stop again should not panic
	if err := comp.Shutdown(t.Context()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestBroadcastEndpointNoAuth(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = []byte("this-is-a-broadcast-token-at-least-32-bytes-long")

	comp, err := New(cfg)
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
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = []byte("this-is-a-broadcast-token-at-least-32-bytes-long")

	comp, err := New(cfg)
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
	broadcastSecret := []byte("this-is-a-broadcast-token-at-least-32-bytes-long")
	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = broadcastSecret

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader(`{"msg":"hi"}`))
	req.Header.Set("Authorization", "Bearer "+string(broadcastSecret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		body := rec.Body.String()
		t.Fatalf("expected 204, got %d; body: %s", rec.Code, body)
	}
}

func TestBroadcastEndpointBodyTooLarge(t *testing.T) {
	secret := validSecret()
	broadcastSecret := []byte("this-is-a-broadcast-token-at-least-32-bytes-long")
	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = broadcastSecret
	cfg.BroadcastMaxBodyBytes = 4

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader("12345"))
	req.Header.Set("Authorization", "Bearer "+string(broadcastSecret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		var body websocketErrorResponse
		_ = json.NewDecoder(rec.Body).Decode(&body)
		t.Fatalf("expected 413, got %d; body: %v", rec.Code, body)
	}
}

func TestBroadcastEndpointInvalidRoomName(t *testing.T) {
	secret := validSecret()
	broadcastSecret := []byte("this-is-a-broadcast-token-at-least-32-bytes-long")
	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = broadcastSecret

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast?room=bad/room", strings.NewReader("hello"))
	req.Header.Set("Authorization", "Bearer "+string(broadcastSecret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		var body websocketErrorResponse
		_ = json.NewDecoder(rec.Body).Decode(&body)
		t.Fatalf("expected 400, got %d; body: %v", rec.Code, body)
	}
}

func TestBroadcastDisabled(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer this-is-a-broadcast-token-at-least-32-bytes-long")
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

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	if err := comp.RegisterRoutes(r); err != nil {
		t.Fatalf("first RegisterRoutes failed: %v", err)
	}
	if err := comp.RegisterRoutes(r); err == nil {
		t.Fatal("expected duplicate RegisterRoutes to return an error")
	}
}

func TestRegisterRoutesRejectsNilRegistrar(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := comp.RegisterRoutes(nil); !errors.Is(err, ErrNilRegistrar) {
		t.Fatalf("expected ErrNilRegistrar, got %v", err)
	}
}

func TestRegisterRoutesRejectsEmptyRoutePathBeforeAddRoute(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	comp.config.WSRoutePath = ""
	r := &countingRouteRegistrar{}
	if err := comp.RegisterRoutes(r); !errors.Is(err, ErrEmptyRoutePath) {
		t.Fatalf("expected ErrEmptyRoutePath, got %v", err)
	}
	if r.count != 0 {
		t.Fatalf("AddRoute called %d times, want 0", r.count)
	}
}

func TestRegisterRoutesRejectsInvalidBroadcastConfigBeforeAddRoute(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	comp.config.BroadcastEnabled = true
	comp.config.BroadcastPath = ""
	r := &countingRouteRegistrar{}
	if err := comp.RegisterRoutes(r); !errors.Is(err, ErrEmptyRoutePath) {
		t.Fatalf("expected ErrEmptyRoutePath, got %v", err)
	}
	if r.count != 0 {
		t.Fatalf("AddRoute called %d times, want 0", r.count)
	}
}

func TestHealthBroadcastEnabledInDetails(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.Secret = validSecret()
	cfg.BroadcastEnabled = false

	comp, _ := New(cfg)
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
	broadcastSecret := []byte("this-is-a-broadcast-token-at-least-32-bytes-long")
	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = broadcastSecret

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+string(broadcastSecret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for empty body, got %d", rec.Code)
	}
}

func TestBroadcastAuthCaseInsensitive(t *testing.T) {
	secret := validSecret()
	broadcastSecret := []byte("this-is-a-broadcast-token-at-least-32-bytes-long")
	cfg := DefaultWebSocketConfig()
	cfg.Secret = secret
	cfg.BroadcastEnabled = true
	cfg.BroadcastSecret = broadcastSecret

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	comp.RegisterRoutes(r)

	// lowercase "bearer" should also work
	req := httptest.NewRequest(http.MethodPost, "/_admin/broadcast", strings.NewReader("test"))
	req.Header.Set("Authorization", "bearer "+string(broadcastSecret))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		var body websocketErrorResponse
		_ = json.NewDecoder(rec.Body).Decode(&body)
		t.Fatalf("expected 204, got %d; body: %v", rec.Code, body)
	}
}

func TestNewCustomConfig(t *testing.T) {
	cfg := WebSocketConfig{
		WorkerCount:          4,
		JobQueueSize:         128,
		SendQueueSize:        64,
		SendTimeout:          100 * time.Millisecond,
		SendBehavior:         SendDrop,
		Secret:               validSecret(),
		AllowUnauthenticated: true,
		AllowedOrigins:       []string{"https://app.example.com"},
		WSRoutePath:          "/custom-ws",
		BroadcastPath:        "/custom-broadcast",
		BroadcastEnabled:     true,
		BroadcastSecret:      []byte("this-is-a-broadcast-token-at-least-32-bytes-long"),
		MaxRoomRegistrations: 100,
		MaxRoomConnections:   10,
	}

	comp, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if comp.Hub() == nil {
		t.Fatal("expected hub")
	}
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("error = nil, want mention of %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want mention of %q", err, want)
	}
}

func assertErrorIsOrContains(t *testing.T, err error, target error, wants ...string) {
	t.Helper()

	if err == nil {
		t.Fatalf("error = nil, want %v or mention of %v", target, wants)
	}
	if errors.Is(err, target) {
		return
	}
	for _, want := range wants {
		if strings.Contains(err.Error(), want) {
			return
		}
	}
	t.Fatalf("error = %v, want %v or mention of %v", err, target, wants)
}
