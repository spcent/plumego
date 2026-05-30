package update

import (
	"testing"
)

func TestNewChecker_WithCustomURL(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
	}

	customURL := "https://my-custom-update-server.com/version.json"
	checker := NewChecker(versionInfo, customURL)

	if checker.updateURL != customURL {
		t.Errorf("expected updateURL to be %s, got %s", customURL, checker.updateURL)
	}

	if checker.currentVersion.Version != versionInfo.Version {
		t.Errorf("expected currentVersion.Version to be %s, got %s",
			versionInfo.Version, checker.currentVersion.Version)
	}
}

func TestNewChecker_WithEmptyURL(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
	}

	checker := NewChecker(versionInfo, "")

	if checker.updateURL != DefaultUpdateURL {
		t.Errorf("expected updateURL to be %s when empty string provided, got %s",
			DefaultUpdateURL, checker.updateURL)
	}
}
