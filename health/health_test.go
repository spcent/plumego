package health

import "testing"

func TestHealthState_IsReady(t *testing.T) {
	tests := []struct {
		name  string
		state HealthState
		want  bool
	}{
		{name: "healthy", state: StatusHealthy, want: true},
		{name: "degraded", state: StatusDegraded, want: true},
		{name: "unhealthy", state: StatusUnhealthy, want: false},
		{name: "unknown", state: "unknown", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.state.IsReady(); got != tc.want {
				t.Errorf("HealthState(%q).IsReady() = %v, want %v", tc.state, got, tc.want)
			}
		})
	}
}

func TestComponentHealth_EmbeddedStatus(t *testing.T) {
	ch := ComponentHealth{
		HealthStatus: HealthStatus{Status: StatusHealthy, Message: "ok"},
		Enabled:      true,
	}
	if ch.Status != StatusHealthy {
		t.Errorf("Status = %q, want healthy", ch.Status)
	}
	if !ch.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestReadinessStatus_Fields(t *testing.T) {
	rs := ReadinessStatus{
		Ready:      true,
		Reason:     "all components up",
		Components: map[string]bool{"db": true, "cache": false},
	}
	if !rs.Ready {
		t.Error("Ready should be true")
	}
	if rs.Components["cache"] {
		t.Error("cache component should be false")
	}
	if !rs.Components["db"] {
		t.Error("db component should be true")
	}
}
