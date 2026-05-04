package security

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
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

func TestSecurityHeadersMiddlewareInvalidPolicyFailsClosed(t *testing.T) {
	policy := headers.DefaultPolicy()
	policy.Additional = map[string]string{"Bad Header": "value"}
	mw := SecurityHeaders(&policy)

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if called {
		t.Fatal("invalid policy should not call downstream handler")
	}
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	var response contract.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Type != contract.TypeInternal {
		t.Fatalf("error type = %q, want %q", response.Error.Type, contract.TypeInternal)
	}
}
