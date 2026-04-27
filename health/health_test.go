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

func TestHealthState_IsKnown(t *testing.T) {
	tests := []struct {
		name  string
		state HealthState
		want  bool
	}{
		{name: "healthy", state: StatusHealthy, want: true},
		{name: "degraded", state: StatusDegraded, want: true},
		{name: "unhealthy", state: StatusUnhealthy, want: true},
		{name: "unknown", state: "unknown", want: false},
		{name: "empty", state: "", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.state.IsKnown(); got != tc.want {
				t.Errorf("HealthState(%q).IsKnown() = %v, want %v", tc.state, got, tc.want)
			}
		})
	}
}
