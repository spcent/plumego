// Package resilience provides resilience wrappers for AI providers.
//
// This package combines rate limiting and circuit breaking to create
// resilient providers that can handle failures gracefully.
package resilience

import (
	"context"
	"errors"
	"fmt"
	"time"

	aicircuitbreaker "github.com/spcent/plumego/x/ai/circuitbreaker"
	"github.com/spcent/plumego/x/ai/provider"
	airatelimit "github.com/spcent/plumego/x/ai/ratelimit"
	sharedcircuitbreaker "github.com/spcent/plumego/x/resilience/circuitbreaker"
	sharedratelimit "github.com/spcent/plumego/x/resilience/ratelimit"
)

var (
	// ErrNilProvider is returned when a resilient provider is built without an
	// underlying provider.
	ErrNilProvider = errors.New("ai resilience: provider cannot be nil")

	// ErrNilRequest is returned when a provider request is nil.
	ErrNilRequest = errors.New("ai resilience: completion request cannot be nil")

	// ErrMultipleRateLimiters is returned when both canonical shared and
	// compatibility rate limiters are configured at once.
	ErrMultipleRateLimiters = errors.New("ai resilience: configure either RateLimiter or LegacyRateLimiter, not both")

	// ErrMultipleCircuitBreakers is returned when both canonical shared and
	// compatibility circuit breakers are configured at once.
	ErrMultipleCircuitBreakers = errors.New("ai resilience: configure either CircuitBreaker or LegacyCircuitBreaker, not both")
)

// ResilientProvider wraps a provider with rate limiting and circuit breaking.
type ResilientProvider struct {
	provider       provider.Provider
	rateLimiter    rateLimiter
	circuitBreaker circuitBreaker
	name           string
}

// Config configures a resilient provider.
type Config struct {
	Provider             provider.Provider
	RateLimiter          *sharedratelimit.KeyedBuckets
	LegacyRateLimiter    *airatelimit.CompatibilityAdapter
	CircuitBreaker       *sharedcircuitbreaker.CircuitBreaker
	LegacyCircuitBreaker *aicircuitbreaker.CompatibilityAdapter
}

type rateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
	Remaining(ctx context.Context, key string) (int, error)
}

type circuitBreaker interface {
	Name() string
	ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error
	State() aicircuitbreaker.State
	Stats() aicircuitbreaker.Stats
}

// NewResilientProvider creates a new resilient provider.
//
// Deprecated: use NewResilientProviderE so invalid resilience composition
// returns an error instead of panicking.
func NewResilientProvider(config Config) *ResilientProvider {
	resilient, err := NewResilientProviderE(config)
	if err != nil {
		panic(err)
	}
	return resilient
}

// NewResilientProviderE creates a new resilient provider and reports invalid
// composition through an error instead of panicking.
func NewResilientProviderE(config Config) (*ResilientProvider, error) {
	if config.Provider == nil {
		return nil, ErrNilProvider
	}

	limiter, err := resolveRateLimiter(config)
	if err != nil {
		return nil, err
	}
	breaker, err := resolveCircuitBreaker(config)
	if err != nil {
		return nil, err
	}

	name := config.Provider.Name()
	if breaker != nil {
		name = breaker.Name()
	}

	return &ResilientProvider{
		provider:       config.Provider,
		rateLimiter:    limiter,
		circuitBreaker: breaker,
		name:           name,
	}, nil
}

func resolveRateLimiter(config Config) (rateLimiter, error) {
	if config.RateLimiter != nil && config.LegacyRateLimiter != nil {
		return nil, ErrMultipleRateLimiters
	}
	if config.RateLimiter != nil {
		return sharedRateLimiterAdapter{inner: config.RateLimiter}, nil
	}
	if config.LegacyRateLimiter != nil {
		return config.LegacyRateLimiter, nil
	}
	return nil, nil
}

