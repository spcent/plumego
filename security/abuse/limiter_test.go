package abuse

import (
	"testing"
	"time"
)

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.now = c.now.Add(d)
}

func TestLimiterAllow(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	limiter := NewLimiter(Config{
		Rate:     1,
		Capacity: 2,
		Now:      clock.Now,
	})
	defer limiter.Stop()

	decision := limiter.Allow("client")
	if !decision.Allowed {
		t.Fatalf("expected first request to be allowed")
	}
	if decision.Remaining != 1 {
		t.Fatalf("expected remaining 1, got %d", decision.Remaining)
	}

	decision = limiter.Allow("client")
	if !decision.Allowed {
		t.Fatalf("expected second request to be allowed")
	}
	if decision.Remaining != 0 {
		t.Fatalf("expected remaining 0, got %d", decision.Remaining)
	}

	decision = limiter.Allow("client")
	if decision.Allowed {
		t.Fatalf("expected third request to be rate limited")
	}
	if decision.RetryAfter == 0 {
		t.Fatalf("expected retry-after to be set")
	}

	clock.Advance(time.Second)
	decision = limiter.Allow("client")
	if !decision.Allowed {
		t.Fatalf("expected request after refill to be allowed")
	}
}

func TestLimiterCleanup(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	limiter := NewLimiter(Config{
		Rate:            1,
		Capacity:        1,
		Now:             clock.Now,
		CleanupInterval: time.Hour,
		MaxIdle:         time.Minute,
	})
	defer limiter.Stop()

	limiter.Allow("client")
	if len(limiter.buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(limiter.buckets))
	}

	clock.Advance(2 * time.Minute)
	limiter.cleanup(clock.Now())

	if len(limiter.buckets) != 0 {
		t.Fatalf("expected buckets to be cleaned up, got %d", len(limiter.buckets))
	}
}

func TestLimiterUnknownKey(t *testing.T) {
	clock := &fakeClock{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	limiter := NewLimiter(Config{
		Rate:            1,
		Capacity:        1,
		Now:             clock.Now,
		CleanupInterval: time.Hour,
		MaxIdle:         time.Hour,
	})
	defer limiter.Stop()

	limiter.Allow("")
	if _, ok := limiter.buckets["unknown"]; !ok {
		t.Fatalf("expected empty key to map to unknown bucket")
	}
}

func BenchmarkLimiterAllow(b *testing.B) {
	limit := b.N + 1
	limiter := NewLimiter(Config{
		Rate:            float64(limit),
		Capacity:        limit,
		CleanupInterval: time.Hour,
		MaxIdle:         time.Hour,
	})
	defer limiter.Stop()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		limiter.Allow("client")
	}
}
