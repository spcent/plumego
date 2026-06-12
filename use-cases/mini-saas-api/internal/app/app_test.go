package app

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRegisterRoutesShape(t *testing.T) {
	a, err := New(testConfig(t))
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	got := a.Core.Routes()
	// Sorted by Method then Path.
	want := [][2]string{
		{http.MethodDelete, "/api/v1/projects/:id"},
		{http.MethodDelete, "/api/v1/tenant/members/:id"},
		{http.MethodGet, "/api/v1/me"},
		{http.MethodGet, "/api/v1/projects"},
		{http.MethodGet, "/api/v1/projects/:id"},
		{http.MethodGet, "/api/v1/tenant"},
		{http.MethodGet, "/api/v1/tenant/audit"},
		{http.MethodGet, "/api/v1/tenant/members"},
		{http.MethodGet, "/healthz"},
		{http.MethodGet, "/metrics"},
		{http.MethodGet, "/readyz"},
		{http.MethodPatch, "/api/v1/tenant"},
		{http.MethodPatch, "/api/v1/tenant/members/:id"},
		{http.MethodPost, "/api/v1/auth/login"},
		{http.MethodPost, "/api/v1/auth/refresh"},
		{http.MethodPost, "/api/v1/auth/signup"},
		{http.MethodPost, "/api/v1/projects"},
		{http.MethodPost, "/api/v1/tenant/members"},
		{http.MethodPut, "/api/v1/projects/:id"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d routes, want %d", len(got), len(want))
	}
	for i, route := range got {
		if route.Method != want[i][0] || route.Path != want[i][1] {
			t.Fatalf("route %d: got %s %s, want %s %s", i, route.Method, route.Path, want[i][0], want[i][1])
		}
	}
}

func TestMiddlewareSecurityHeaders(t *testing.T) {
	a, err := New(testConfig(t))
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	h := rec.Header()
	for _, name := range []string{"X-Frame-Options", "X-Content-Type-Options", "Referrer-Policy"} {
		if h.Get(name) == "" {
			t.Errorf("security header %s missing", name)
		}
	}
}

func TestMiddlewarePanicRecovery(t *testing.T) {
	a, err := New(testConfig(t))
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Get("/test-panic", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("intentional test panic")
	})); err != nil {
		t.Fatalf("register panic route: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/test-panic", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("panic recovery: status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestMiddlewareBodyLimit(t *testing.T) {
	cfg := testConfig(t)
	cfg.App.MaxBodyBytes = 10
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	// Register a writable route that reads the body so the bodylimit middleware can trigger.
	// bodylimit wraps the body with a limited reader; overruns are detected on read, not on Content-Length.
	if err := a.Core.Post("/test-body", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body) // triggers 413 from middleware when body exceeds limit
		w.WriteHeader(http.StatusOK)
	})); err != nil {
		t.Fatalf("register route: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/test-body", strings.NewReader(strings.Repeat("x", 11))))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("body limit: status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

// TestAcceptanceHealthLiveness verifies GET /healthz returns 200.
func TestAcceptanceHealthLiveness(t *testing.T) {
	a, err := New(testConfig(t))
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestAcceptanceHealthReadiness verifies GET /readyz returns 200 with no failing checkers.
func TestAcceptanceHealthReadiness(t *testing.T) {
	a, err := New(testConfig(t))
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /readyz: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestAcceptanceGracefulShutdown verifies that canceling ctx makes Start return nil.
func TestAcceptanceGracefulShutdown(t *testing.T) {
	cfg := testConfig(t)
	cfg.Core.Addr = ":0"
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Start(ctx) }()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error after shutdown: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return within 5s after context cancellation")
	}
}
