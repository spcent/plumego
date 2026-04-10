package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tenantcore "github.com/spcent/plumego/x/tenant/core"
)

func TestMiddleware(t *testing.T) {
	provider := tenantcore.NewInMemoryRateLimitManager()
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

func TestMiddlewareTenantIsolation(t *testing.T) {
	provider := tenantcore.NewInMemoryRateLimitManager()
	provider.SetRateLimit("t-1", tenantcore.RateLimitConfig{
		RequestsPerSecond: 1,
		Burst:             1,
	})
	provider.SetRateLimit("t-2", tenantcore.RateLimitConfig{
		RequestsPerSecond: 1,
		Burst:             1,
	})

	limiter := tenantcore.NewTokenBucketRateLimiter(provider)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(Options{Limiter: limiter})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1 = tenantcore.RequestWithTenantID(req1, "t-1")
	rec1 := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("tenant t-1 first status = %d, want 200", rec1.Code)
	}

	rec1 = httptest.NewRecorder()
	mw(handler).ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusTooManyRequests {
		t.Fatalf("tenant t-1 second status = %d, want 429", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2 = tenantcore.RequestWithTenantID(req2, "t-2")
	rec2 := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("tenant t-2 first status = %d, want 200", rec2.Code)
	}
}
