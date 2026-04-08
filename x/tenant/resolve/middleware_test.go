package resolve

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/security/authn"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
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
	req = authn.RequestWithPrincipal(req, &authn.Principal{TenantID: "t-1"})
	rec := httptest.NewRecorder()

	mw := Middleware(Options{})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestMiddlewarePrincipalTakesPrecedenceOverHeaderAndExtractor(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tenantcore.TenantIDFromContext(r.Context()) != "principal-tenant" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/?tenant=query-tenant", nil)
	req.Header.Set("X-Tenant-ID", "header-tenant")
	req = authn.RequestWithPrincipal(req, &authn.Principal{TenantID: "principal-tenant"})
	rec := httptest.NewRecorder()

	mw := Middleware(Options{
		Extractor: tenantcore.FromQuery("tenant"),
	})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestMiddlewareInvalidTenantID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "bad tenant id")
	rec := httptest.NewRecorder()

	mw := Middleware(Options{})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
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
