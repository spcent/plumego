package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestRequestIDUsesHeader(t *testing.T) {
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := contract.TraceIDFromContext(r.Context()); got != "abc-123" {
			t.Fatalf("expected trace id to be propagated, got %q", got)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "abc-123" {
		t.Fatalf("expected response header to match, got %q", got)
	}
}

func TestRequestIDUsesFallbackHeader(t *testing.T) {
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := contract.TraceIDFromContext(r.Context()); got != "trace-xyz" {
			t.Fatalf("expected trace id to be propagated, got %q", got)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Trace-ID", "trace-xyz")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "trace-xyz" {
		t.Fatalf("expected response header to match fallback, got %q", got)
	}
}

func TestRequestIDGeneratesWhenMissing(t *testing.T) {
	handler := RequestID(WithRequestIDGenerator(func() string { return "gen-1" }))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := contract.TraceIDFromContext(r.Context()); got != "gen-1" {
			t.Fatalf("expected generated trace id, got %q", got)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "gen-1" {
		t.Fatalf("expected generated header, got %q", got)
	}
}
