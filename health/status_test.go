package health

import "testing"

func TestBuildInfoAndReadiness(t *testing.T) {
	mgr, err := NewHealthManager(HealthCheckConfig{})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Initial state: not ready
	status := mgr.Readiness()
	if status.Ready {
		t.Fatalf("expected initial not-ready status, got ready")
	}

	mgr.MarkReady()
	if ready := mgr.Readiness(); !ready.Ready || ready.Reason != "" {
		t.Fatalf("expected ready without reason, got %+v", ready)
	}

	mgr.MarkNotReady("maintenance")
	notReady := mgr.Readiness()
	if notReady.Ready || notReady.Reason != "maintenance" {
		t.Fatalf("unexpected readiness after MarkNotReady: %+v", notReady)
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
