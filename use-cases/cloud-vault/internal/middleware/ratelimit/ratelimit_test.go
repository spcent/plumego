package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_AllowsWithinBurst(t *testing.T) {
	mw := Middleware(Config{RequestsPerSecond: 1, Burst: 2})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: got status %d, want 200", i, rec.Code)
		}
	}
}

func TestMiddleware_RejectsOverBurst(t *testing.T) {
	mw := Middleware(Config{RequestsPerSecond: 1, Burst: 1})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.6.7.8:1234"

	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request: got status %d, want 200", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: got status %d, want 429", rec2.Code)
	}
}

func TestMiddleware_DisabledWhenZeroConfig(t *testing.T) {
	mw := Middleware(Config{})
	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("expected pass-through when disabled, got called=%v status=%d", called, rec.Code)
	}
}

func TestExtractIP_ForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "9.9.9.9, 10.0.0.1")
	req.RemoteAddr = "127.0.0.1:5555"
	if ip := extractIP(req); ip != "9.9.9.9" {
		t.Errorf("extractIP() = %q, want 9.9.9.9", ip)
	}
}
