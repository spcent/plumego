package circuitbreaker

import "context"

// CompatibilityAdapter exposes an x/ai/circuitbreaker.CircuitBreaker through
// the narrower breaker contract expected by x/ai/resilience during migration
// to shared x/resilience/circuitbreaker primitives.
//
// New x/ai/resilience callers should prefer shared x/resilience/circuitbreaker
// breakers directly. Keep this adapter only for backward compatibility with
// older AI-local circuit-breaker wiring.
type CompatibilityAdapter struct {
	inner *CircuitBreaker
}

// NewCompatibilityAdapter wraps a legacy AI-local circuit breaker so callers
// can keep using it through x/ai/resilience during migration to shared
// breakers. A nil input returns nil so config wiring can remain explicit.
func NewCompatibilityAdapter(inner *CircuitBreaker) *CompatibilityAdapter {
	if inner == nil {
		return nil
	}
	return &CompatibilityAdapter{inner: inner}
}

// Name returns the wrapped circuit breaker name.
func (a *CompatibilityAdapter) Name() string {
	if a == nil || a.inner == nil {
		return ""
	}
	return a.inner.Name()
}

// ExecuteWithContext delegates execution to the wrapped circuit breaker.
func (a *CompatibilityAdapter) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	if a == nil || a.inner == nil {
		return nil
	}
	return a.inner.ExecuteWithContext(ctx, fn)
}

// State returns the wrapped circuit breaker state.
func (a *CompatibilityAdapter) State() State {
	if a == nil || a.inner == nil {
		return StateClosed
	}
	return a.inner.State()
}

// Stats returns wrapped circuit breaker statistics.
func (a *CompatibilityAdapter) Stats() Stats {
	if a == nil || a.inner == nil {
		return Stats{State: StateClosed}
	}
	return a.inner.Stats()
}
