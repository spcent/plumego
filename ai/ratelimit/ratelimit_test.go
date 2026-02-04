package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestTokenBucketLimiter_Allow(t *testing.T) {
	limiter := NewTokenBucketLimiter(5, 1.0) // 5 tokens, refill 1/sec
	ctx := context.Background()

	// Should allow first 5 requests (burst)
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, "user1")
		if err != nil {
			t.Fatalf("Allow() error = %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// 6th request should be blocked
	allowed, err := limiter.Allow(ctx, "user1")
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if allowed {
		t.Error("6th request should be blocked")
	}

	// After 1 second, 1 token should be refilled
	time.Sleep(1100 * time.Millisecond)

	allowed, err = limiter.Allow(ctx, "user1")
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if !allowed {
		t.Error("Request after refill should be allowed")
	}
}

func TestTokenBucketLimiter_MultipleKeys(t *testing.T) {
	limiter := NewTokenBucketLimiter(3, 1.0)
	ctx := context.Background()

	// user1 uses 3 tokens
	for i := 0; i < 3; i++ {
		allowed, _ := limiter.Allow(ctx, "user1")
		if !allowed {
			t.Errorf("user1 request %d should be allowed", i)
		}
	}

	// user1 exhausted
	allowed, _ := limiter.Allow(ctx, "user1")
	if allowed {
		t.Error("user1 should be rate limited")
	}

	// user2 should still have tokens
	for i := 0; i < 3; i++ {
		allowed, _ := limiter.Allow(ctx, "user2")
		if !allowed {
			t.Errorf("user2 request %d should be allowed", i)
		}
	}
}

func TestTokenBucketLimiter_Remaining(t *testing.T) {
	limiter := NewTokenBucketLimiter(10, 2.0) // 10 tokens, refill 2/sec
	ctx := context.Background()

	// Check initial remaining
	remaining, err := limiter.Remaining(ctx, "user1")
	if err != nil {
		t.Fatalf("Remaining() error = %v", err)
	}
	if remaining != 10 {
		t.Errorf("Remaining = %d, want 10", remaining)
	}

	// Consume 3 tokens
	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, "user1")
	}

	remaining, _ = limiter.Remaining(ctx, "user1")
	if remaining != 7 {
		t.Errorf("Remaining = %d, want 7", remaining)
	}

	// After 1 second, should have refilled 2 tokens
	time.Sleep(1100 * time.Millisecond)
	remaining, _ = limiter.Remaining(ctx, "user1")
	if remaining != 9 {
		t.Errorf("Remaining after refill = %d, want 9", remaining)
	}
}

func TestTokenBucketLimiter_Reset(t *testing.T) {
	limiter := NewTokenBucketLimiter(5, 1.0)
	ctx := context.Background()

	// Exhaust tokens
	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, "user1")
	}

	// Should be blocked
	allowed, _ := limiter.Allow(ctx, "user1")
	if allowed {
		t.Error("Should be rate limited before reset")
	}

	// Reset
	err := limiter.Reset(ctx, "user1")
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Should be allowed again
	allowed, _ = limiter.Allow(ctx, "user1")
	if !allowed {
		t.Error("Should be allowed after reset")
	}
}

func TestTokenBucketLimiter_ResetAll(t *testing.T) {
	limiter := NewTokenBucketLimiter(3, 1.0)
	ctx := context.Background()

	// Exhaust tokens for multiple keys
	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, "user1")
		limiter.Allow(ctx, "user2")
	}

	// Both should be blocked
	allowed1, _ := limiter.Allow(ctx, "user1")
	allowed2, _ := limiter.Allow(ctx, "user2")
	if allowed1 || allowed2 {
		t.Error("Both users should be rate limited")
	}

	// Reset all
	err := limiter.ResetAll(ctx)
	if err != nil {
		t.Fatalf("ResetAll() error = %v", err)
	}

	// Both should be allowed
	allowed1, _ = limiter.Allow(ctx, "user1")
	allowed2, _ = limiter.Allow(ctx, "user2")
	if !allowed1 || !allowed2 {
		t.Error("Both users should be allowed after reset")
	}
}

