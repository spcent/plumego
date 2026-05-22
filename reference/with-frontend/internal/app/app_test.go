package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"with-frontend/internal/config"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	return a
}

func TestRegisterRoutesIncludesAPIAndFrontend(t *testing.T) {
	a := newTestApp(t)

	routes := a.Core.Routes()
	if len(routes) == 0 {
		t.Fatal("expected at least one route")
	}

	hasAPI := false
	for _, r := range routes {
		if r.Method == http.MethodGet && r.Path == "/api/status" {
			hasAPI = true
		}
	}
	if !hasAPI {
		t.Errorf("route GET /api/status not found; routes = %v", routes)
	}
}

func TestAPIStatusEndpoint(t *testing.T) {
	a := newTestApp(t)
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
}

func TestFrontendIndexServed(t *testing.T) {
	a := newTestApp(t)
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); body == "" {
		t.Error("expected non-empty HTML response for /")
	}
}
