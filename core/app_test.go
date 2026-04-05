package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/router"
)

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

	mustRegisterRoute(t, app.Get("/router", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("router"))
	}))
	mustRegisterRoute(t, app.Get("/mux", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mux"))
	}))

	app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "applied")
			next.ServeHTTP(w, r)
		})
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

func TestServeHTTPLazilyBuildsHandler(t *testing.T) {
	app := New()

	app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Lazy", "true")
			next.ServeHTTP(w, r)
		})
	})

	mustRegisterRoute(t, app.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hi"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rr := httptest.NewRecorder()

	app.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Lazy") != "true" {
		t.Fatalf("expected middleware header to be set")
	}
}

func TestUseAfterServeHTTPReturnsError(t *testing.T) {
	app := New()

	mustRegisterRoute(t, app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	if err := app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}); err == nil {
		t.Fatalf("expected error when adding middleware after handler is built")
	}
}

func TestPrepareBuildsHTTPServer(t *testing.T) {
	app := New(WithAddr(":5555"))
	mustRegisterRoute(t, app.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	server, err := app.Server()
	if err != nil {
		t.Fatalf("Server returned error: %v", err)
	}
	if server.Addr != ":5555" {
		t.Fatalf("httpServer addr should be :5555, got %s", server.Addr)
	}
	if server.Handler == nil {
		t.Fatalf("httpServer handler should not be nil")
	}
}

func TestSetupServerBuildsHTTPServer(t *testing.T) {
	app := New(WithAddr(":5555"))

	// add middleware to ensure chain builds without panic
	app.Use(func(next http.Handler) http.Handler {
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

	err := app.Use(func(next http.Handler) http.Handler { return next })
	if err == nil {
		t.Fatalf("expected error when adding middleware after start")
	}
}

type funcRunner struct {
	start func(context.Context) error
	stop  func(context.Context) error
}

func (f funcRunner) Start(ctx context.Context) error { return f.start(ctx) }
func (f funcRunner) Stop(ctx context.Context) error  { return f.stop(ctx) }
