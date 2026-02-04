package proxy

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrNoBackends is returned when no backends are configured
	ErrNoBackends = errors.New("no backends configured")

	// ErrNoHealthyBackends is returned when all backends are unhealthy
	ErrNoHealthyBackends = errors.New("no healthy backends available")

	// ErrBackendTimeout is returned when a backend times out
	ErrBackendTimeout = errors.New("backend timeout")

	// ErrBackendRefused is returned when a backend refuses connection
	ErrBackendRefused = errors.New("backend refused connection")

	// ErrWebSocketUpgrade is returned when WebSocket upgrade fails
	ErrWebSocketUpgrade = errors.New("websocket upgrade failed")

	// ErrInvalidConfig is returned when proxy configuration is invalid
	ErrInvalidConfig = errors.New("invalid proxy configuration")

	// ErrTooManyRetries is returned when all retries are exhausted
	ErrTooManyRetries = errors.New("too many retries")
)

// ProxyError wraps an error with additional context
type ProxyError struct {
	Backend string
	Err     error
	Attempt int
}

// NewProxyError creates a new proxy error
func NewProxyError(backend string, err error, attempt int) *ProxyError {
	return &ProxyError{
		Backend: backend,
		Err:     err,
		Attempt: attempt,
	}
}

// Error implements the error interface
func (e *ProxyError) Error() string {
	return fmt.Sprintf("proxy error (backend=%s, attempt=%d): %v", 
		e.Backend, e.Attempt, e.Err)
}

// Unwrap returns the wrapped error
func (e *ProxyError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target
func (e *ProxyError) Is(target error) bool {
	return errors.Is(e.Err, target)
}
