package resolve

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	tenantcore "github.com/spcent/plumego/tenant"
)

func TestMiddlewareFromPrincipal(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tenantcore.TenantIDFromContext(r.Context()) != "t-1" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = contract.RequestWithPrincipal(req, &contract.Principal{TenantID: "t-1"})
	rec := httptest.NewRecorder()

	mw := Middleware(Options{})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestMiddlewareMissing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw := Middleware(Options{})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
