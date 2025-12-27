package core

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

type stubComponent struct {
	path           string
	middlewareName string
	started        bool
	stopped        bool
	startErr       error
}

func (s *stubComponent) RegisterRoutes(r *router.Router) {
	if s.path == "" {
		s.path = "/stub"
	}

	r.GetFunc(s.path, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func (s *stubComponent) RegisterMiddleware(reg *middleware.Registry) {
	if s.middlewareName == "" {
		return
	}

	reg.Use(func(next middleware.Handler) middleware.Handler {
		return middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Component", s.middlewareName)
			next.ServeHTTP(w, r)
		}))
	})
}

func (s *stubComponent) Start(_ context.Context) error {
	s.started = true
	return s.startErr
}

func (s *stubComponent) Stop(_ context.Context) error {
	s.stopped = true
	return nil
}

func (s *stubComponent) Health() (string, health.HealthStatus) {
	return "stub", health.HealthStatus{Status: health.StatusHealthy, Details: map[string]any{"started": s.started, "stopped": s.stopped}}
}

func TestNewDefaults(t *testing.T) {
	app := New()

	if app.config.Addr != ":8080" {
		t.Fatalf("default addr should be :8080, got %s", app.config.Addr)
	}
	if app.config.EnvFile != ".env" {
		t.Fatalf("default env file should be .env, got %s", app.config.EnvFile)
	}
	if app.config.ShutdownTimeout != 5*time.Second {
		t.Fatalf("default shutdown timeout should be 5s, got %v", app.config.ShutdownTimeout)
	}
	if app.config.TLS.Enabled {
		t.Fatalf("TLS should be disabled by default")
	}
}

func TestOptionsApply(t *testing.T) {
	customRouter := router.NewRouter()

	app := New(
		WithRouter(customRouter),
		WithAddr(":9090"),
		WithEnvPath(".custom.env"),
		WithShutdownTimeout(2*time.Second),
		WithTLS("cert", "key"),
		WithDebug(),
	)

	if app.router != customRouter {
		t.Fatalf("custom router should be set")
	}
	if app.config.Addr != ":9090" {
		t.Fatalf("addr should be :9090, got %s", app.config.Addr)
	}
	if app.config.EnvFile != ".custom.env" {
		t.Fatalf("env file should be .custom.env, got %s", app.config.EnvFile)
	}
	if app.config.ShutdownTimeout != 2*time.Second {
		t.Fatalf("shutdown timeout should be 2s, got %v", app.config.ShutdownTimeout)
	}
	if !app.config.TLS.Enabled || app.config.TLS.CertFile != "cert" || app.config.TLS.KeyFile != "key" {
		t.Fatalf("TLS config should be populated when enabled")
	}
	if !app.config.Debug {
		t.Fatalf("debug flag should be true when WithDebug is used")
	}
}

func TestUseMiddlewareAppliedAfterSetup(t *testing.T) {
	app := New()

	app.Get("/router", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("router"))
	})
	app.HandleFunc("/mux", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mux"))
	})

	app.Use(func(next middleware.Handler) middleware.Handler {
		return middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "applied")
			next.ServeHTTP(w, r)
		}))
	})

	if err := app.setupServer(); err != nil {
		t.Fatalf("setupServer returned error: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "app_env")
	if err != nil {
		t.Fatalf("failed to create temp env file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		path     string
		expected string
	}{
		{path: "/router", expected: "router"},
		{path: "/mux", expected: "mux"},
	}
	tmpFile.Close()

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rr := httptest.NewRecorder()

		app.handler.ServeHTTP(rr, req)

		if rr.Header().Get("X-Test") != "applied" {
			t.Fatalf("middleware header missing for path %s", tt.path)
		}
		if !strings.Contains(rr.Body.String(), tt.expected) {
			t.Fatalf("expected body to contain %q for path %s, got %q", tt.expected, tt.path, rr.Body.String())
		}
	}
}

func TestLoadEnvFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "app_env")
	if err != nil {
		t.Fatalf("failed to create temp env file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("APP_TEST_KEY=123\n"); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}
	tmpFile.Close()

	app := New(WithEnvPath(tmpFile.Name()))
	if err := app.loadEnv(); err != nil {
		t.Fatalf("loadEnv returned error: %v", err)
	}
	if val := os.Getenv("APP_TEST_KEY"); val != "123" {
		t.Fatalf("expected APP_TEST_KEY to be set to 123, got %s", val)
	}
}

func TestSetupServerBuildsHTTPServer(t *testing.T) {
	app := New(WithAddr(":5555"))

	// add middleware to ensure chain builds without panic
	app.Use(func(next middleware.Handler) middleware.Handler {
		return next
	})

	if err := app.setupServer(); err != nil {
		t.Fatalf("setupServer returned error: %v", err)
	}

	if app.httpServer == nil {
		t.Fatalf("httpServer should be created during setupServer")
	}
	if app.httpServer.Addr != ":5555" {
		t.Fatalf("httpServer addr should be :5555, got %s", app.httpServer.Addr)
	}
	if app.httpServer.Handler == nil {
		t.Fatalf("httpServer handler should not be nil")
	}
}

func TestUseAfterStartPanics(t *testing.T) {
	app := New()
	app.started = true

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when adding middleware after start")
		}
	}()

	app.Use(func(next middleware.Handler) middleware.Handler { return next })
}

