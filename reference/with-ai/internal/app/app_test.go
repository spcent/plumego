package app

import (
	"net/http"
	"testing"

	"with-ai/internal/config"
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
		{http.MethodPost, "/api/chat"},
		{http.MethodGet, "/api/ai/status"},
	}

	if len(routes) != len(want) {
		t.Fatalf("routes count = %d, want %d; routes = %v", len(routes), len(want), routes)
	}

	index := make(map[string]bool, len(routes))
	for _, r := range routes {
		index[r.Method+":"+r.Path] = true
	}
	for _, w := range want {
		if !index[w.Method+":"+w.Path] {
			t.Errorf("route %s %s not found", w.Method, w.Path)
		}
	}
}