func resolveCircuitBreaker(config Config) (circuitBreaker, error) {
	if config.CircuitBreaker != nil && config.LegacyCircuitBreaker != nil {
		return nil, ErrMultipleCircuitBreakers
	}
	if config.CircuitBreaker != nil {
		return sharedCircuitBreakerAdapter{inner: config.CircuitBreaker}, nil
	}
	if config.LegacyCircuitBreaker != nil {
		return config.LegacyCircuitBreaker, nil
	}
	return nil, nil
}

// Name implements provider.Provider
func (rp *ResilientProvider) Name() string {
	if rp == nil || rp.provider == nil {
		return ""
	}
	return rp.provider.Name()
}

// Complete implements provider.Provider with rate limiting and circuit breaking
func (rp *ResilientProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	if err := rp.validateRequest(req); err != nil {
		return nil, err
	}

	// Apply rate limiting first
	if rp.rateLimiter != nil {
		allowed, err := rp.rateLimiter.Allow(ctx, rp.getRateLimitKey(req))
		if err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}
		if !allowed {
			return nil, airatelimit.ErrRateLimitExceeded
		}
	}

	// Apply circuit breaking
	if rp.circuitBreaker != nil {
		var resp *provider.CompletionResponse
		var err error

		breakerErr := rp.circuitBreaker.ExecuteWithContext(ctx, func(ctx context.Context) error {
			resp, err = rp.provider.Complete(ctx, req)
			return err
		})

		if breakerErr != nil {
			return nil, breakerErr
		}

		return resp, err
	}

	// No circuit breaker, call directly
	return rp.provider.Complete(ctx, req)
}

// CompleteStream implements provider.Provider with rate limiting and circuit breaking
func (rp *ResilientProvider) CompleteStream(ctx context.Context, req *provider.CompletionRequest) (*provider.StreamReader, error) {
	if err := rp.validateRequest(req); err != nil {
		return nil, err
	}

	// Apply rate limiting first
	if rp.rateLimiter != nil {
		allowed, err := rp.rateLimiter.Allow(ctx, rp.getRateLimitKey(req))
		if err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}
		if !allowed {
			return nil, airatelimit.ErrRateLimitExceeded
		}
	}

	// Apply circuit breaking
	if rp.circuitBreaker != nil {
		var reader *provider.StreamReader
		var err error

		breakerErr := rp.circuitBreaker.ExecuteWithContext(ctx, func(ctx context.Context) error {
			reader, err = rp.provider.CompleteStream(ctx, req)
			return err
		})

		if breakerErr != nil {
			return nil, breakerErr
		}

		return reader, err
	}

	// No circuit breaker, call directly
	return rp.provider.CompleteStream(ctx, req)
}

// ListModels implements provider.Provider
func (rp *ResilientProvider) ListModels(ctx context.Context) ([]provider.Model, error) {
	if rp == nil || rp.provider == nil {
		return nil, ErrNilProvider
	}

	// Apply circuit breaking for list models
	if rp.circuitBreaker != nil {
		var models []provider.Model
		var err error

		breakerErr := rp.circuitBreaker.ExecuteWithContext(ctx, func(ctx context.Context) error {
			models, err = rp.provider.ListModels(ctx)
			return err
		})

		if breakerErr != nil {
			return nil, breakerErr
		}

		return models, err
	}

	return rp.provider.ListModels(ctx)
}

// GetModel implements provider.Provider
func (rp *ResilientProvider) GetModel(ctx context.Context, modelID string) (*provider.Model, error) {
	if rp == nil || rp.provider == nil {
		return nil, ErrNilProvider
	}

	// Apply circuit breaking for get model
	if rp.circuitBreaker != nil {
		var model *provider.Model
		var err error

		breakerErr := rp.circuitBreaker.ExecuteWithContext(ctx, func(ctx context.Context) error {
			model, err = rp.provider.GetModel(ctx, modelID)
			return err
		})

		if breakerErr != nil {
			return nil, breakerErr
		}

		return model, err
	}

	return rp.provider.GetModel(ctx, modelID)
}

