// Package ratelimit provides advanced rate limiting middleware
//
// This package implements token bucket algorithm for rate limiting with:
//   - Configurable token refill rate
//   - Burst capacity
//   - Per-endpoint quotas
//   - Distributed rate limiting via storage contracts
//
// Example usage:
//
//	import (
//		"github.com/spcent/plumego/middleware/ratelimit"
//		"github.com/spcent/plumego/core"
//	)
//
//	app := core.New()
//
//	// Simple rate limiting
//	app.Use(ratelimit.TokenBucket(ratelimit.Config{
//		Capacity:   100,     // 100 tokens max
//		RefillRate: 10,      // 10 tokens per second
//		KeyFunc:    ratelimit.IPKeyFunc,
//	}))
package ratelimit

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
)

// Config holds token bucket rate limiter configuration
type Config struct {
	// Capacity is the maximum number of tokens
	// This allows burst traffic up to this limit
	// Default: 100
	Capacity int64

	// RefillRate is the number of tokens added per second
	// Default: 10
	RefillRate int64

	// KeyFunc generates rate limit keys from requests
	// Default: IPKeyFunc (limit by client IP)
	KeyFunc KeyFunc

	// Store is the storage backend for distributed rate limiting
	// Default: NewMemoryStore() (in-memory, single instance)
	Store Store

	// OnRateLimit is called when request is rate limited (optional)
	OnRateLimit func(key string, retryAfter time.Duration)

	// OnAllow is called when request is allowed (optional)
	OnAllow func(key string, tokensRemaining int64)
}

// KeyFunc generates a rate limit key from request
type KeyFunc func(r *http.Request) string

// Store is the interface for rate limit storage backends
// Users can implement this for distributed rate limiting (Redis, etcd, etc.)
type Store interface {
	// GetBucket retrieves a token bucket
	GetBucket(key string) (*Bucket, error)

	// UpdateBucket updates a token bucket
	UpdateBucket(key string, bucket *Bucket) error

	// DeleteBucket removes a token bucket
	DeleteBucket(key string) error
}

// AtomicTokenStore extends Store with atomic token consumption semantics.
// Implementations should perform refill + consume as one critical section.
type AtomicTokenStore interface {
	Store
	TakeToken(key string, capacity, refillRate int64, now time.Time) (allowed bool, bucket Bucket, err error)
}

// Bucket represents a token bucket
type Bucket struct {
	// Tokens is the current number of tokens
	Tokens int64

	// LastRefill is the last time tokens were refilled
	LastRefill time.Time

	// Capacity is the maximum number of tokens
	Capacity int64

	// RefillRate is tokens added per second
	RefillRate int64
}

// TokenBucket creates a token bucket rate limiting middleware
func TokenBucket(config Config) func(http.Handler) http.Handler {
	cfg := config.withDefaults()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate key
			key := cfg.KeyFunc(r)

			now := time.Now()
			var (
				bucket  Bucket
				allowed bool
				err     error
			)
			if atomicStore, ok := cfg.Store.(AtomicTokenStore); ok {
				allowed, bucket, err = atomicStore.TakeToken(key, cfg.Capacity, cfg.RefillRate, now)
			} else {
				allowed, bucket, err = consumeTokenNonAtomic(cfg, key, now)
			}
			if err != nil {
				contract.WriteError(w, r, contract.NewInternalError("Rate limiter internal error"))
				return
			}

			if allowed {
				// Add rate limit headers
				w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(bucket.Capacity, 10))
				w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(bucket.Tokens, 10))
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(bucket.LastRefill.Add(time.Second).Unix(), 10))

				// Call hook
				if cfg.OnAllow != nil {
					cfg.OnAllow(key, bucket.Tokens)
				}

				// Allow request
				next.ServeHTTP(w, r)
				return
			}

			// Rate limited - calculate retry-after
			retryAfter := time.Second / time.Duration(bucket.RefillRate)

			// Call hook
			if cfg.OnRateLimit != nil {
				cfg.OnRateLimit(key, retryAfter)
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(bucket.Capacity, 10))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(bucket.LastRefill.Add(retryAfter).Unix(), 10))
			w.Header().Set("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds())+1, 10))

			// Return 429 Too Many Requests
			contract.WriteError(w, r, contract.NewRateLimitError("Rate limit exceeded"))
		})
	}
}

