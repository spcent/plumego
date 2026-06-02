package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSameOriginMiddlewareAllowsSameOrigin(t *testing.T) {
	mw := SameOriginMiddleware(testLogger{})
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "http://app.local/api/connections", nil)
	req.Host = "app.local"
	req.Header.Set("Origin", "http://app.local")
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestSameOriginMiddlewareRejectsCrossOrigin(t *testing.T) {
	mw := SameOriginMiddleware(testLogger{})
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "http://app.local/api/connections", nil)
	req.Host = "app.local"
	req.Header.Set("Origin", "http://evil.local")
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestSameOriginMiddlewareAllowsHeaderlessLocalAutomation(t *testing.T) {
	mw := SameOriginMiddleware(testLogger{})
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "http://app.local/api/connections", nil)
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
