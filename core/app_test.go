package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestNewDefaults(t *testing.T) {
	app := New(DefaultConfig(), AppDependencies{})

	if app.config.Addr != ":8080" {
		t.Fatalf("default addr should be :8080, got %s", app.config.Addr)
	}
	if app.config.TLS.Enabled {
		t.Fatalf("TLS should be disabled by default")
	}
}

func TestNilAppQueryEntrypoints(t *testing.T) {
	var app *App

	if logger := app.Logger(); logger == nil {
		t.Fatal("expected nil app Logger to return discard logger")
	}
	if routes := app.Routes(); routes != nil {
		t.Fatalf("expected nil routes, got %+v", routes)
	}
	if got := app.URL("missing"); got != "" {
		t.Fatalf("expected empty URL from nil app, got %q", got)
	}
}

func TestNilAppRegistrationEntrypointsReturnErrors(t *testing.T) {
	var app *App
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	if err := app.Use(func(next http.Handler) http.Handler { return next }); err == nil || !strings.Contains(err.Error(), "core use_middleware: app is nil") {
		t.Fatalf("expected nil app middleware error, got %v", err)
	}
	if err := app.Get("/nil", handler); err == nil || !strings.Contains(err.Error(), "core add_route") || !strings.Contains(err.Error(), "app is nil") {
		t.Fatalf("expected nil app route error, got %v", err)
	}
	if err := app.AddRoute(http.MethodPost, "/nil", handler); err == nil || !strings.Contains(err.Error(), "core add_route") || !strings.Contains(err.Error(), "app is nil") {
		t.Fatalf("expected nil app add route error, got %v", err)
	}
}

func TestNilAppServeHTTPWritesUnavailable(t *testing.T) {
	var app *App
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nil", nil)

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}
	var response contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Code != contract.CodeUnavailable {
		t.Fatalf("expected unavailable code, got %s", response.Error.Code)
	}
	if response.Error.Message != "app not configured" {
		t.Fatalf("expected app not configured message, got %q", response.Error.Message)
	}
}

func TestZeroValueAppEntrypoints(t *testing.T) {
	var app App
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	if logger := app.Logger(); logger == nil {
		t.Fatal("expected zero-value app Logger to return discard logger")
	}
	if err := app.Use(func(next http.Handler) http.Handler { return next }); err == nil || err.Error() != "core use_middleware: app not initialized" {
		t.Fatalf("expected zero-value app middleware error, got %v", err)
	}
	if err := app.Get("/zero", handler); err == nil || !strings.Contains(err.Error(), "core add_route") || !strings.Contains(err.Error(), "app not initialized") {
		t.Fatalf("expected zero-value app route error, got %v", err)
	}
	if err := app.Prepare(); err == nil || err.Error() != "core prepare_server: app not initialized" {
		t.Fatalf("expected zero-value app prepare error, got %v", err)
	}
	if _, err := app.Server(); err == nil || err.Error() != "core get_server: app not initialized" {
		t.Fatalf("expected zero-value app server error, got %v", err)
	}
	if err := app.Shutdown(nil); err == nil || err.Error() != "core shutdown_app: app not initialized" {
		t.Fatalf("expected zero-value app shutdown error, got %v", err)
	}
}

func TestZeroValueAppServeHTTPWritesUnavailable(t *testing.T) {
	var app App
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/zero", nil)

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}
	var response contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Message != "handler not configured" {
		t.Fatalf("expected handler not configured message, got %q", response.Error.Message)
	}
}

func TestNewAppliesTypedConfigAndOptions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = ":9090"
	cfg.TLS = TLSConfig{Enabled: true, CertFile: "cert", KeyFile: "key"}

	app := New(cfg, AppDependencies{})

	if app.config.Addr != ":9090" {
		t.Fatalf("addr should be :9090, got %s", app.config.Addr)
	}
	if !app.config.TLS.Enabled || app.config.TLS.CertFile != "cert" || app.config.TLS.KeyFile != "key" {
		t.Fatalf("TLS config should be populated when enabled")
	}
}

func TestUseMiddlewareAppliedAfterSetup(t *testing.T) {
	app := newTestApp()

	mustRegisterRoute(t, app.Get("/router", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("router"))
	})))
	mustRegisterRoute(t, app.Get("/mux", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mux"))
	})))

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

	mustRegisterRoute(t, app.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hi"))
	})))

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

	mustRegisterRoute(t, app.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

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

func TestUseRejectsNilMiddlewareWithoutMutatingChain(t *testing.T) {
	app := newTestApp()
	valid := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Valid", "true")
			next.ServeHTTP(w, r)
		})
	}

	if err := app.Use(nil, valid); err == nil {
		t.Fatal("expected nil middleware registration error")
	}
	if got := app.middlewareChain.Len(); got != 0 {
		t.Fatalf("expected chain to stay empty after rejected registration, got %d", got)
	}

	if err := app.Use(valid); err != nil {
		t.Fatalf("expected valid middleware registration to succeed, got %v", err)
	}
	mustRegisterRoute(t, app.Get("/middleware", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/middleware", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if rr.Header().Get("X-Valid") != "true" {
		t.Fatal("expected valid middleware to run after nil registration was rejected")
	}
}

func TestPrepareBuildsHTTPServer(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = ":5555"
	app := New(cfg, AppDependencies{})
	mustRegisterRoute(t, app.Get("/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

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
	if app.preparationState != PreparationStateServerPrepared {
		t.Fatalf("preparation_state = %q, want %q", app.preparationState, PreparationStateServerPrepared)
	}
}

func TestServeHTTPOnlyPreparesHandler(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = ":5555"
	app := New(cfg, AppDependencies{})

	app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Prepared", "true")
			next.ServeHTTP(w, r)
		})
	})
	mustRegisterRoute(t, app.Get("/prepared", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/prepared", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Prepared") != "true" {
		t.Fatalf("expected prepared middleware to run")
	}

	if app.preparationState != PreparationStateHandlerPrepared {
		t.Fatalf("preparation_state = %q, want %q", app.preparationState, PreparationStateHandlerPrepared)
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
	app.preparationState = PreparationStateServerPrepared

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
