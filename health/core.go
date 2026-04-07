package health

import (
	"context"
	"time"
)

// HealthState represents an allowed component health status.
type HealthState string

const (
	StatusHealthy   HealthState = "healthy"
	StatusDegraded  HealthState = "degraded"
	StatusUnhealthy HealthState = "unhealthy"
)

// ComponentChecker defines the interface for health check components.
type ComponentChecker interface {
	Name() string                    // Return component name
	Check(ctx context.Context) error // Perform health check
}

// HealthStatus describes the health of a component in a structured format.
type HealthStatus struct {
	Status       HealthState    `json:"status"`
	Message      string         `json:"message,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
	Timestamp    time.Time      `json:"timestamp"`
	Duration     time.Duration  `json:"duration,omitempty"`
	Dependencies []string       `json:"dependencies,omitempty"`
}

// ComponentHealth represents the health status of a specific component.
type ComponentHealth struct {
	HealthStatus
	Enabled bool `json:"enabled"`
}

// HealthCheckConfig holds configuration for health checks.
type HealthCheckConfig struct {
	Enabled             bool          `json:"enabled"`
	Timeout             time.Duration `json:"timeout"`
	RetryCount          int           `json:"retry_count"`
	RetryDelay          time.Duration `json:"retry_delay"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// ComponentRegistry handles component registration and removal.
type ComponentRegistry interface {
	RegisterComponent(checker ComponentChecker) error
	UnregisterComponent(name string) error
}

// HealthChecker executes health checks and reads component status.
type HealthChecker interface {
	CheckComponent(ctx context.Context, name string) error
	CheckAllComponents(ctx context.Context) HealthStatus
	GetComponentHealth(name string) (*ComponentHealth, bool)
	GetAllHealth() map[string]*ComponentHealth
	GetOverallHealth() HealthStatus
	Readiness() ReadinessStatus
}

// HealthManager is the in-process management interface for health state and readiness.
type HealthManager interface {
	ComponentRegistry
	HealthChecker
	MarkReady()
	MarkNotReady(reason string)
	SetConfig(config HealthCheckConfig) error
	GetConfig() HealthCheckConfig
	Close() error
}

// isReady checks if the health status indicates the service is ready to serve traffic.
func (hs HealthState) isReady() bool {
	return hs == StatusHealthy || hs == StatusDegraded
}

// isValidHealthState checks whether a HealthState value is one of the known states.
func isValidHealthState(state HealthState) bool {
	switch state {
	case StatusHealthy, StatusDegraded, StatusUnhealthy:
		return true
	default:
		return false
	}
}
