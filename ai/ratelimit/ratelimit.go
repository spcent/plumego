// Package ratelimit provides rate limiting capabilities for the AI Agent Gateway.
//
// Design Philosophy:
// - Token bucket algorithm for smooth rate limiting
// - Thread-safe concurrent access
// - Zero third-party dependencies
// - Configurable refill rates and burst capacity
package ratelimit

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// RateLimiter is the interface for rate limiting.
type RateLimiter interface {
	// Allow checks if a request is allowed for the given key.
	// Returns true if allowed, false if rate limit exceeded.
	Allow(ctx context.Context, key string) (bool, error)

	// Remaining returns the number of remaining tokens for the key.
	Remaining(ctx context.Context, key string) (int, error)

	// Reset resets the rate limiter for a specific key.
	Reset(ctx context.Context, key string) error

	// ResetAll resets all rate limiters.
	ResetAll(ctx context.Context) error
}

// TokenBucketLimiter implements rate limiting using the token bucket algorithm.
//
// The token bucket algorithm allows bursts of traffic while maintaining
// a steady average rate. Tokens are added to the bucket at a constant rate,
// and each request consumes one token.
type TokenBucketLimiter struct {
	capacity   int           // Maximum tokens in bucket
	refillRate float64       // Tokens per second
	buckets    map[string]*bucket
	mu         sync.RWMutex
}

// bucket represents a token bucket for a single key.
type bucket struct {
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// TokenBucketConfig configures a token bucket rate limiter.
type TokenBucketConfig struct {
	Capacity   int           // Maximum burst size (tokens)
	RefillRate float64       // Tokens per second
	CleanupInt time.Duration // Cleanup interval for unused buckets
}

// NewTokenBucketLimiter creates a new token bucket rate limiter.
func NewTokenBucketLimiter(capacity int, refillRate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*bucket),
	}
}

// NewTokenBucketLimiterWithConfig creates a rate limiter with custom config.
func NewTokenBucketLimiterWithConfig(config TokenBucketConfig) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		capacity:   config.Capacity,
		refillRate: config.RefillRate,
		buckets:    make(map[string]*bucket),
	}

	// Start cleanup goroutine if configured
	if config.CleanupInt > 0 {
		go limiter.cleanup(config.CleanupInt)
	}

	return limiter
}

// Allow implements RateLimiter.
func (tbl *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	// Get or create bucket for this key
	b := tbl.getOrCreateBucket(key)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	tokensToAdd := elapsed * tbl.refillRate

	// Calculate new token count (capped at capacity)
	b.tokens = math.Min(b.tokens+tokensToAdd, float64(tbl.capacity))
	b.lastRefill = now

	// Check if request allowed
	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true, nil
	}

	return false, nil
}

// Remaining implements RateLimiter.
func (tbl *TokenBucketLimiter) Remaining(ctx context.Context, key string) (int, error) {
	tbl.mu.RLock()
	b, exists := tbl.buckets[key]
	tbl.mu.RUnlock()

	if !exists {
		return tbl.capacity, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens first
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	tokensToAdd := elapsed * tbl.refillRate
	tokens := math.Min(b.tokens+tokensToAdd, float64(tbl.capacity))

	return int(math.Floor(tokens)), nil
}

// Reset implements RateLimiter.
func (tbl *TokenBucketLimiter) Reset(ctx context.Context, key string) error {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	delete(tbl.buckets, key)
	return nil
}

// ResetAll implements RateLimiter.
func (tbl *TokenBucketLimiter) ResetAll(ctx context.Context) error {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	tbl.buckets = make(map[string]*bucket)
	return nil
}

// getOrCreateBucket gets or creates a bucket for the given key.
func (tbl *TokenBucketLimiter) getOrCreateBucket(key string) *bucket {
	tbl.mu.RLock()
	b, exists := tbl.buckets[key]
	tbl.mu.RUnlock()

	if exists {
		return b
	}

	// Create new bucket
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	// Double-check after acquiring write lock
	b, exists = tbl.buckets[key]
	if exists {
		return b
	}

	b = &bucket{
		tokens:     float64(tbl.capacity),
		lastRefill: time.Now(),
	}
	tbl.buckets[key] = b

	return b
}

// cleanup removes unused buckets periodically to prevent memory growth.
func (tbl *TokenBucketLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		tbl.mu.Lock()

		// Remove buckets that haven't been used recently
		cutoff := time.Now().Add(-interval * 2)
		for key, b := range tbl.buckets {
			b.mu.Lock()
			if b.lastRefill.Before(cutoff) {
				delete(tbl.buckets, key)
			}
			b.mu.Unlock()
		}

		tbl.mu.Unlock()
	}
}

