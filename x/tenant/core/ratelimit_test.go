package tenant

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryRateLimitProvider(t *testing.T) {
	provider := NewInMemoryRateLimitProvider()
	provider.SetRateLimit("t-1", RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             20,
	})

	cfg, err := provider.RateLimitConfig(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.RequestsPerSecond != 10 || cfg.Burst != 20 {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	// Unknown tenant returns ErrTenantNotFound.
	_, err = provider.RateLimitConfig(context.Background(), "missing")
	if err != ErrTenantNotFound {
		t.Fatalf("expected ErrTenantNotFound for unconfigured tenant, got %v", err)
	}
}

func TestTokenBucketRateLimiter_UnknownTenantAllowed(t *testing.T) {
	// Unknown tenants (ErrTenantNotFound from provider) are allowed — treated as unlimited.
	provider := NewInMemoryRateLimitProvider()
	limiter := NewTokenBucketRateLimiter(provider)
	res, err := limiter.Allow(context.Background(), "unknown-tenant", RateLimitRequest{})
	if err != nil || !res.Allowed {
		t.Fatalf("expected unknown tenant to be allowed, got err=%v res=%+v", err, res)
	}
}

func TestTokenBucketRateLimiter_AllowAndRefill(t *testing.T) {
	provider := NewInMemoryRateLimitProvider()
	provider.SetRateLimit("t-1", RateLimitConfig{
		RequestsPerSecond: 2,
		Burst:             2,
	})

	limiter := NewTokenBucketRateLimiter(provider)
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	res, err := limiter.Allow(context.Background(), "t-1", RateLimitRequest{
		Tokens: 1,
		Now:    now,
	})
	if err != nil || !res.Allowed {
		t.Fatalf("expected first request allowed, got err=%v res=%+v", err, res)
	}

	res, err = limiter.Allow(context.Background(), "t-1", RateLimitRequest{
		Tokens: 1,
		Now:    now,
	})
	if err != nil || !res.Allowed {
		t.Fatalf("expected second request allowed, got err=%v res=%+v", err, res)
	}
	if res.Remaining != 0 {
		t.Fatalf("expected remaining 0, got %d", res.Remaining)
	}

	res, err = limiter.Allow(context.Background(), "t-1", RateLimitRequest{
		Tokens: 1,
		Now:    now,
	})
	if err != ErrRateLimitExceeded || res.Allowed {
		t.Fatalf("expected rate limited, got err=%v res=%+v", err, res)
	}
	if res.RetryAfter <= 0 {
		t.Fatalf("expected retry after > 0, got %v", res.RetryAfter)
	}

	halfSecond := now.Add(500 * time.Millisecond)
	res, err = limiter.Allow(context.Background(), "t-1", RateLimitRequest{
		Tokens: 1,
		Now:    halfSecond,
	})
	if err != nil || !res.Allowed {
		t.Fatalf("expected refill allow, got err=%v res=%+v", err, res)
	}
}

func TestTokenBucketRateLimiter_Unlimited(t *testing.T) {
	provider := NewInMemoryRateLimitProvider()
	provider.SetRateLimit("t-1", RateLimitConfig{})

	limiter := NewTokenBucketRateLimiter(provider)
	res, err := limiter.Allow(context.Background(), "t-1", RateLimitRequest{})
	if err != nil || !res.Allowed {
		t.Fatalf("expected allowed for unlimited config, got err=%v res=%+v", err, res)
	}
}
