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
	EnableHistory       bool          `json:"enable_history"`
	MaxHistoryEntries   int           `json:"max_history_entries"`
	HistoryRetention    time.Duration `json:"history_retention"`
	AutoCleanupEnabled  bool          `json:"auto_cleanup_enabled"`
	CleanupInterval     time.Duration `json:"cleanup_interval"`
}

// BuildInfo describes the version metadata of the running binary.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

// These variables can be overridden at build time using -ldflags.
var (
	// Version reports the application version.
	Version = version
	// Commit reports the git commit hash used for the build.
	Commit = commit
	// BuildTime reports when the binary was built.
	BuildTime = buildTime
)

// GetBuildInfo returns the current build metadata.
func GetBuildInfo() BuildInfo {
	return BuildInfo{Version: Version, Commit: Commit, BuildTime: BuildTime}
}

// HealthManager defines the interface for health check management.
type HealthManager interface {
	RegisterComponent(checker ComponentChecker) error
	UnregisterComponent(name string) error
	GetComponentHealth(name string) (*ComponentHealth, bool)
	GetAllHealth() map[string]*ComponentHealth
	CheckComponent(ctx context.Context, name string) error
	CheckAllComponents(ctx context.Context) HealthStatus
	GetOverallHealth() HealthStatus
	GetHealthHistory() []HealthHistoryEntry
	QueryHealthHistory(query HealthHistoryQuery) HealthHistoryQueryResult
	GetHealthHistoryStats() map[string]any
	SetConfig(config HealthCheckConfig) error
	GetConfig() HealthCheckConfig
	ForceCleanup()
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
