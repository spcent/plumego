package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/log"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected bool
	}{
		{
			name:     "latest is newer",
			current:  "1.0.0",
			latest:   "1.0.1",
			expected: true,
		},
		{
			name:     "latest is major version newer",
			current:  "1.0.0",
			latest:   "2.0.0",
			expected: true,
		},
		{
			name:     "versions are equal",
			current:  "1.0.0",
			latest:   "1.0.0",
			expected: false,
		},
		{
			name:     "current is newer",
			current:  "1.0.1",
			latest:   "1.0.0",
			expected: false,
		},
		{
			name:     "minor version difference",
			current:  "1.1.0",
			latest:   "1.2.0",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.current, tt.latest)
			if result != tt.expected {
				t.Errorf("CompareVersions(%s, %s) = %v, want %v",
					tt.current, tt.latest, result, tt.expected)
			}
		})
	}
}

func TestChecker_CheckForUpdate(t *testing.T) {
	// Create a test server that returns mock release info
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := LatestRelease{
			Version:     "1.1.0",
			Commit:      "abc123",
			BuildTime:   time.Now().Format(time.RFC3339),
			Channel:     "stable",
			ReleaseDate: time.Now().Format(time.RFC3339),
			ReleaseURL:  "https://example.com/releases/v1.1.0",
			DownloadURL: "https://example.com/download/v1.1.0",
			Changelog:   "Bug fixes and improvements",
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	currentVersion := VersionInfo{
		Version:   "1.0.0",
		Commit:    "xyz789",
		BuildTime: time.Now().Format(time.RFC3339),
		Channel:   "stable",
	}

	checker := NewChecker(currentVersion, server.URL)

	ctx := context.Background()
	release, err := checker.CheckForUpdate(ctx)
	if err != nil {
		t.Fatalf("CheckForUpdate failed: %v", err)
	}

	if release.Version != "1.1.0" {
		t.Errorf("expected version 1.1.0, got %s", release.Version)
	}

	if release.Commit != "abc123" {
		t.Errorf("expected commit abc123, got %s", release.Commit)
	}
}

func TestChecker_CheckForUpdate_ServerError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	currentVersion := VersionInfo{
		Version: "1.0.0",
	}

	checker := NewChecker(currentVersion, server.URL)

	ctx := context.Background()
	_, err := checker.CheckForUpdate(ctx)
	if err == nil {
		t.Error("expected error for server error response, got nil")
	}
}

func TestService_GetStatus_NoCache(t *testing.T) {
	currentVersion := VersionInfo{
		Version: "1.0.0",
		Channel: "stable",
	}

	checker := NewChecker(currentVersion, "")
	config := &Config{
		Enabled: true,
	}
	logger := log.NewLogger()

	service := NewService(checker, config, logger)

	ctx := context.Background()
	status := service.GetStatus(ctx)

	if status.CurrentVersion.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", status.CurrentVersion.Version)
	}

	if status.CheckEnabled != true {
		t.Error("expected CheckEnabled to be true")
	}

	if status.UpdateAvailable {
		t.Error("expected UpdateAvailable to be false when no cache")
	}
}

func TestService_CheckNow(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := LatestRelease{
			Version: "1.1.0",
			Channel: "stable",
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	currentVersion := VersionInfo{
		Version: "1.0.0",
		Channel: "stable",
	}

	checker := NewChecker(currentVersion, server.URL)
	config := &Config{
		Enabled:          true,
		CheckIntervalMin: 60,
	}
	logger := log.NewLogger()

	service := NewService(checker, config, logger)

	ctx := context.Background()
	status, err := service.CheckNow(ctx)
	if err != nil {
		t.Fatalf("CheckNow failed: %v", err)
	}

	if !status.UpdateAvailable {
		t.Error("expected UpdateAvailable to be true")
	}

	if status.LatestRelease.Version != "1.1.0" {
		t.Errorf("expected latest version 1.1.0, got %s", status.LatestRelease.Version)
	}

	if status.LastCheck == nil {
		t.Error("expected LastCheck to be set")
	}

	if status.NextCheck == nil {
		t.Error("expected NextCheck to be set")
	}

	// Verify cache is populated
	cachedStatus := service.GetStatus(ctx)
	if cachedStatus.LatestRelease == nil {
		t.Error("expected cached status to have LatestRelease")
	}
}

func TestService_CheckNow_Disabled(t *testing.T) {
	currentVersion := VersionInfo{
		Version: "1.0.0",
	}

	checker := NewChecker(currentVersion, "")
	config := &Config{
		Enabled: false,
	}
	logger := log.NewLogger()

	service := NewService(checker, config, logger)

	ctx := context.Background()
	status, err := service.CheckNow(ctx)
	if err != nil {
		t.Fatalf("CheckNow failed: %v", err)
	}

	if status.CheckEnabled {
		t.Error("expected CheckEnabled to be false")
	}

	if status.LatestRelease != nil {
		t.Error("expected LatestRelease to be nil when disabled")
	}
}

func TestNewChecker_DefaultURL(t *testing.T) {
	currentVersion := VersionInfo{
		Version: "1.0.0",
	}

	checker := NewChecker(currentVersion, "")

	if checker.updateURL != DefaultUpdateURL {
		t.Errorf("expected default URL %s, got %s", DefaultUpdateURL, checker.updateURL)
	}
}

func TestNewChecker_CustomURL(t *testing.T) {
	currentVersion := VersionInfo{
		Version: "1.0.0",
	}

	customURL := "https://custom.example.com/update.json"
	checker := NewChecker(currentVersion, customURL)

	if checker.updateURL != customURL {
		t.Errorf("expected custom URL %s, got %s", customURL, checker.updateURL)
	}
}
