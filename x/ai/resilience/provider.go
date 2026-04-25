// Package resilience provides resilience wrappers for AI providers.
//
// This package combines rate limiting and circuit breaking to create
// resilient providers that can handle failures gracefully.
package resilience

import (
	"context"
	"errors"
	"fmt"

	"github.com/spcent/plumego/x/ai/circuitbreaker"
	"github.com/spcent/plumego/x/ai/provider"
	"github.com/spcent/plumego/x/ai/ratelimit"
)

var (
	// ErrNilProvider is returned when a resilient provider is built without an
	// underlying provider.
	ErrNilProvider = errors.New("ai resilience: provider cannot be nil")

	// ErrNilRequest is returned when a provider request is nil.
	ErrNilRequest = errors.New("ai resilience: completion request cannot be nil")
)

// ResilientProvider wraps a provider with rate limiting and circuit breaking.
type ResilientProvider struct {
	provider       provider.Provider
	rateLimiter    ratelimit.RateLimiter
	circuitBreaker *circuitbreaker.CircuitBreaker
	name           string
}

// Config configures a resilient provider.
type Config struct {
	Provider       provider.Provider
	RateLimiter    ratelimit.RateLimiter
	CircuitBreaker *circuitbreaker.CircuitBreaker
}

// NewResilientProvider creates a new resilient provider.
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

	name := config.Provider.Name()
	if config.CircuitBreaker != nil {
		name = config.CircuitBreaker.Name()
	}

	return &ResilientProvider{
		provider:       config.Provider,
		rateLimiter:    config.RateLimiter,
		circuitBreaker: config.CircuitBreaker,
		name:           name,
	}, nil
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
			return nil, ratelimit.ErrRateLimitExceeded
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
			return nil, ratelimit.ErrRateLimitExceeded
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
func (rp *ResilientProvider) CircuitBreakerState() circuitbreaker.State {
	if rp.circuitBreaker != nil {
		return rp.circuitBreaker.State()
	}
	return circuitbreaker.StateClosed
}

// CircuitBreakerStats returns circuit breaker statistics
func (rp *ResilientProvider) CircuitBreakerStats() circuitbreaker.Stats {
	if rp.circuitBreaker != nil {
		return rp.circuitBreaker.Stats()
	}
	return circuitbreaker.Stats{
		Name:  rp.name,
		State: circuitbreaker.StateClosed,
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
