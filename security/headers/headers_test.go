package headers

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()

	policy.Apply(w, req)

	resp := w.Result()
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options nosniff, got %q", got)
	}
	if got := resp.Header.Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Fatalf("expected X-Frame-Options SAMEORIGIN, got %q", got)
	}
	if got := resp.Header.Get("Referrer-Policy"); got == "" {
		t.Fatalf("expected Referrer-Policy to be set")
	}
}

func TestStrictPolicyHSTS(t *testing.T) {
	policy := StrictPolicy()

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()
	policy.Apply(w, req)
	if got := w.Result().Header.Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("expected no HSTS on non-TLS request, got %q", got)
	}

	tlsReq := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	tlsReq.TLS = &tls.ConnectionState{}
	w = httptest.NewRecorder()
	policy.Apply(w, tlsReq)
	if got := w.Result().Header.Get("Strict-Transport-Security"); got == "" {
		t.Fatalf("expected HSTS on TLS request")
	}
}

func TestAdditionalHeadersValidation(t *testing.T) {
	policy := Policy{
		Additional: map[string]string{
			"X-Test":       "ok",
			"Bad Header":   "nope",
			"X-Injected":   "bad\nvalue",
			"X-Also-Valid": "still-ok",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()
	policy.Apply(w, req)

	resp := w.Result()
	if got := resp.Header.Get("X-Test"); got != "ok" {
		t.Fatalf("expected X-Test to be set, got %q", got)
	}
	if got := resp.Header.Get("Bad Header"); got != "" {
		t.Fatalf("expected invalid header name to be skipped")
	}
	if got := resp.Header.Get("X-Injected"); got != "" {
		t.Fatalf("expected invalid header value to be skipped")
	}
	if got := resp.Header.Get("X-Also-Valid"); got != "still-ok" {
		t.Fatalf("expected X-Also-Valid to be set, got %q", got)
	}
}