func TestTokenBucketLimiter_Concurrent(t *testing.T) {
	limiter := NewTokenBucketLimiter(100, 50.0) // High limits for concurrency test
	ctx := context.Background()

	var wg sync.WaitGroup
	allowedCount := 0
	var mu sync.Mutex

	// 200 concurrent requests
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _ := limiter.Allow(ctx, "user1")
			if allowed {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should allow around 100 requests (initial burst)
	if allowedCount < 90 || allowedCount > 110 {
		t.Errorf("Allowed count = %d, expected around 100", allowedCount)
	}
}

func TestTokenBucketLimiter_Stats(t *testing.T) {
	limiter := NewTokenBucketLimiter(10, 5.0)
	ctx := context.Background()

	// Initially no buckets
	stats := limiter.Stats()
	if stats.ActiveBuckets != 0 {
		t.Errorf("ActiveBuckets = %d, want 0", stats.ActiveBuckets)
	}

	// Create buckets
	limiter.Allow(ctx, "user1")
	limiter.Allow(ctx, "user2")
	limiter.Allow(ctx, "user3")

	stats = limiter.Stats()
	if stats.ActiveBuckets != 3 {
		t.Errorf("ActiveBuckets = %d, want 3", stats.ActiveBuckets)
	}
	if stats.TotalCapacity != 10 {
		t.Errorf("TotalCapacity = %d, want 10", stats.TotalCapacity)
	}
	if stats.RefillRate != 5.0 {
		t.Errorf("RefillRate = %v, want 5.0", stats.RefillRate)
	}
}

func TestSlidingWindowLimiter_Allow(t *testing.T) {
	limiter := NewSlidingWindowLimiter(3, 1*time.Second)
	ctx := context.Background()

	// Should allow first 3 requests
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, "user1")
		if err != nil {
			t.Fatalf("Allow() error = %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// 4th request should be blocked
	allowed, _ := limiter.Allow(ctx, "user1")
	if allowed {
		t.Error("4th request should be blocked")
	}

	// After window expires, should allow again
	time.Sleep(1100 * time.Millisecond)
	allowed, _ = limiter.Allow(ctx, "user1")
	if !allowed {
		t.Error("Request after window should be allowed")
	}
}

func TestSlidingWindowLimiter_Remaining(t *testing.T) {
	limiter := NewSlidingWindowLimiter(5, 1*time.Second)
	ctx := context.Background()

	// Check initial remaining
	remaining, err := limiter.Remaining(ctx, "user1")
	if err != nil {
		t.Fatalf("Remaining() error = %v", err)
	}
	if remaining != 5 {
		t.Errorf("Remaining = %d, want 5", remaining)
	}

	// Consume 2 requests
	limiter.Allow(ctx, "user1")
	limiter.Allow(ctx, "user1")

	remaining, _ = limiter.Remaining(ctx, "user1")
	if remaining != 3 {
		t.Errorf("Remaining = %d, want 3", remaining)
	}

	// After window, should reset
	time.Sleep(1100 * time.Millisecond)
	remaining, _ = limiter.Remaining(ctx, "user1")
	if remaining != 5 {
		t.Errorf("Remaining after window = %d, want 5", remaining)
	}
}

func TestSlidingWindowLimiter_Reset(t *testing.T) {
	limiter := NewSlidingWindowLimiter(2, 1*time.Second)
	ctx := context.Background()

	// Exhaust limit
	limiter.Allow(ctx, "user1")
	limiter.Allow(ctx, "user1")

	// Should be blocked
	allowed, _ := limiter.Allow(ctx, "user1")
	if allowed {
		t.Error("Should be rate limited before reset")
	}

	// Reset
	err := limiter.Reset(ctx, "user1")
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Should be allowed again
	allowed, _ = limiter.Allow(ctx, "user1")
	if !allowed {
		t.Error("Should be allowed after reset")
	}
}

func TestNoOpLimiter(t *testing.T) {
	limiter := &NoOpLimiter{}
	ctx := context.Background()

	// Should always allow
	for i := 0; i < 1000; i++ {
		allowed, err := limiter.Allow(ctx, "user1")
		if err != nil {
			t.Fatalf("Allow() error = %v", err)
		}
		if !allowed {
			t.Error("NoOpLimiter should always allow requests")
		}
	}

	// Remaining should be max
	remaining, _ := limiter.Remaining(ctx, "user1")
	if remaining <= 0 {
		t.Errorf("Remaining = %d, should be very large", remaining)
	}
}

func BenchmarkTokenBucketLimiter_Allow(b *testing.B) {
	limiter := NewTokenBucketLimiter(1000, 100.0)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ctx, "user1")
	}
}

func BenchmarkTokenBucketLimiter_Concurrent(b *testing.B) {
	limiter := NewTokenBucketLimiter(10000, 1000.0)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Allow(ctx, "user1")
		}
	})
}

func BenchmarkSlidingWindowLimiter_Allow(b *testing.B) {
	limiter := NewSlidingWindowLimiter(1000, 1*time.Second)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ctx, "user1")
	}
}
