package devserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
)

func TestGetDashboardInfo(t *testing.T) {
	d := &Dashboard{
		dashboardAddr: ":9999",
		appAddr:       ":8080",
		projectDir:    "/tmp/test-project",
		startTime:     time.Now().Add(-5 * time.Second),
		runner:        NewAppRunner("/tmp/test-project", nil),
	}

	info := d.getDashboardInfo()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	if info.DashboardURL != "http://localhost:9999" {
		t.Errorf("DashboardURL = %q, want %q", info.DashboardURL, "http://localhost:9999")
	}

	if info.AppURL != "http://localhost:8080" {
		t.Errorf("AppURL = %q, want %q", info.AppURL, "http://localhost:8080")
	}

	if info.Uptime == "" {
		t.Error("Uptime should not be empty")
	}

	if info.UptimeMS < 5000 {
		t.Errorf("UptimeMS = %d, expected >= 5000", info.UptimeMS)
	}

	if info.StartTime == "" {
		t.Error("StartTime should not be empty")
	}

	if _, err := time.Parse(time.RFC3339, info.StartTime); err != nil {
		t.Errorf("StartTime should be valid RFC3339: %v", err)
	}

	if info.ProjectDir != "/tmp/test-project" {
		t.Errorf("ProjectDir = %q, want %q", info.ProjectDir, "/tmp/test-project")
	}

	goVer := runtime.Version()
	if info.GoVersion != goVer {
		t.Errorf("GoVersion = %q, want %q", info.GoVersion, goVer)
	}

	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Errorf("GoVersion should start with 'go', got %q", info.GoVersion)
	}

	if info.AppRunning {
		t.Error("AppRunning should be false when runner is not running")
	}

	if info.AppPID != 0 {
		t.Errorf("AppPID should be 0 when app is not running, got %d", info.AppPID)
	}
}

func TestGetDashboardInfoFieldsPopulated(t *testing.T) {
	d := &Dashboard{
		dashboardAddr: ":3000",
		appAddr:       ":4000",
		projectDir:    "/home/user/myapp",
		startTime:     time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		runner:        NewAppRunner("/home/user/myapp", nil),
	}

	info := d.getDashboardInfo()

	// Verify all fields are non-zero/non-empty
	checks := []struct {
		name  string
		empty bool
	}{
		{"Version", info.Version == ""},
		{"DashboardURL", info.DashboardURL == ""},
		{"AppURL", info.AppURL == ""},
		{"Uptime", info.Uptime == ""},
		{"StartTime", info.StartTime == ""},
		{"ProjectDir", info.ProjectDir == ""},
		{"GoVersion", info.GoVersion == ""},
	}

	for _, check := range checks {
		if check.empty {
			t.Errorf("%s should not be empty", check.name)
		}
	}

	if info.UptimeMS <= 0 {
		t.Errorf("UptimeMS should be > 0, got %d", info.UptimeMS)
	}
}

func TestConfigEditReadErrorUsesStableSafeResponse(t *testing.T) {
	tmp := t.TempDir()
	projectFile := filepath.Join(tmp, "project-file")
	if err := os.WriteFile(projectFile, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write project file: %v", err)
	}

	d := &Dashboard{
		projectDir: projectFile,
		runner:     NewAppRunner(projectFile, nil),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/config-edit", nil)
	rec := httptest.NewRecorder()

	d.handleConfigEditGet(rec, req)

	assertDevserverError(t, rec, http.StatusInternalServerError, devserverCodeConfigEditReadFailed, "config edit file could not be read")
	assertDevserverBodyOmits(t, rec.Body.String(), "not a directory")
}

func TestDashboardAppNotRunningUsesStableCode(t *testing.T) {
	tmp := t.TempDir()
	d := &Dashboard{
		projectDir: tmp,
		runner:     NewAppRunner(tmp, nil),
		analyzer:   NewAnalyzer("http://127.0.0.1:1"),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/routes", nil)
	rec := httptest.NewRecorder()

	d.handleRoutes(rec, req)

	assertDevserverError(t, rec, http.StatusServiceUnavailable, devserverCodeAppNotRunning, "application is not running")
}

func TestDepsErrorUsesStableSafeResponse(t *testing.T) {
	tmp := t.TempDir()
	d := &Dashboard{
		projectDir: tmp,
		depsCache:  newDepsCache(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/deps?refresh=1", nil)
	rec := httptest.NewRecorder()

	d.handleDeps(rec, req)

	assertDevserverError(t, rec, http.StatusInternalServerError, devserverCodeDependencyGraphFailed, "dependency graph unavailable")
	assertDevserverBodyOmits(t, rec.Body.String(), "go list")
}

func assertDevserverError(t *testing.T, rec *httptest.ResponseRecorder, status int, code, message string) {
	t.Helper()

	if rec.Code != status {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, status, rec.Body.String())
	}

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v; body: %s", err, rec.Body.String())
	}
	if resp.Error.Code != code {
		t.Fatalf("code = %q, want %q", resp.Error.Code, code)
	}
	if resp.Error.Message != message {
		t.Fatalf("message = %q, want %q", resp.Error.Message, message)
	}
}

func assertDevserverBodyOmits(t *testing.T, body, value string) {
	t.Helper()
	if strings.Contains(body, value) {
		t.Fatalf("response leaked %q: %s", value, body)
	}
}
