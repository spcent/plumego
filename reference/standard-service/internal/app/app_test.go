package app

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
	"standard-service/internal/config"
)

func TestRegisterRoutesCanonicalShape(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	// router.Routes() returns entries sorted by Method then Path, so the want
	// slice must follow the same order: DELETE before GET before POST,
	// and paths alphabetically within each method.
	got := a.Core.Routes()
	want := []router.RouteInfo{
		{Method: http.MethodDelete, Path: "/api/v1/items/:id"},
		{Method: http.MethodGet, Path: "/"},
		{Method: http.MethodGet, Path: "/api/hello"},
		{Method: http.MethodGet, Path: "/api/status"},
		{Method: http.MethodGet, Path: "/api/v1/greet"},
		{Method: http.MethodGet, Path: "/api/v1/items"},
		{Method: http.MethodGet, Path: "/api/v1/items/:id"},
		{Method: http.MethodGet, Path: "/healthz"},
		{Method: http.MethodGet, Path: "/readyz"},
		{Method: http.MethodPost, Path: "/api/v1/items"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("routes = %#v, want %#v", got, want)
	}
}

func TestNewWiresCanonicalMiddlewareShape(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	req.Header.Set(contract.RequestIDHeader, "req-test-1")
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(contract.RequestIDHeader); got != "req-test-1" {
		t.Fatalf("%s = %q, want %q", contract.RequestIDHeader, got, "req-test-1")
	}
}
