package ratelimit

import "context"

// CompatibilityAdapter exposes an x/ai/ratelimit.RateLimiter through the
// narrower rate-limit contract expected by x/ai/resilience during migration to
// shared x/resilience/ratelimit primitives.
//
// New x/ai/resilience callers should prefer shared x/resilience/ratelimit
// keyed buckets directly. Keep this adapter only for backward compatibility
// with older AI-local limiter wiring.
type CompatibilityAdapter struct {
	inner RateLimiter
}

// NewCompatibilityAdapter wraps a legacy AI-local rate limiter so callers can
// keep using it through x/ai/resilience during migration to shared keyed
// buckets. A nil input returns nil so config wiring can remain explicit.
func NewCompatibilityAdapter(inner RateLimiter) *CompatibilityAdapter {
	if inner == nil {
		return nil
	}
	return &CompatibilityAdapter{inner: inner}
}

// Allow reports whether the wrapped limiter allows the key.
func (a *CompatibilityAdapter) Allow(ctx context.Context, key string) (bool, error) {
	if a == nil || a.inner == nil {
		return false, nil
	}
	return a.inner.Allow(ctx, key)
}

// Remaining reports remaining tokens for the key through the wrapped limiter.
func (a *CompatibilityAdapter) Remaining(ctx context.Context, key string) (int, error) {
	if a == nil || a.inner == nil {
		return 0, nil
	}
	return a.inner.Remaining(ctx, key)
}
