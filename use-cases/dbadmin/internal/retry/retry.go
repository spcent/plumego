package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// Config defines retry behavior.
type Config struct {
	MaxAttempts  int           // Maximum number of attempts (0 = no retry, 1 = one retry)
	BaseDelay    time.Duration // Initial delay between retries
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Exponential backoff multiplier (typically 2.0)
	JitterFactor float64       // Random jitter factor (0.0-1.0, typically 0.1-0.2)
}

// DefaultConfig returns sensible defaults for connection retries.
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		BaseDelay:    100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	}
}

// Do executes fn with retry logic according to config.
// Returns the last error if all attempts fail.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	if cfg.MaxAttempts <= 0 {
		return fn()
	}

	var lastErr error
	for attempt := 0; attempt <= cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff and jitter
			delay := calculateDelay(cfg, attempt)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Don't retry if context is cancelled
		if ctx.Err() != nil {
			return lastErr
		}

		// Don't retry non-transient errors
		if !IsTransient(lastErr) {
			return lastErr
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", cfg.MaxAttempts+1, lastErr)
}

// calculateDelay computes delay with exponential backoff and jitter.
func calculateDelay(cfg Config, attempt int) time.Duration {
	// Exponential backoff
	delay := float64(cfg.BaseDelay) * math.Pow(cfg.Multiplier, float64(attempt-1))

	// Apply jitter
	if cfg.JitterFactor > 0 {
		jitter := delay * cfg.JitterFactor
		delay += (rand.Float64()*2 - 1) * jitter // Random value in [-jitter, +jitter]
	}

	// Clamp to max delay
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}

	return time.Duration(delay)
}

// IsTransient checks if an error is likely transient and worth retrying.
func IsTransient(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network errors
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"connection timed out",
		"no route to host",
		"network is unreachable",
		"timeout",
		"deadline exceeded",
		"temporary failure",
		"server busy",
		"too many connections",
		"max connections reached",
		"resource temporarily unavailable",
		"broken pipe",
		"EOF",
	}

	lowerErr := strings.ToLower(errStr)
	for _, pattern := range transientPatterns {
		if strings.Contains(lowerErr, pattern) {
			return true
		}
	}

	// Check for wrapped transient errors
	var timeoutErr interface {
		Timeout() bool
	}
	if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
		return true
	}

	var tempErr interface {
		Temporary() bool
	}
	if errors.As(err, &tempErr) && tempErr.Temporary() {
		return true
	}

	return false
}

// WithResult executes fn with retry and returns both result and error.
func WithResult[T any](ctx context.Context, cfg Config, fn func() (T, error)) (T, error) {
	var zero T
	if cfg.MaxAttempts <= 0 {
		return fn()
	}

	var lastErr error
	for attempt := 0; attempt <= cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateDelay(cfg, attempt)

			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
				// Continue
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if ctx.Err() != nil {
			return zero, lastErr
		}

		if !IsTransient(lastErr) {
			return zero, lastErr
		}
	}

	return zero, fmt.Errorf("failed after %d attempts: %w", cfg.MaxAttempts+1, lastErr)
}