// CountTokens implements provider.Provider
func (rp *ResilientProvider) CountTokens(text string) (int, error) {
	if rp == nil || rp.provider == nil {
		return 0, ErrNilProvider
	}

	// Token counting doesn't need circuit breaking or rate limiting
	return rp.provider.CountTokens(text)
}

func (rp *ResilientProvider) validateRequest(req *provider.CompletionRequest) error {
	if rp == nil || rp.provider == nil {
		return ErrNilProvider
	}
	if req == nil {
		return ErrNilRequest
	}
	return nil
}

// getRateLimitKey returns the rate limit key for a request
func (rp *ResilientProvider) getRateLimitKey(req *provider.CompletionRequest) string {
	// Use provider:model as the rate limit key
	return fmt.Sprintf("%s:%s", rp.provider.Name(), req.Model)
}

// CircuitBreakerState returns the current circuit breaker state
func (rp *ResilientProvider) CircuitBreakerState() aicircuitbreaker.State {
	if rp.circuitBreaker != nil {
		return rp.circuitBreaker.State()
	}
	return aicircuitbreaker.StateClosed
}

// CircuitBreakerStats returns circuit breaker statistics
func (rp *ResilientProvider) CircuitBreakerStats() aicircuitbreaker.Stats {
	if rp.circuitBreaker != nil {
		return rp.circuitBreaker.Stats()
	}
	return aicircuitbreaker.Stats{
		Name:  rp.name,
		State: aicircuitbreaker.StateClosed,
	}
}

// RateLimitRemaining returns remaining rate limit tokens
func (rp *ResilientProvider) RateLimitRemaining(ctx context.Context, model string) (int, error) {
	if rp == nil || rp.provider == nil {
		return 0, ErrNilProvider
	}
	if rp.rateLimiter != nil {
		key := fmt.Sprintf("%s:%s", rp.provider.Name(), model)
		return rp.rateLimiter.Remaining(ctx, key)
	}
	return -1, nil // No limit
}

type sharedRateLimiterAdapter struct {
	inner *sharedratelimit.KeyedBuckets
}

func (a sharedRateLimiterAdapter) Allow(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return a.inner.Allow(key), nil
}

func (a sharedRateLimiterAdapter) Remaining(ctx context.Context, key string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return a.inner.Remaining(key), nil
}

type sharedCircuitBreakerAdapter struct {
	inner *sharedcircuitbreaker.CircuitBreaker
}

func (a sharedCircuitBreakerAdapter) Name() string {
	return a.inner.Name()
}

func (a sharedCircuitBreakerAdapter) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	err := a.inner.CallWithContext(ctx, func() error {
		return fn(ctx)
	})
	if errors.Is(err, sharedcircuitbreaker.ErrCircuitOpen) {
		return aicircuitbreaker.ErrCircuitBreakerOpen
	}
	return err
}

func (a sharedCircuitBreakerAdapter) State() aicircuitbreaker.State {
	return toAICircuitState(a.inner.State())
}

func (a sharedCircuitBreakerAdapter) Stats() aicircuitbreaker.Stats {
	stats := a.inner.Stats()
	lastFailTime := time.Time{}
	if stats.State == sharedcircuitbreaker.StateOpen {
		lastFailTime = stats.StateChanged
	}
	return aicircuitbreaker.Stats{
		Name:            stats.Name,
		State:           toAICircuitState(stats.State),
		Failures:        int(stats.Counts.Failures),
		Successes:       int(stats.Counts.Successes),
		LastFailTime:    lastFailTime,
		LastStateChange: stats.StateChanged,
	}
}

func toAICircuitState(state sharedcircuitbreaker.State) aicircuitbreaker.State {
	switch state {
	case sharedcircuitbreaker.StateClosed:
		return aicircuitbreaker.StateClosed
	case sharedcircuitbreaker.StateOpen:
		return aicircuitbreaker.StateOpen
	case sharedcircuitbreaker.StateHalfOpen:
		return aicircuitbreaker.StateHalfOpen
	default:
		return aicircuitbreaker.StateClosed
	}
}
