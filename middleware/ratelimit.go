package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter represents a rate limiter that uses token bucket algorithm
type RateLimiter struct {
	// mu protects the buckets map
	mu sync.RWMutex
	// buckets stores token buckets per key (e.g., IP address)
	buckets map[string]*tokenBucket
	// rate is the number of tokens added per second
	rate float64
	// capacity is the maximum number of tokens a bucket can hold
	capacity int
	// cleanupInterval is how often to clean up old buckets
	cleanupInterval time.Duration
	// maxIdleTime is how long a bucket can be idle before being cleaned up
	maxIdleTime time.Duration
}

// tokenBucket represents a single token bucket for rate limiting
type tokenBucket struct {
	// tokens is the current number of tokens in the bucket
	tokens float64
	// lastRefill is the last time tokens were added to the bucket
	lastRefill time.Time
	// lastAccess is the last time this bucket was accessed
	lastAccess time.Time
}

// NewRateLimiter creates a new RateLimiter with the given rate and capacity
// rate: tokens per second
// capacity: maximum tokens per bucket
// cleanupInterval: how often to clean up idle buckets
// maxIdleTime: how long a bucket can be idle before cleanup
func NewRateLimiter(rate float64, capacity int, cleanupInterval, maxIdleTime time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:         make(map[string]*tokenBucket),
		rate:            rate,
		capacity:        capacity,
		cleanupInterval: cleanupInterval,
		maxIdleTime:     maxIdleTime,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given key is allowed
func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()

	// Get or create bucket for this key
	bucket := rl.getBucket(key, now)

	// Calculate tokens to add since last refill
	elapsed := now.Sub(bucket.lastRefill)
	addTokens := rl.rate * elapsed.Seconds()

	// Update bucket
	bucket.tokens = min(float64(rl.capacity), bucket.tokens+addTokens)
	bucket.lastRefill = now
	bucket.lastAccess = now

	// Check if we have enough tokens
	if bucket.tokens >= 1 {
		// Consume one token
		bucket.tokens--
		return true
	}

	return false
}

// getBucket returns the token bucket for the given key, creating it if needed
func (rl *RateLimiter) getBucket(key string, now time.Time) *tokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(rl.capacity),
			lastRefill: now,
			lastAccess: now,
		}
		rl.buckets[key] = bucket
	}

	return bucket
}

// cleanup removes idle buckets to free memory
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()

		for key, bucket := range rl.buckets {
			if now.Sub(bucket.lastAccess) > rl.maxIdleTime {
				delete(rl.buckets, key)
			}
		}

		rl.mu.Unlock()
	}
}

// RateLimit creates a middleware that limits requests based on IP address
// Example: RateLimit(10, 20, time.Minute, 5*time.Minute) allows 10 requests/second with burst up to 20
func RateLimit(rate float64, capacity int, cleanupInterval, maxIdleTime time.Duration) Middleware {
	// Create rate limiter
	limiter := NewRateLimiter(rate, capacity, cleanupInterval, maxIdleTime)

	return func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP address
			ip := r.RemoteAddr

			// Check if request is allowed
			if !limiter.Allow(ip) {
				// Request not allowed, return 429 Too Many Requests
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Too many requests, please try again later"}`))
				return
			}

			// Request allowed, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitFunc creates a FuncMiddleware version of RateLimit
func RateLimitFunc(rate float64, capacity int, cleanupInterval, maxIdleTime time.Duration) FuncMiddleware {
	// Create rate limiter
	limiter := NewRateLimiter(rate, capacity, cleanupInterval, maxIdleTime)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get client IP address
			ip := r.RemoteAddr

			// Check if request is allowed
			if !limiter.Allow(ip) {
				// Request not allowed, return 429 Too Many Requests
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Too many requests, please try again later"}`))
				return
			}

			// Request allowed, proceed to next handler
			next(w, r)
		}
	}
}