func TestConfigureWebSocketRequiresSecret(t *testing.T) {
	app := New()
	if _, err := app.ConfigureWebSocketWithOptions(WebSocketConfig{}); err == nil {
		t.Fatalf("expected error when secret is missing")
	}
}

func TestConfigureWebSocketLoadsEnvFile(t *testing.T) {
	os.Unsetenv("WS_SECRET")
	defer os.Unsetenv("WS_SECRET")

	tmpFile, err := os.CreateTemp("", "app_env")
	if err != nil {
		t.Fatalf("failed to create temp env file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("WS_SECRET=from_env\n"); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}
	tmpFile.Close()

	app := New(WithEnvPath(tmpFile.Name()))
	if _, err := app.ConfigureWebSocket(); err != nil {
		t.Fatalf("expected websocket configuration to load env, got error: %v", err)
	}
}

func TestComponentRegistrationAndMiddleware(t *testing.T) {
	app := New(WithComponent(&stubComponent{path: "/component", middlewareName: "stub"}))

	comps := app.mountComponents()
	if len(comps) == 0 {
		t.Fatalf("expected mounted components")
	}

	if err := app.setupServer(); err != nil {
		t.Fatalf("setupServer returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/component", nil)
	resp := httptest.NewRecorder()

	app.handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	if resp.Header().Get("X-Component") != "stub" {
		t.Fatalf("expected middleware header from component")
	}
}

func TestStartComponentsStopsOnError(t *testing.T) {
	good := &stubComponent{}
	bad := &stubComponent{startErr: fmt.Errorf("boom")}

	app := New()
	err := app.startComponents(context.Background(), []Component{good, bad})
	if err == nil {
		t.Fatalf("expected error from component start")
	}

	if !good.started || !good.stopped {
		t.Fatalf("expected successful component to be stopped after failure")
	}
}

func TestBroadcastAuthAndToggle(t *testing.T) {
	secret := []byte("super-secret")

	// Non-debug mode requires authorization
	config := DefaultWebSocketConfig()
	config.Secret = secret
	config.BroadcastPath = "/broadcast"
	config.BroadcastEnabled = true

	app := New()
	if _, err := app.ConfigureWebSocketWithOptions(config); err != nil {
		t.Fatalf("configure websocket failed: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/broadcast", strings.NewReader("hello"))
	app.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized without secret, got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/broadcast", strings.NewReader("hello"))
	req.Header.Set("Authorization", "Bearer "+string(secret))
	app.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected success when authorized, got %d", rr.Code)
	}

	// Debug mode bypasses broadcast auth
	debugApp := New(WithDebug())
	debugConfig := config
	if _, err := debugApp.ConfigureWebSocketWithOptions(debugConfig); err != nil {
		t.Fatalf("configure websocket failed: %v", err)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/broadcast", strings.NewReader("hello"))
	debugApp.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected success in debug without auth, got %d", rr.Code)
	}

	// Disable broadcast endpoint
	disabled := New()
	disabledConfig := config
	disabledConfig.BroadcastEnabled = false
	if _, err := disabled.ConfigureWebSocketWithOptions(disabledConfig); err != nil {
		t.Fatalf("configure websocket failed: %v", err)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, config.BroadcastPath, strings.NewReader("hello"))
	disabled.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected not found when broadcast disabled, got %d", rr.Code)
	}
}
