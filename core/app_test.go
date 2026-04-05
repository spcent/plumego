package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewDefaults(t *testing.T) {
	app := New(DefaultConfig())

	if app.config.Addr != ":8080" {
		t.Fatalf("default addr should be :8080, got %s", app.config.Addr)
	}
	if app.config.TLS.Enabled {
		t.Fatalf("TLS should be disabled by default")
	}
}

func TestNewAppliesTypedConfigAndOptions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = ":9090"
	cfg.TLS = TLSConfig{Enabled: true, CertFile: "cert", KeyFile: "key"}

	app := New(cfg)

	if app.config.Addr != ":9090" {
		t.Fatalf("addr should be :9090, got %s", app.config.Addr)
	}
	if !app.config.TLS.Enabled || app.config.TLS.CertFile != "cert" || app.config.TLS.KeyFile != "key" {
		t.Fatalf("TLS config should be populated when enabled")
	}
}

func TestUseMiddlewareAppliedAfterSetup(t *testing.T) {
	app := newTestApp()

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

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	tests := []struct {
		path     string
		expected string
	}{
		{path: "/router", expected: "router"},
		{path: "/mux", expected: "mux"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rr := httptest.NewRecorder()

		app.ServeHTTP(rr, req)

		if rr.Header().Get("X-Test") != "applied" {
			t.Fatalf("middleware header missing for path %s", tt.path)
		}
		if !strings.Contains(rr.Body.String(), tt.expected) {
			t.Fatalf("expected body to contain %q for path %s, got %q", tt.expected, tt.path, rr.Body.String())
		}
	}
}

func TestServeHTTPLazilyBuildsHandler(t *testing.T) {
	app := newTestApp()

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
	app := newTestApp()

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
	cfg := DefaultConfig()
	cfg.Addr = ":5555"
	app := New(cfg)
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
	snapshot := app.RuntimeSnapshot()
	if !snapshot.ConfigFrozen {
		t.Fatalf("expected config to be frozen after Prepare")
	}
	if !snapshot.ServerPrepared {
		t.Fatalf("expected server to be prepared after Prepare")
	}
}

func TestServeHTTPOnlyPreparesHandler(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = ":5555"
	app := New(cfg)

	app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Prepared", "true")
			next.ServeHTTP(w, r)
		})
	})
	mustRegisterRoute(t, app.Get("/prepared", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/prepared", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Prepared") != "true" {
		t.Fatalf("expected prepared middleware to run")
	}

	snapshot := app.RuntimeSnapshot()
	if !snapshot.ConfigFrozen {
		t.Fatalf("expected config to be frozen after ServeHTTP")
	}
	if snapshot.ServerPrepared {
		t.Fatalf("expected server to remain unprepared after ServeHTTP")
	}
	if _, err := app.Server(); err == nil {
		t.Fatalf("expected Server to fail before Prepare")
	}

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error after ServeHTTP: %v", err)
	}
	server, err := app.Server()
	if err != nil {
		t.Fatalf("Server returned error after Prepare: %v", err)
	}
	if server.Addr != ":5555" {
		t.Fatalf("httpServer addr should be :5555, got %s", server.Addr)
	}
}

func TestUseAfterStartPanics(t *testing.T) {
	app := newTestApp()
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