func consumeTokenNonAtomic(cfg *Config, key string, now time.Time) (bool, Bucket, error) {
	bucket, err := cfg.Store.GetBucket(key)
	if err != nil {
		return false, Bucket{}, err
	}
	if bucket == nil {
		bucket = &Bucket{
			Tokens:     cfg.Capacity,
			LastRefill: now,
			Capacity:   cfg.Capacity,
			RefillRate: cfg.RefillRate,
		}
	}

	elapsed := now.Sub(bucket.LastRefill).Seconds()
	newTokens := int64(elapsed * float64(bucket.RefillRate))
	if newTokens > 0 {
		bucket.Tokens = min(bucket.Capacity, bucket.Tokens+newTokens)
		bucket.LastRefill = now
	}

	allowed := bucket.Tokens >= 1
	if allowed {
		bucket.Tokens--
	}

	if err := cfg.Store.UpdateBucket(key, bucket); err != nil {
		return false, Bucket{}, err
	}

	return allowed, *bucket, nil
}

func (c *Config) withDefaults() *Config {
	config := *c

	if config.Capacity == 0 {
		config.Capacity = 100
	}

	if config.RefillRate == 0 {
		config.RefillRate = 10
	}

	if config.KeyFunc == nil {
		config.KeyFunc = IPKeyFunc
	}

	if config.Store == nil {
		config.Store = NewMemoryStore()
	}

	return &config
}

// ========================================
// Key Functions
// ========================================

// IPKeyFunc generates key from client IP
func IPKeyFunc(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return fmt.Sprintf("ip:%s", ip)
}

// UserKeyFunc generates key from user ID header
func UserKeyFunc(headerName string) KeyFunc {
	return func(r *http.Request) string {
		userID := r.Header.Get(headerName)
		if userID == "" {
			return "anonymous"
		}
		return fmt.Sprintf("user:%s", userID)
	}
}

// EndpointKeyFunc generates key from endpoint path
func EndpointKeyFunc(r *http.Request) string {
	return fmt.Sprintf("endpoint:%s", r.URL.Path)
}

// CompositeKeyFunc combines multiple key components
func CompositeKeyFunc(funcs ...KeyFunc) KeyFunc {
	return func(r *http.Request) string {
		h := fnv.New64a()
		for _, f := range funcs {
			h.Write([]byte(f(r)))
			h.Write([]byte("|"))
		}
		return fmt.Sprintf("%x", h.Sum64())
	}
}

// ========================================
// Memory Store (Single Instance)
// ========================================

// MemoryStore is an in-memory token bucket store
type MemoryStore struct {
	mu      sync.RWMutex
	buckets map[string]*Bucket
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		buckets: make(map[string]*Bucket),
	}
}

// GetBucket retrieves a token bucket
func (s *MemoryStore) GetBucket(key string) (*Bucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucket, exists := s.buckets[key]
	if !exists {
		return nil, nil
	}

	// Return a copy
	return &Bucket{
		Tokens:     bucket.Tokens,
		LastRefill: bucket.LastRefill,
		Capacity:   bucket.Capacity,
		RefillRate: bucket.RefillRate,
	}, nil
}

// UpdateBucket updates a token bucket
func (s *MemoryStore) UpdateBucket(key string, bucket *Bucket) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buckets[key] = &Bucket{
		Tokens:     bucket.Tokens,
		LastRefill: bucket.LastRefill,
		Capacity:   bucket.Capacity,
		RefillRate: bucket.RefillRate,
	}

	return nil
}

// DeleteBucket removes a token bucket
func (s *MemoryStore) DeleteBucket(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.buckets, key)
	return nil
}

// Cleanup removes old buckets (should be called periodically)
func (s *MemoryStore) Cleanup(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	removed := 0

	for key, bucket := range s.buckets {
		if now.Sub(bucket.LastRefill) > maxAge {
			delete(s.buckets, key)
			removed++
		}
	}

	return removed
}

// Count returns the number of buckets
func (s *MemoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.buckets)
}

// TakeToken atomically refills and consumes a single token for the key.
func (s *MemoryStore) TakeToken(key string, capacity, refillRate int64, now time.Time) (bool, Bucket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[key]
	if !exists {
		bucket = &Bucket{
			Tokens:     capacity,
			LastRefill: now,
			Capacity:   capacity,
			RefillRate: refillRate,
		}
		s.buckets[key] = bucket
	}

	elapsed := now.Sub(bucket.LastRefill).Seconds()
	newTokens := int64(elapsed * float64(bucket.RefillRate))
	if newTokens > 0 {
		bucket.Tokens = min(bucket.Capacity, bucket.Tokens+newTokens)
		bucket.LastRefill = now
	}

	if bucket.Tokens < 1 {
		return false, *bucket, nil
	}

	bucket.Tokens--
	return true, *bucket, nil
}
