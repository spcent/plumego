package tenant

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/tenant"
)

func TestTenantRateLimit(t *testing.T) {
	provider := tenant.NewInMemoryRateLimitProvider()
	provider.SetRateLimit("t-1", tenant.RateLimitConfig{
		RequestsPerSecond: 1,
		Burst:             1,
	})

	limiter := tenant.NewTokenBucketRateLimiter(provider)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := TenantRateLimit(TenantRateLimitOptions{Limiter: limiter})

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
