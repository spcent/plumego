package app

import (
	"net/http"
	"testing"

	"with-tenant/internal/config"
)

func TestRegisterRoutesCanonicalShape(t *testing.T) {
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

	routes := a.Core.Routes()
	want := []struct {
		Method string
		Path   string
	}{
		{http.MethodGet, "/api/models"},
	}

	if len(routes) != len(want) {
		t.Fatalf("routes count = %d, want %d; routes = %v", len(routes), len(want), routes)
	}
	if routes[0].Method != http.MethodGet || routes[0].Path != "/api/models" {
		t.Errorf("route = %s %s, want GET /api/models", routes[0].Method, routes[0].Path)
	}
}
