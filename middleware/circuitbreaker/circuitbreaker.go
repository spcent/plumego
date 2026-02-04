// Package circuitbreaker provides circuit breaker pattern implementation
//
// The circuit breaker prevents cascading failures by detecting when a service
// is failing and temporarily blocking requests to give it time to recover.
//
// States:
//   - Closed: Normal operation, requests pass through
//   - Open: Service is failing, requests fail fast
//   - Half-Open: Testing if service recovered, limited requests allowed
//
// State transitions:
//   - Closed -> Open: Failure rate exceeds threshold
//   - Open -> Half-Open: After timeout period
//   - Half-Open -> Closed: Success rate meets threshold
//   - Half-Open -> Open: Failures continue
//
// Example:
//
//	cb := circuitbreaker.New(circuitbreaker.Config{
//		Name:             "user-service",
//		FailureThreshold: 0.5,  // Open after 50% failures
//		SuccessThreshold: 3,    // Need 3 successes to close
//		Timeout:          30 * time.Second,
//		MinRequests:      10,   // Need 10 requests before evaluating
//	})
//
//	err := cb.Call(func() error {
//		return callBackend()
//	})
package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// State represents the circuit breaker state
type State int32

const (
	// StateClosed means requests are passing through normally
	StateClosed State = iota

	// StateOpen means the circuit is open and requests fail fast
	StateOpen

	// StateHalfOpen means the circuit is testing if the service recovered
	StateHalfOpen
)

// String returns the state name
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

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name   string
	config Config

	// State management
	state         atomic.Int32
	stateChanged  time.Time
	stateMu       sync.RWMutex

	// Metrics (closed state)
	requests atomic.Uint64
	failures atomic.Uint64
	successes atomic.Uint64

	// Metrics (half-open state)
	halfOpenRequests  atomic.Uint64
	halfOpenSuccesses atomic.Uint64

	// Hooks
	onStateChange func(from, to State)

	// Testing
	clock Clock
}

// Config holds circuit breaker configuration
type Config struct {
	// Name is the circuit breaker name (for logging/metrics)
	Name string

	// FailureThreshold is the failure rate threshold (0.0-1.0)
	// When exceeded, the circuit opens
	// Default: 0.5 (50%)
	FailureThreshold float64

	// SuccessThreshold is the number of consecutive successes
	// needed in half-open state to close the circuit
	// Default: 3
	SuccessThreshold uint64

	// Timeout is how long the circuit stays open before
	// transitioning to half-open
	// Default: 30 seconds
	Timeout time.Duration

	// MinRequests is the minimum number of requests needed
	// before calculating failure rate
	// Default: 10
	MinRequests uint64

	// MaxRequests is the maximum number of concurrent requests
	// allowed in half-open state
	// Default: 1
	MaxRequests uint64

	// OnStateChange is called when the state changes
	OnStateChange func(from, to State)

	// ShouldTrip is a custom function to determine if circuit should open
	// If nil, uses default failure rate calculation
	ShouldTrip func(counts Counts) bool
}

// Counts holds circuit breaker metrics
type Counts struct {
	Requests  uint64
	Successes uint64
	Failures  uint64

	// Consecutive counts (for half-open state)
	ConsecutiveSuccesses uint64
	ConsecutiveFailures  uint64
}

// FailureRate returns the failure rate (0.0-1.0)
func (c Counts) FailureRate() float64 {
	if c.Requests == 0 {
		return 0.0
	}
	return float64(c.Failures) / float64(c.Requests)
}

// SuccessRate returns the success rate (0.0-1.0)
func (c Counts) SuccessRate() float64 {
	if c.Requests == 0 {
		return 0.0
	}
	return float64(c.Successes) / float64(c.Requests)
}

// New creates a new circuit breaker
func New(config Config) *CircuitBreaker {
	// Apply defaults
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 0.5
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MinRequests == 0 {
		config.MinRequests = 10
	}
	if config.MaxRequests == 0 {
		config.MaxRequests = 1
	}
	if config.Name == "" {
		config.Name = "circuit-breaker"
	}

	cb := &CircuitBreaker{
		name:          config.Name,
		config:        config,
		stateChanged:  time.Now(),
		onStateChange: config.OnStateChange,
		clock:         &realClock{},
	}

	cb.state.Store(int32(StateClosed))

	return cb
}

// Call executes the function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	// Check if we can proceed
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute function
	err := fn()

	// Record result
	cb.afterRequest(err)

	return err
}

// CallWithContext executes the function with context
func (cb *CircuitBreaker) CallWithContext(ctx context.Context, fn func() error) error {
	// Check context first
	if err := ctx.Err(); err != nil {
		return err
	}

	return cb.Call(fn)
}

