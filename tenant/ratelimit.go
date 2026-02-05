package tenant

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// RateLimitConfig defines per-tenant rate limiting configuration.
// Zero values mean unlimited.
type RateLimitConfig struct {
	// RequestsPerSecond controls sustained rate.
	RequestsPerSecond int64
	// Burst controls the maximum burst capacity.
	// If zero, defaults to RequestsPerSecond.
	Burst int64
}

// RateLimitRequest is the input to rate limit checks.
type RateLimitRequest struct {
	// Tokens is the number of tokens to consume (default: 1).
	Tokens int64
	// Now overrides the current time (optional).
	Now time.Time
}

// RateLimitResult describes a rate limit decision.
type RateLimitResult struct {
	Allowed    bool
	Remaining  int64
	RetryAfter time.Duration
	Limit      int64
	Burst      int64
}

// RateLimitConfigProvider loads rate limit configuration for a tenant.
type RateLimitConfigProvider interface {
	RateLimitConfig(ctx context.Context, tenantID string) (RateLimitConfig, error)
}

// RateLimiter enforces per-tenant rate limits.
type RateLimiter interface {
	Allow(ctx context.Context, tenantID string, req RateLimitRequest) (RateLimitResult, error)
}

// InMemoryRateLimitProvider stores per-tenant rate limit configs in memory.
type InMemoryRateLimitProvider struct {
	mu      sync.RWMutex
	configs map[string]RateLimitConfig
}

// NewInMemoryRateLimitProvider creates an in-memory rate limit config provider.
func NewInMemoryRateLimitProvider() *InMemoryRateLimitProvider {
	return &InMemoryRateLimitProvider{
		configs: make(map[string]RateLimitConfig),
	}
}

// SetRateLimit sets rate limit config for a tenant.
func (p *InMemoryRateLimitProvider) SetRateLimit(tenantID string, cfg RateLimitConfig) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.configs[tenantID] = cfg
}

// RateLimitConfig returns the rate limit config for a tenant.
func (p *InMemoryRateLimitProvider) RateLimitConfig(ctx context.Context, tenantID string) (RateLimitConfig, error) {
	if p == nil {
		return RateLimitConfig{}, ErrTenantNotFound
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	cfg, ok := p.configs[tenantID]
	if !ok {
		return RateLimitConfig{}, ErrTenantNotFound
	}
	return cfg, nil
}

// TokenBucketRateLimiter implements per-tenant token bucket rate limiting.
type TokenBucketRateLimiter struct {
	provider RateLimitConfigProvider
	buckets  sync.Map // map[string]*tokenBucket
}

// NewTokenBucketRateLimiter creates a token bucket limiter from a config provider.
func NewTokenBucketRateLimiter(provider RateLimitConfigProvider) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{provider: provider}
}

// Allow checks rate limits for a tenant using token bucket algorithm.
func (l *TokenBucketRateLimiter) Allow(ctx context.Context, tenantID string, req RateLimitRequest) (RateLimitResult, error) {
	if l == nil || l.provider == nil {
		return RateLimitResult{Allowed: true}, nil
	}
	if tenantID == "" {
		return RateLimitResult{Allowed: true}, nil
	}

	cfg, err := l.provider.RateLimitConfig(ctx, tenantID)
	if err != nil {
		return RateLimitResult{Allowed: false}, err
	}

	if cfg.RequestsPerSecond <= 0 && cfg.Burst <= 0 {
		return RateLimitResult{Allowed: true}, nil
	}
	if cfg.RequestsPerSecond <= 0 {
		return RateLimitResult{Allowed: true}, nil
	}
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.RequestsPerSecond
	}

	if req.Tokens <= 0 {
		req.Tokens = 1
	}
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	}

	bucket := l.getBucket(tenantID)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if bucket.last.IsZero() {
		bucket.last = req.Now
		bucket.tokens = float64(cfg.Burst)
	}

	if bucket.capacity != cfg.Burst || bucket.refillRate != cfg.RequestsPerSecond {
		bucket.capacity = cfg.Burst
		bucket.refillRate = cfg.RequestsPerSecond
		if bucket.tokens > float64(bucket.capacity) {
			bucket.tokens = float64(bucket.capacity)
		}
	}

	elapsed := req.Now.Sub(bucket.last).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > 0 && bucket.refillRate > 0 {
		bucket.tokens = minFloat(float64(bucket.capacity), bucket.tokens+elapsed*float64(bucket.refillRate))
		bucket.last = req.Now
	}

	allowed := bucket.tokens >= float64(req.Tokens)
	if allowed {
		bucket.tokens -= float64(req.Tokens)
	}

	retryAfter := time.Duration(0)
	if !allowed && bucket.refillRate > 0 {
		needed := float64(req.Tokens) - bucket.tokens
		if needed < 0 {
			needed = 0
		}
		retryAfter = time.Duration((needed / float64(bucket.refillRate)) * float64(time.Second))
		if retryAfter < 0 {
			retryAfter = 0
		}
	}

	result := RateLimitResult{
		Allowed:    allowed,
		Remaining:  int64(bucket.tokens),
		RetryAfter: retryAfter,
		Limit:      cfg.RequestsPerSecond,
		Burst:      cfg.Burst,
	}

	if !allowed {
		return result, ErrRateLimitExceeded
	}
	return result, nil
}

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	last       time.Time
	capacity   int64
	refillRate int64
}

func (l *TokenBucketRateLimiter) getBucket(tenantID string) *tokenBucket {
	if bucket, ok := l.buckets.Load(tenantID); ok {
		return bucket.(*tokenBucket)
	}
	newBucket := &tokenBucket{}
	loaded, _ := l.buckets.LoadOrStore(tenantID, newBucket)
	return loaded.(*tokenBucket)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
