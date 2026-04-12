package resolve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/security/authn"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	tenanttransport "github.com/spcent/plumego/x/tenant/transport"
)

func TestMiddlewareFromHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tenantcore.TenantIDFromContext(r.Context()) != "t-1" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(authn.WithPrincipal(req.Context(), &authn.Principal{TenantID: "t-1"}))
	rec := httptest.NewRecorder()

	mw := Middleware(Options{})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestMiddlewareExtractorTakesPrecedenceOverHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tenantcore.TenantIDFromContext(r.Context()) != "query-tenant" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/?tenant=query-tenant", nil)
	req.Header.Set("X-Tenant-ID", "header-tenant")
	rec := httptest.NewRecorder()

	mw := Middleware(Options{
		Extractor:        tenantcore.FromQuery("tenant"),
		DisablePrincipal: true,
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

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != tenanttransport.CodeInvalidID {
		t.Fatalf("unexpected error code %q", body.Error.Code)
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

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != tenanttransport.CodeRequired {
		t.Fatalf("unexpected error code %q", body.Error.Code)
	}
}
