package health

import "testing"

func TestBuildInfoAndReadiness(t *testing.T) {
	// default readiness set in init
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
