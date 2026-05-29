package securityheaders

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/security/headers"
)

func mustSecurityHeaders(t *testing.T, config Config) func(http.Handler) http.Handler {
	t.Helper()
	mw, err := Middleware(config)
	if err != nil {
		t.Fatalf("Middleware returned error: %v", err)
	}
	return mw
}

func TestSecurityHeadersMiddlewareDefault(t *testing.T) {
	mw := mustSecurityHeaders(t, Config{})
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
	mw := mustSecurityHeaders(t, Config{Policy: &policy})
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
	if _, err := Middleware(Config{Policy: &policy}); err == nil {
		t.Fatal("expected invalid header policy error")
	}
}

func TestSecurityHeadersMiddlewareInvalidSemanticPolicyFailsClosed(t *testing.T) {
	policy := headers.DefaultPolicy()
	policy.FrameOptions = "ALLOWALL"
	if _, err := Middleware(Config{Policy: &policy}); err == nil {
		t.Fatal("expected invalid semantic policy error")
	}
}
