package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tenantcore "github.com/spcent/plumego/x/tenant/core"
)

func TestMiddleware(t *testing.T) {
	provider := tenantcore.NewInMemoryRateLimitProvider()
	provider.SetRateLimit("t-1", tenantcore.RateLimitConfig{
		RequestsPerSecond: 1,
		Burst:             1,
	})

	limiter := tenantcore.NewTokenBucketRateLimiter(provider)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(Options{Limiter: limiter})

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
