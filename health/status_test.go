package health

import "testing"

func TestHealthStatusStates(t *testing.T) {
	status := HealthStatus{Status: StatusHealthy}

	if status.Status != StatusHealthy {
		t.Fatalf("expected healthy status by default")
	}

	status.Status = StatusDegraded
	status.Message = "component disabled"
	status.Details = map[string]any{"enabled": false}

	if status.Status != StatusDegraded {
		t.Fatalf("expected status to update to degraded")
	}
	if status.Message == "" {
		t.Fatalf("expected message to be set for degraded status")
	}
	if enabled, ok := status.Details["enabled"].(bool); !ok || enabled {
		t.Fatalf("expected enabled detail to reflect degraded state")
	}
}
