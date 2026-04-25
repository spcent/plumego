package health

import "testing"

const testDetailEnabled = "enabled"

func assertBoolDetail(t *testing.T, details map[string]any, key string, want bool) {
	t.Helper()

	got, ok := details[key].(bool)
	if !ok || got != want {
		t.Fatalf("expected detail %q = %v, got %v", key, want, details[key])
	}
}

func TestHealthStatusStates(t *testing.T) {
	status := HealthStatus{Status: StatusHealthy}

	if status.Status != StatusHealthy {
		t.Fatalf("expected healthy status by default")
	}

	status.Status = StatusDegraded
	status.Message = "component disabled"
	status.Details = map[string]any{testDetailEnabled: false}

	if status.Status != StatusDegraded {
		t.Fatalf("expected status to update to degraded")
	}
	if status.Message == "" {
		t.Fatalf("expected message to be set for degraded status")
	}
	assertBoolDetail(t, status.Details, testDetailEnabled, false)
}
