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

// IsReady reports whether the health state can serve traffic.
func (hs HealthState) IsReady() bool {
	return hs == StatusHealthy || hs == StatusDegraded
}
