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

// Clone returns a copy of rs with component readiness decisions detached.
func (rs ReadinessStatus) Clone() ReadinessStatus {
	clone := rs
	if rs.Components != nil {
		clone.Components = make(map[string]bool, len(rs.Components))
		for name, ready := range rs.Components {
			clone.Components[name] = ready
		}
	}
	return clone
}
