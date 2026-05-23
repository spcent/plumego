package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"with-webhook/internal/config"
)

func TestWebhookSmokeHealthz(t *testing.T) {
	cfg := config.Defaults()
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	a.Core.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("/healthz status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
}

func TestWebhookRegisterRoutesIncludesHealthz(t *testing.T) {
	cfg := config.Defaults()
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	found := false
	for _, r := range a.Core.Routes() {
		if r.Method == http.MethodGet && r.Path == "/healthz" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("route GET /healthz not registered; routes = %v", a.Core.Routes())
	}
}
