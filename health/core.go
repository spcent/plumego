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

// ComponentChecker defines a transport-agnostic component health check.
type ComponentChecker interface {
	// Name returns the stable component name used by the owning orchestrator.
	Name() string
	// Check reports nil when the component is healthy.
	Check(ctx context.Context) error
}

// HealthStatus describes component or aggregate health in a structured format.
type HealthStatus struct {
	Status       HealthState    `json:"status"`
	Message      string         `json:"message,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
	Timestamp    time.Time      `json:"timestamp"`
	Duration     time.Duration  `json:"duration,omitempty"`
	Dependencies []string       `json:"dependencies,omitempty"`
}

// ComponentHealth describes a specific component and whether it is enabled.
type ComponentHealth struct {
	HealthStatus
	Enabled bool `json:"enabled"`
}

// IsReady reports whether the health state can serve traffic.
func (hs HealthState) IsReady() bool {
	return hs == StatusHealthy || hs == StatusDegraded
}

// IsKnown reports whether the health state is one of the stable values.
func (hs HealthState) IsKnown() bool {
	switch hs {
	case StatusHealthy, StatusDegraded, StatusUnhealthy:
		return true
	default:
		return false
	}
}
