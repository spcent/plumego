package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"with-messaging/internal/config"
)

func TestMessagingSmokeHealthz(t *testing.T) {
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

func TestMessagingRegisterRoutesShape(t *testing.T) {
	cfg := config.Defaults()
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	want := map[string]bool{
		"GET:/healthz":         true,
		"POST:/events/publish": true,
	}
	for _, r := range a.Core.Routes() {
		delete(want, r.Method+":"+r.Path)
	}
	for missing := range want {
		t.Errorf("route %s not registered", missing)
	}
}
