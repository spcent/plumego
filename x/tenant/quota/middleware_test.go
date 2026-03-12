package quota

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tenantcore "github.com/spcent/plumego/tenant"
)

func TestMiddlewareExceeded(t *testing.T) {
	cfg := tenantcore.NewInMemoryConfigManager()
	cfg.SetTenantConfig(tenantcore.Config{
		TenantID: "t-1",
		Quota: tenantcore.QuotaConfig{
			RequestsPerMinute: 1,
		},
	})
	manager := tenantcore.NewInMemoryQuotaManager(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(Options{Manager: manager})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = tenantcore.RequestWithTenantID(req, "t-1")

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
