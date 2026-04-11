package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestMiddlewareUsesHeader(t *testing.T) {
	handler := Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := contract.RequestIDFromContext(r.Context()); got != "abc-123" {
			t.Fatalf("expected request id to be propagated, got %q", got)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(contract.RequestIDHeader, "abc-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(contract.RequestIDHeader); got != "abc-123" {
		t.Fatalf("expected response header to match, got %q", got)
	}
}

func TestMiddlewareGeneratesWhenMissing(t *testing.T) {
	handler := Middleware(WithGenerator(func() string { return "gen-1" }))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := contract.RequestIDFromContext(r.Context()); got != "gen-1" {
			t.Fatalf("expected generated request id, got %q", got)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(contract.RequestIDHeader); got != "gen-1" {
		t.Fatalf("expected generated header, got %q", got)
	}
}