// beforeRequest checks if the request can proceed
func (cb *CircuitBreaker) beforeRequest() error {
	state := cb.State()

	switch state {
	case StateClosed:
		// Allow request
		return nil

	case StateOpen:
		// Check if we should transition to half-open
		cb.stateMu.RLock()
		elapsed := cb.clock.Since(cb.stateChanged)
		cb.stateMu.RUnlock()

		if elapsed >= cb.config.Timeout {
			// Transition to half-open
			cb.setState(StateHalfOpen)
			return nil
		}

		// Circuit is open, fail fast
		return ErrCircuitOpen

	case StateHalfOpen:
		// Check if we've reached max concurrent requests
		current := cb.halfOpenRequests.Load()
		if current >= cb.config.MaxRequests {
			return ErrTooManyRequests
		}

		// Allow request
		return nil

	default:
		return ErrCircuitOpen
	}
}

// afterRequest records the request result
func (cb *CircuitBreaker) afterRequest(err error) {
	state := cb.State()

	switch state {
	case StateClosed:
		cb.requests.Add(1)

		if err != nil {
			cb.failures.Add(1)
			cb.checkShouldOpen()
		} else {
			cb.successes.Add(1)
		}

	case StateHalfOpen:
		cb.halfOpenRequests.Add(1)

		if err != nil {
			// Failure in half-open, reopen circuit
			cb.setState(StateOpen)
		} else {
			cb.halfOpenSuccesses.Add(1)

			// Check if we've met success threshold
			if cb.halfOpenSuccesses.Load() >= cb.config.SuccessThreshold {
				cb.setState(StateClosed)
			}
		}
	}
}

// checkShouldOpen checks if the circuit should open
func (cb *CircuitBreaker) checkShouldOpen() {
	requests := cb.requests.Load()
	if requests < cb.config.MinRequests {
		return
	}

	counts := cb.Counts()

	// Use custom trip function if provided
	if cb.config.ShouldTrip != nil {
		if cb.config.ShouldTrip(counts) {
			cb.setState(StateOpen)
		}
		return
	}

	// Default: check failure rate
	if counts.FailureRate() >= cb.config.FailureThreshold {
		cb.setState(StateOpen)
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState State) {
	cb.stateMu.Lock()
	defer cb.stateMu.Unlock()

	oldState := State(cb.state.Load())
	if oldState == newState {
		return
	}

	// Update state
	cb.state.Store(int32(newState))
	cb.stateChanged = cb.clock.Now()

	// Reset metrics based on new state
	switch newState {
	case StateClosed:
		cb.requests.Store(0)
		cb.failures.Store(0)
		cb.successes.Store(0)

	case StateOpen:
		// Keep metrics for analysis

	case StateHalfOpen:
		cb.halfOpenRequests.Store(0)
		cb.halfOpenSuccesses.Store(0)
	}

	// Call hook
	if cb.onStateChange != nil {
		go cb.onStateChange(oldState, newState)
	}
}

// State returns the current state
func (cb *CircuitBreaker) State() State {
	return State(cb.state.Load())
}

// Counts returns current metrics
func (cb *CircuitBreaker) Counts() Counts {
	return Counts{
		Requests:             cb.requests.Load(),
		Successes:            cb.successes.Load(),
		Failures:             cb.failures.Load(),
		ConsecutiveSuccesses: cb.halfOpenSuccesses.Load(),
		ConsecutiveFailures:  0, // Not tracked currently
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.setState(StateClosed)
}

// Trip manually opens the circuit
func (cb *CircuitBreaker) Trip() {
	cb.setState(StateOpen)
}

// Name returns the circuit breaker name
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() Stats {
	cb.stateMu.RLock()
	defer cb.stateMu.RUnlock()

	return Stats{
		Name:          cb.name,
		State:         cb.State(),
		Counts:        cb.Counts(),
		StateChanged:  cb.stateChanged,
		StateDuration: cb.clock.Since(cb.stateChanged),
	}
}

// Stats holds circuit breaker statistics
type Stats struct {
	Name          string
	State         State
	Counts        Counts
	StateChanged  time.Time
	StateDuration time.Duration
}

// String returns a string representation
func (s Stats) String() string {
	return fmt.Sprintf("%s [%s] requests=%d successes=%d failures=%d rate=%.2f%%",
		s.Name, s.State, s.Counts.Requests, s.Counts.Successes, s.Counts.Failures,
		s.Counts.FailureRate()*100)
}

// Common errors
var (
	// ErrCircuitOpen is returned when the circuit is open
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrTooManyRequests is returned when too many requests in half-open state
	ErrTooManyRequests = errors.New("too many requests")
)

// Clock interface for time operations (allows testing)
type Clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type realClock struct{}

func (c *realClock) Now() time.Time {
	return time.Now()
}

func (c *realClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}
