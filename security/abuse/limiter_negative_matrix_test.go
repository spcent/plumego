package abuse

import (
	"testing"
	"time"
)

func TestLimiterNegativeMatrix_EmptyKeyFailsClosed(t *testing.T) {
	limiter := NewLimiter(Config{
		Rate:            1,
		Capacity:        1,
		MaxEntries:      10,
		CleanupInterval: time.Minute,
		MaxIdle:         time.Minute,
		Shards:          2,
	})
	defer limiter.Stop()

	first := limiter.Allow("")
	if first.Allowed {
		t.Fatalf("request with empty key should be denied")
	}
	if limiter.Metrics().Rejected != 1 {
		t.Fatalf("empty key rejected metrics = %d, want 1", limiter.Metrics().Rejected)
	}
}

func TestLimiterNegativeMatrix_NonMonotonicClockDoesNotOverRefill(t *testing.T) {
	now := time.Now()
	clock := now
	limiter := NewLimiter(Config{
		Rate:            1,
		Capacity:        2,
		MaxEntries:      10,
		CleanupInterval: time.Minute,
		MaxIdle:         time.Minute,
		Shards:          1,
		Now: func() time.Time {
			return clock
		},
	})
	defer limiter.Stop()

	_ = limiter.Allow("k")
	_ = limiter.Allow("k")
	third := limiter.Allow("k")
	if third.Allowed {
		t.Fatalf("expected third request to be limited with capacity=2 and no refill")
	}

	// Move clock backwards; limiter should clamp elapsed to zero and stay safe.
	clock = clock.Add(-10 * time.Second)
	got := limiter.Allow("k")
	if got.Allowed {
		t.Fatalf("request should stay limited when clock moves backwards")
	}
}
