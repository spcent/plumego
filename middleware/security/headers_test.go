package security

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/security/headers"
)

func TestSecurityHeadersMiddlewareDefault(t *testing.T) {
	mw := SecurityHeaders(nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options nosniff, got %q", got)
	}
}

func TestSecurityHeadersMiddlewareStrictHSTS(t *testing.T) {
	policy := headers.StrictPolicy()
	mw := SecurityHeaders(&policy)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	req.TLS = &tls.ConnectionState{}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	if got := resp.Header.Get("Strict-Transport-Security"); got == "" {
		t.Fatalf("expected HSTS header to be set")
	}
}
