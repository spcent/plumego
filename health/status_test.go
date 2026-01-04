package health

import "testing"

func TestBuildInfoAndReadiness(t *testing.T) {
	// Reset readiness to ensure clean test
	SetNotReady("test-reset")

	// default readiness should be not ready
	status := GetReadiness()
	if status.Ready {
		t.Fatalf("expected default not ready status")
	}

	SetReady()
	if ready := GetReadiness(); !ready.Ready || ready.Reason != "" {
		t.Fatalf("expected ready without reason, got %+v", ready)
	}

	SetNotReady("maintenance")
	notReady := GetReadiness()
	if notReady.Ready || notReady.Reason != "maintenance" {
		t.Fatalf("unexpected readiness after SetNotReady: %+v", notReady)
	}

	info := GetBuildInfo()
	if info.Version == "" || info.Commit == "" || info.BuildTime == "" {
		t.Fatalf("expected build info to include defaults, got %+v", info)
	}
}

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
