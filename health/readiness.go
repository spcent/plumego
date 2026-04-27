package health

import "time"

// ReadinessStatus describes whether the application is ready to serve traffic.
// Components maps component names to their readiness decisions when available.
type ReadinessStatus struct {
	Ready      bool            `json:"ready"`
	Reason     string          `json:"reason,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
	Components map[string]bool `json:"components,omitempty"`
}
