// Package circuitbreaker implements the circuit breaker pattern for the AI Agent Gateway.
//
// Design Philosophy:
// - Prevent cascading failures when providers are unhealthy
// - Fast failure for known-bad backends
// - Automatic recovery testing (half-open state)
// - Thread-safe state transitions
// - Zero third-party dependencies
package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	// StateClosed means normal operation, requests pass through.
	StateClosed State = iota

	// StateOpen means circuit is broken, requests fail fast.
	StateOpen

	// StateHalfOpen means testing if backend has recovered.
	StateHalfOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern.
//
// The circuit breaker monitors failures and automatically opens
// when the failure threshold is exceeded, failing fast to prevent
// cascading failures.
type CircuitBreaker struct {
	name string

	// Configuration
	maxFailures    int           // Max failures before opening
	timeout        time.Duration // Request timeout
	resetTimeout   time.Duration // Time to wait before half-open

	// State
	state        State
	failures     int
	successes    int
	lastFailTime time.Time
	lastStateChange time.Time

	// Callbacks
	onStateChange func(from, to State)

	mu sync.RWMutex
}

// Config configures a circuit breaker.
type Config struct {
	Name           string
	MaxFailures    int
	Timeout        time.Duration
	ResetTimeout   time.Duration
	OnStateChange  func(from, to State)
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
		lastStateChange: time.Now(),
	}
}

// NewCircuitBreakerWithConfig creates a circuit breaker with custom config.
func NewCircuitBreakerWithConfig(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		name:          config.Name,
		maxFailures:   config.MaxFailures,
		timeout:       config.Timeout,
		resetTimeout:  config.ResetTimeout,
		state:         StateClosed,
		onStateChange: config.OnStateChange,
		lastStateChange: time.Now(),
	}
}

// Execute executes the given function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check state
	cb.mu.RLock()
	state := cb.state
	lastFailTime := cb.lastFailTime
	cb.mu.RUnlock()

	switch state {
	case StateOpen:
		// Check if should transition to half-open
		if time.Since(lastFailTime) > cb.resetTimeout {
			cb.transitionTo(StateHalfOpen)
			return cb.tryRequest(fn)
		}
		return ErrCircuitBreakerOpen

	case StateHalfOpen:
		return cb.tryRequest(fn)

	default: // StateClosed
		return cb.tryRequest(fn)
	}
}

// ExecuteWithContext executes the function with context support.
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	// Check state
	cb.mu.RLock()
	state := cb.state
	lastFailTime := cb.lastFailTime
	cb.mu.RUnlock()

	switch state {
	case StateOpen:
		// Check if should transition to half-open
		if time.Since(lastFailTime) > cb.resetTimeout {
			cb.transitionTo(StateHalfOpen)
			return cb.tryRequestWithContext(ctx, fn)
		}
		return ErrCircuitBreakerOpen

	case StateHalfOpen:
		return cb.tryRequestWithContext(ctx, fn)

	default: // StateClosed
		return cb.tryRequestWithContext(ctx, fn)
	}
}

// tryRequest attempts to execute the request and records the result.
func (cb *CircuitBreaker) tryRequest(fn func() error) error {
	// Apply timeout if configured
	if cb.timeout > 0 {
		done := make(chan error, 1)
		go func() {
			done <- fn()
		}()

		select {
		case err := <-done:
			if err != nil {
				cb.recordFailure()
				return err
			}
			cb.recordSuccess()
			return nil
		case <-time.After(cb.timeout):
			cb.recordFailure()
			return ErrTimeout
		}
	}

	// No timeout, execute directly
	err := fn()
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// tryRequestWithContext attempts to execute the request with context.
func (cb *CircuitBreaker) tryRequestWithContext(ctx context.Context, fn func(context.Context) error) error {
	err := fn(ctx)
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// recordSuccess records a successful request.
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0

	if cb.state == StateHalfOpen {
		cb.successes++
		// After a success in half-open, transition back to closed
		cb.transitionToLocked(StateClosed)
	}
}

// recordFailure records a failed request.
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.transitionToLocked(StateOpen)
		}
	case StateHalfOpen:
		// Failure in half-open immediately reopens circuit
		cb.transitionToLocked(StateOpen)
	}
}

// transitionTo transitions to a new state (external call).
func (cb *CircuitBreaker) transitionTo(newState State) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionToLocked(newState)
}

// transitionToLocked transitions to a new state (assumes lock is held).
func (cb *CircuitBreaker) transitionToLocked(newState State) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	// Reset counters on state change
	if newState == StateClosed {
		cb.failures = 0
		cb.successes = 0
	} else if newState == StateHalfOpen {
		cb.successes = 0
	}

	// Call callback if configured
	if cb.onStateChange != nil {
		go cb.onStateChange(oldState, newState)
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Name returns the name of the circuit breaker.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Stats returns statistics about the circuit breaker.
type Stats struct {
	Name            string
	State           State
	Failures        int
	Successes       int
	LastFailTime    time.Time
	LastStateChange time.Time
}

// Stats returns circuit breaker statistics.
func (cb *CircuitBreaker) Stats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Stats{
		Name:            cb.name,
		State:           cb.state,
		Failures:        cb.failures,
		Successes:       cb.successes,
		LastFailTime:    cb.lastFailTime,
		LastStateChange: cb.lastStateChange,
	}
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.lastStateChange = time.Now()
}

// Error types
var (
	ErrCircuitBreakerOpen = fmt.Errorf("circuit breaker is open")
	ErrTimeout            = fmt.Errorf("request timeout")
)

// NoOpCircuitBreaker is a circuit breaker that always allows requests (for testing/disabled circuit breaking).
type NoOpCircuitBreaker struct{}

// Execute implements the circuit breaker interface.
func (n *NoOpCircuitBreaker) Execute(fn func() error) error {
	return fn()
}

// ExecuteWithContext implements the circuit breaker interface.
func (n *NoOpCircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// State returns always closed.
func (n *NoOpCircuitBreaker) State() State {
	return StateClosed
}

// Stats returns empty stats.
func (n *NoOpCircuitBreaker) Stats() Stats {
	return Stats{
		Name:  "noop",
		State: StateClosed,
	}
}

// Reset does nothing.
func (n *NoOpCircuitBreaker) Reset() {}
