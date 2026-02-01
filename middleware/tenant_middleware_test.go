package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/tenant"
)

func TestTenantResolverFromPrincipal(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tenant.TenantIDFromContext(r.Context()) != "t-1" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = contract.RequestWithPrincipal(req, &contract.Principal{TenantID: "t-1"})
	rec := httptest.NewRecorder()

	mw := TenantResolver(TenantResolverOptions{})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestTenantResolverMissing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw := TenantResolver(TenantResolverOptions{})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestTenantPolicyDenied(t *testing.T) {
	evaluator := tenant.NewConfigPolicyEvaluator(&tenant.InMemoryConfigManager{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = tenant.RequestWithTenantID(req, "t-1")
	req.Header.Set("X-Model", "gpt-4o")
	rec := httptest.NewRecorder()

	mw := TenantPolicy(TenantPolicyOptions{Evaluator: evaluator})
	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
}

func TestTenantQuotaExceeded(t *testing.T) {
	cfg := tenant.NewInMemoryConfigManager()
	cfg.SetTenantConfig(tenant.Config{
		TenantID: "t-1",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 1,
		},
	})
	manager := tenant.NewInMemoryQuotaManager(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := TenantQuota(TenantQuotaOptions{Manager: manager})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = tenant.RequestWithTenantID(req, "t-1")

	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}
}
