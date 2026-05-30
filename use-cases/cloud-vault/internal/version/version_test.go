package version

import (
	"testing"
)

func TestGetBuildInfo(t *testing.T) {
	info := GetBuildInfo()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	if info.Commit == "" {
		t.Error("Commit should not be empty")
	}

	if info.BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}

	if info.Channel == "" {
		t.Error("Channel should not be empty")
	}
}

func TestGetBuildInfo_DefaultValues(t *testing.T) {
	// Test that default values are set when ldflags are not used
	info := GetBuildInfo()

	if info.Version != "1.0.0" {
		t.Errorf("Expected default Version '1.0.0', got '%s'", info.Version)
	}

	if info.Commit != "dev" {
		t.Errorf("Expected default Commit 'dev', got '%s'", info.Commit)
	}

	if info.BuildTime != "unknown" {
		t.Errorf("Expected default BuildTime 'unknown', got '%s'", info.BuildTime)
	}

	if info.Channel != "stable" {
		t.Errorf("Expected default Channel 'stable', got '%s'", info.Channel)
	}
}
