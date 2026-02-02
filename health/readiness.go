package health

import (
	"sync"
	"time"
)

// ReadinessStatus describes whether the application is ready to serve traffic.
type ReadinessStatus struct {
	Ready      bool            `json:"ready"`
	Reason     string          `json:"reason,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
	Components map[string]bool `json:"components,omitempty"`
}

// Global readiness management (for backward compatibility with external packages)
var (
	readinessMu     sync.RWMutex
	globalReadiness ReadinessStatus
)

func init() {
	globalReadiness = ReadinessStatus{
		Ready:     false,
		Reason:    "starting",
		Timestamp: time.Now(),
	}
}

// updateReadiness updates the global readiness status.
func updateReadiness(status ReadinessStatus) {
	readinessMu.Lock()
	defer readinessMu.Unlock()
	globalReadiness = status
}

// GetReadiness returns the current readiness status.
func GetReadiness() ReadinessStatus {
	readinessMu.RLock()
	defer readinessMu.RUnlock()
	return globalReadiness
}

// SetReady marks the application as ready to serve traffic.
func SetReady() {
	readinessMu.Lock()
	defer readinessMu.Unlock()
	globalReadiness = ReadinessStatus{
		Ready:     true,
		Timestamp: time.Now(),
	}
}

// SetNotReady marks the application as not ready and records the reason.
func SetNotReady(reason string) {
	readinessMu.Lock()
	defer readinessMu.Unlock()
	globalReadiness = ReadinessStatus{
		Ready:     false,
		Reason:    reason,
		Timestamp: time.Now(),
	}
}
