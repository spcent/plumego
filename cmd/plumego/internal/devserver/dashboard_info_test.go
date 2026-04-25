package devserver

import (
	"bytes"
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

func TestConfigEditSaveUsesTypedResponse(t *testing.T) {
	tmp := t.TempDir()
	d := &Dashboard{
		projectDir: tmp,
		runner:     NewAppRunner(tmp, nil),
	}

	body := bytes.NewBufferString(`{"entries":[{"key":"APP_NAME","value":"demo"},{"key":"APP_DEBUG","value":"true"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/config-edit", body)
	rec := httptest.NewRecorder()

	d.handleConfigEditSave(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	resp := decodeDevserverData[ConfigEditSaveResponse](t, rec)
	if !resp.Success || resp.Path != defaultConfigEditFile || resp.Count != 2 || resp.Restarted {
		t.Fatalf("unexpected config edit save response: %+v", resp)
	}

	data, err := os.ReadFile(filepath.Join(tmp, defaultConfigEditFile))
	if err != nil {
		t.Fatalf("read saved env: %v", err)
	}
	if got := string(data); !strings.Contains(got, "APP_NAME=demo") || !strings.Contains(got, "APP_DEBUG=true") {
		t.Fatalf("unexpected env file content: %q", got)
	}
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

func TestDashboardStatusUsesTypedResponse(t *testing.T) {
	tmp := t.TempDir()
	d := &Dashboard{
		dashboardAddr: ":9999",
		appAddr:       ":8080",
		projectDir:    tmp,
		startTime:     time.Now().Add(-2 * time.Second),
		runner:        NewAppRunner(tmp, nil),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	d.handleStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	resp := decodeDevserverData[dashboardStatusResponse](t, rec)
	if resp.Dashboard.URL != "http://localhost:9999" {
		t.Fatalf("dashboard url = %q, want http://localhost:9999", resp.Dashboard.URL)
	}
	if resp.App.URL != "http://localhost:8080" || resp.App.Running {
		t.Fatalf("unexpected app status: %+v", resp.App)
	}
	if resp.Project.Dir != tmp || resp.Project.GoVersion == "" {
		t.Fatalf("unexpected project status: %+v", resp.Project)
	}
}

func TestDashboardHealthAndMetricsUseTypedResponses(t *testing.T) {
	tmp := t.TempDir()
	d := &Dashboard{
		projectDir: tmp,
		startTime:  time.Now().Add(-2 * time.Second),
		runner:     NewAppRunner(tmp, nil),
	}

	healthReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	healthRec := httptest.NewRecorder()
	d.handleHealth(healthRec, healthReq)

	if healthRec.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthRec.Code, http.StatusOK)
	}
	health := decodeDevserverData[dashboardHealthResponse](t, healthRec)
	if health.Healthy || health.Checks.App != "stopped" {
		t.Fatalf("unexpected health response: %+v", health)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	metricsRec := httptest.NewRecorder()
	d.handleMetrics(metricsRec, metricsReq)

	if metricsRec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want %d", metricsRec.Code, http.StatusOK)
	}
	metrics := decodeDevserverData[dashboardMetricsResponse](t, metricsRec)
	if metrics.App.Running || metrics.App.PID != 0 {
		t.Fatalf("unexpected metrics app response: %+v", metrics.App)
	}
	if metrics.Dashboard.StartTime == "" || metrics.Thresholds.MinTotalCount == 0 {
		t.Fatalf("unexpected metrics response: %+v", metrics)
	}
}

func TestDashboardPprofTypesUseTypedResponse(t *testing.T) {
	d := &Dashboard{}
	req := httptest.NewRequest(http.MethodGet, "/api/pprof/types", nil)
	rec := httptest.NewRecorder()

	d.handlePprofTypes(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	resp := decodeDevserverData[dashboardPprofTypesResponse](t, rec)
	if len(resp.Types) == 0 {
		t.Fatal("expected pprof types")
	}
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

func decodeDevserverData[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	if got := rec.Header().Get(contract.HeaderContentType); got != contract.ContentTypeJSON {
		t.Fatalf("content type = %q, want %q", got, contract.ContentTypeJSON)
	}

	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode success envelope: %v; body: %s", err, rec.Body.String())
	}
	if len(env.Data) == 0 {
		t.Fatalf("success envelope missing data; body: %s", rec.Body.String())
	}

	var body T
	if err := json.Unmarshal(env.Data, &body); err != nil {
		t.Fatalf("decode success data: %v; data: %s", err, string(env.Data))
	}
	return body
}