// Stats returns statistics about the rate limiter.
type Stats struct {
	ActiveBuckets int
	TotalCapacity int
	RefillRate    float64
}

// Stats returns rate limiter statistics.
func (tbl *TokenBucketLimiter) Stats() Stats {
	tbl.mu.RLock()
	defer tbl.mu.RUnlock()

	return Stats{
		ActiveBuckets: len(tbl.buckets),
		TotalCapacity: tbl.capacity,
		RefillRate:    tbl.refillRate,
	}
}

// SlidingWindowLimiter implements rate limiting using a sliding window algorithm.
//
// The sliding window algorithm provides more accurate rate limiting by
// counting requests in a sliding time window, preventing burst abuse.
type SlidingWindowLimiter struct {
	limit      int
	windowSize time.Duration
	windows    map[string]*window
	mu         sync.RWMutex
}

// window represents a sliding window of requests.
type window struct {
	requests []time.Time
	mu       sync.Mutex
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter.
func NewSlidingWindowLimiter(limit int, windowSize time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:      limit,
		windowSize: windowSize,
		windows:    make(map[string]*window),
	}
}

// Allow implements RateLimiter.
func (swl *SlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	w := swl.getOrCreateWindow(key)

	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-swl.windowSize)

	// Remove old requests outside the window
	validRequests := make([]time.Time, 0, len(w.requests))
	for _, reqTime := range w.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	w.requests = validRequests

	// Check if under limit
	if len(w.requests) < swl.limit {
		w.requests = append(w.requests, now)
		return true, nil
	}

	return false, nil
}

// Remaining implements RateLimiter.
func (swl *SlidingWindowLimiter) Remaining(ctx context.Context, key string) (int, error) {
	swl.mu.RLock()
	w, exists := swl.windows[key]
	swl.mu.RUnlock()

	if !exists {
		return swl.limit, nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-swl.windowSize)

	// Count valid requests
	validCount := 0
	for _, reqTime := range w.requests {
		if reqTime.After(cutoff) {
			validCount++
		}
	}

	remaining := swl.limit - validCount
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// Reset implements RateLimiter.
func (swl *SlidingWindowLimiter) Reset(ctx context.Context, key string) error {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	delete(swl.windows, key)
	return nil
}

// ResetAll implements RateLimiter.
func (swl *SlidingWindowLimiter) ResetAll(ctx context.Context) error {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	swl.windows = make(map[string]*window)
	return nil
}

// getOrCreateWindow gets or creates a window for the given key.
func (swl *SlidingWindowLimiter) getOrCreateWindow(key string) *window {
	swl.mu.RLock()
	w, exists := swl.windows[key]
	swl.mu.RUnlock()

	if exists {
		return w
	}

	swl.mu.Lock()
	defer swl.mu.Unlock()

	// Double-check after acquiring write lock
	w, exists = swl.windows[key]
	if exists {
		return w
	}

	w = &window{
		requests: make([]time.Time, 0),
	}
	swl.windows[key] = w

	return w
}

// NoOpLimiter is a rate limiter that always allows requests (for testing/disabled rate limiting).
type NoOpLimiter struct{}

// Allow implements RateLimiter.
func (n *NoOpLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return true, nil
}

// Remaining implements RateLimiter.
func (n *NoOpLimiter) Remaining(ctx context.Context, key string) (int, error) {
	return math.MaxInt32, nil
}

// Reset implements RateLimiter.
func (n *NoOpLimiter) Reset(ctx context.Context, key string) error {
	return nil
}

// ResetAll implements RateLimiter.
func (n *NoOpLimiter) ResetAll(ctx context.Context) error {
	return nil
}

// Error types
var (
	ErrRateLimitExceeded = fmt.Errorf("rate limit exceeded")
)
