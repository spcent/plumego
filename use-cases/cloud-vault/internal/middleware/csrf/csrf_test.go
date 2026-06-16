package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func handlerOK() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_RejectsPostWithoutHeader(t *testing.T) {
	h := Middleware(handlerOK())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("got status %d, want 403", rec.Code)
	}
}

func TestMiddleware_AllowsPostWithHeader(t *testing.T) {
	h := Middleware(handlerOK())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
}

func TestMiddleware_AllowsGetWithoutHeader(t *testing.T) {
	h := Middleware(handlerOK())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
}
