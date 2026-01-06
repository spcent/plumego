package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultSecurityHeadersGuardrail(t *testing.T) {
	app := New()
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if got := resp.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options nosniff, got %q", got)
	}
}

func TestDisableSecurityHeadersGuardrail(t *testing.T) {
	app := New(WithSecurityHeadersEnabled(false))
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if got := resp.Header().Get("X-Content-Type-Options"); got != "" {
		t.Fatalf("expected X-Content-Type-Options to be empty, got %q", got)
	}
}

func TestDefaultAbuseGuardHeaders(t *testing.T) {
	app := New()
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if got := resp.Header().Get("X-RateLimit-Limit"); got == "" {
		t.Fatalf("expected X-RateLimit-Limit header to be set")
	}
}

func TestDisableAbuseGuardHeaders(t *testing.T) {
	app := New(WithAbuseGuardEnabled(false))
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if got := resp.Header().Get("X-RateLimit-Limit"); got != "" {
		t.Fatalf("expected X-RateLimit-Limit to be empty, got %q", got)
	}
}
