package health

import (
	"context"
	"sync"
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

// HealthHistoryEntry represents a single entry in health state history.
type HealthHistoryEntry struct {
	Timestamp  time.Time    `json:"timestamp"`
	State      HealthState  `json:"state"`
	Message    string       `json:"message"`
	Components []string     `json:"components,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
}

// HealthManager manages health checks for all components.
type HealthManager struct {
	mu           sync.RWMutex
	components   map[string]ComponentChecker
	health       map[string]*ComponentHealth
	history      []HealthHistoryEntry
	readiness    ReadinessStatus
	lastCheck    time.Time
	config       HealthCheckConfig  // 数据保留策略配置
	cleanupTimer *time.Timer        // 清理定时器
}

// NewHealthManager creates a new HealthManager instance.
func NewHealthManager() *HealthManager {
	return &HealthManager{
		components: make(map[string]ComponentChecker),
		health:     make(map[string]*ComponentHealth),
		readiness:  ReadinessStatus{Ready: false, Reason: "starting"},
		config: HealthCheckConfig{
			MaxHistoryEntries:     100,  // 默认保留100条记录
			HistoryRetention:      24 * time.Hour,  // 默认保留24小时
			AutoCleanupEnabled:    true,  // 默认启用自动清理
			CleanupInterval:       1 * time.Hour,   // 默认每小时清理一次
		},
	}
}
