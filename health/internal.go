package health

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var errManagerClosed = errors.New("manager is closed")

// validateConfig validates the health check configuration.
func validateConfig(config HealthCheckConfig) error {
	if config.MaxHistoryEntries < 0 {
		return errors.New("max history entries cannot be negative")
	}
	if config.HistoryRetention < 0 {
		return errors.New("history retention cannot be negative")
	}
	if config.CleanupInterval < 0 {
		return errors.New("cleanup interval cannot be negative")
	}
	if config.Timeout < 0 {
		return errors.New("timeout cannot be negative")
	}
	if config.RetryCount < 0 {
		return errors.New("retry count cannot be negative")
	}
	if config.RetryDelay < 0 {
		return errors.New("retry delay cannot be negative")
	}
	return nil
}

// withCheckTimeout creates a context with timeout for health checks.
func withCheckTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		return ctx, func() {}
	}

	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining <= timeout {
			return ctx, func() {}
		}
	}

	return context.WithTimeout(ctx, timeout)
}

// runCheckWithRetry runs a health check with retry logic.
func runCheckWithRetry(ctx context.Context, checker ComponentChecker, config HealthCheckConfig) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	maxAttempts := config.RetryCount + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = checker.Check(ctx)
		if err == nil {
			return attempt, nil
		}
		if attempt >= maxAttempts {
			return attempt, err
		}
		if ctx != nil && ctx.Err() != nil {
			return attempt, err
		}
		if !sleepWithContext(ctx, config.RetryDelay) {
			return attempt, err
		}
	}

	return maxAttempts, err
}

// sleepWithContext sleeps for the specified duration, respecting context cancellation.
func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	if ctx == nil {
		time.Sleep(delay)
		return true
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// componentCheckResult represents the result of a single component health check.
type componentCheckResult struct {
	Name  string
	Error error
}

// requireManager checks that the manager is not nil and sends an error response
// if it is. Returns true if the manager is available, false if it is nil.
func requireManager(manager HealthManager, w http.ResponseWriter, r *http.Request) bool {
	if manager == nil {
		sendErrorResponse(w, r, http.StatusServiceUnavailable,
			"HEALTH_MANAGER_UNAVAILABLE", "Health manager is not configured", "")
		return false
	}
	return true
}

// httpStatusForHealth maps a HealthState to the appropriate HTTP status code.
func httpStatusForHealth(state HealthState) int {
	switch state {
	case StatusUnhealthy:
		return http.StatusServiceUnavailable
	case StatusDegraded:
		return http.StatusPartialContent
	default:
		return http.StatusOK
	}
}

// copyComponentHealth returns a shallow copy of ComponentHealth with a cloned Details map.
func copyComponentHealth(src *ComponentHealth) *ComponentHealth {
	dst := *src
	dst.HealthStatus.Details = make(map[string]any, len(src.HealthStatus.Details))
	for k, v := range src.HealthStatus.Details {
		dst.HealthStatus.Details[k] = v
	}
	return &dst
}

// calculateOverallStatus determines the aggregate HealthState and message from a set of component healths.
func calculateOverallStatus(healths map[string]*ComponentHealth) (HealthState, string) {
	allHealthy := true
	var messages []string

	for _, h := range healths {
		if h.Status != StatusHealthy {
			allHealthy = false
			if h.Message != "" {
				messages = append(messages, fmt.Sprintf("%s: %s", h.Status, h.Message))
			}
		}
	}

	if allHealthy {
		return StatusHealthy, ""
	}

	status := StatusDegraded
	for _, h := range healths {
		if h.Status == StatusUnhealthy {
			status = StatusUnhealthy
			break
		}
	}

	var message string
	if len(messages) > 0 {
		message = fmt.Sprintf("Issues detected: %v", messages)
	}
	return status, message
}
