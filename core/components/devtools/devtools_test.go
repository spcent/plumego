package devtools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/router"
)

func TestNewComponentDefaults(t *testing.T) {
	c := NewComponent(Options{})
	if c == nil {
		t.Fatal("expected non-nil component")
	}
	if c.logger == nil {
		t.Fatal("expected default logger")
	}
	if c.devMetrics == nil {
		t.Fatal("expected devMetrics to be initialized")
	}
	if c.debug {
		t.Fatal("expected debug false by default")
	}
}

func TestNewComponentWithDebug(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	if !c.debug {
		t.Fatal("expected debug true")
	}
}

func TestHealthDebugEnabled(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	name, status := c.Health()
	if name != "devtools" {
		t.Fatalf("expected name devtools, got %q", name)
	}
	if status.Status != health.StatusHealthy {
		t.Fatalf("expected healthy, got %s", status.Status)
	}
	if status.Details == nil {
		t.Fatal("expected details")
	}
	if val, ok := status.Details["enabled"]; !ok || val != true {
		t.Fatalf("expected enabled=true in details, got %v", status.Details)
	}
}

func TestHealthDebugDisabled(t *testing.T) {
	c := NewComponent(Options{Debug: false})
	name, status := c.Health()
	if name != "devtools" {
		t.Fatalf("expected name devtools, got %q", name)
	}
	if status.Status != health.StatusDegraded {
		t.Fatalf("expected degraded, got %s", status.Status)
	}
	if status.Message == "" {
		t.Fatal("expected non-empty message")
	}
}

func TestDependencies(t *testing.T) {
	c := NewComponent(Options{})
	if deps := c.Dependencies(); deps != nil {
		t.Fatalf("expected nil dependencies, got %v", deps)
	}
}

func TestRegisterRoutesDebugFalse(t *testing.T) {
	c := NewComponent(Options{Debug: false})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	// Debug routes should not be registered
	req := httptest.NewRequest(http.MethodGet, DevToolsRoutesPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should get 404 since routes weren't registered
	if rec.Code == http.StatusOK {
		t.Fatal("debug routes should not be registered when debug=false")
	}
}

func TestRegisterRoutesDebugTrue(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, DevToolsRoutesPath},
		{http.MethodGet, DevToolsRoutesJSONPath},
		{http.MethodGet, DevToolsMiddlewarePath},
		{http.MethodGet, DevToolsConfigPath},
		{http.MethodGet, DevToolsMetricsPath},
		{http.MethodGet, "/_routes"},
		{http.MethodGet, "/_config"},
		{http.MethodGet, "/_info"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code == http.StatusNotFound {
			t.Errorf("expected route %s %s to be registered, got 404", tt.method, tt.path)
		}
	}
}

func TestRoutesJSONEndpoint(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, DevToolsRoutesJSONPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data field")
	}
	if _, ok := data["routes"]; !ok {
		t.Fatal("expected routes in response")
	}
}

func TestConfigEndpointNoHook(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, DevToolsConfigPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data field")
	}
	if val, ok := data["debug"]; !ok || val != true {
		t.Fatalf("expected debug=true in config snapshot, got %v", data)
	}
}

func TestConfigEndpointWithHook(t *testing.T) {
	c := NewComponent(Options{
		Debug: true,
		Hooks: Hooks{
			ConfigSnapshot: func() map[string]any {
				return map[string]any{"custom": "value"}
			},
		},
	})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, DevToolsConfigPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["custom"] != "value" {
		t.Fatalf("expected custom=value, got %v", data)
	}
}

func TestMiddlewareEndpointNoHook(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, DevToolsMiddlewarePath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["middlewares"] != nil {
		t.Fatalf("expected nil middlewares, got %v", data["middlewares"])
	}
}

func TestMiddlewareEndpointWithHook(t *testing.T) {
	c := NewComponent(Options{
		Debug: true,
		Hooks: Hooks{
			MiddlewareList: func() []string {
				return []string{"logging", "recovery"}
			},
		},
	})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, DevToolsMiddlewarePath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	data := body["data"].(map[string]any)
	mws, ok := data["middlewares"].([]any)
	if !ok || len(mws) != 2 {
		t.Fatalf("expected 2 middlewares, got %v", data["middlewares"])
	}
}

func TestMetricsEndpoint(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, DevToolsMetricsPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["enabled"] != true {
		t.Fatalf("expected enabled=true, got %v", data["enabled"])
	}
}

func TestMetricsClearEndpoint(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, DevToolsMetricsClear, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReloadEndpointNoEnvFile(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, DevToolsReloadPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should fail because envFile is empty
	if rec.Code == http.StatusOK {
		t.Fatal("expected error when envFile is empty")
	}
}

func TestReloadEndpointEnvFileNotFound(t *testing.T) {
	c := NewComponent(Options{
		Debug:   true,
		EnvFile: "/nonexistent/.env",
	})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, DevToolsReloadPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Fatal("expected error when env file doesn't exist")
	}
}

func TestReloadEndpointSuccess(t *testing.T) {
	tmp := t.TempDir()
	envFile := filepath.Join(tmp, ".env")
	if err := os.WriteFile(envFile, []byte("TEST_KEY=test_value\n"), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewComponent(Options{
		Debug:   true,
		EnvFile: envFile,
	})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, DevToolsReloadPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestStartNoDebug(t *testing.T) {
	c := NewComponent(Options{Debug: false})
	if err := c.Start(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestStartNoEnvFile(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	if err := c.Start(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestStartEnvFileNotExist(t *testing.T) {
	c := NewComponent(Options{
		Debug:   true,
		EnvFile: "/nonexistent/.env",
	})
	if err := c.Start(context.Background()); err != nil {
		t.Fatalf("expected nil (graceful skip), got %v", err)
	}
}

func TestStartAndStop(t *testing.T) {
	tmp := t.TempDir()
	envFile := filepath.Join(tmp, ".env")
	if err := os.WriteFile(envFile, []byte("K=V\n"), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewComponent(Options{
		Debug:   true,
		EnvFile: envFile,
	})

	if err := c.Start(context.Background()); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	if err := c.Stop(context.Background()); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}

func TestStopWithoutStart(t *testing.T) {
	c := NewComponent(Options{})
	if err := c.Stop(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRegisterMiddlewareAttachesDevMetrics(t *testing.T) {
	var attached *metrics.DevCollector
	c := NewComponent(Options{
		Hooks: Hooks{
			AttachDevMetrics: func(dc *metrics.DevCollector) {
				attached = dc
			},
		},
	})
	c.RegisterMiddleware(nil)
	if attached == nil {
		t.Fatal("expected dev metrics to be attached via hook")
	}
}

func TestRegisterMiddlewareNoHook(t *testing.T) {
	c := NewComponent(Options{})
	// Should not panic even without the hook
	c.RegisterMiddleware(nil)
}

func TestInfoEndpoint(t *testing.T) {
	c := NewComponent(Options{
		Debug: true,
		Hooks: Hooks{
			ConfigSnapshot: func() map[string]any {
				return map[string]any{"app": "test"}
			},
		},
	})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/_info", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	data := body["data"].(map[string]any)
	if _, ok := data["config"]; !ok {
		t.Fatal("expected config in /_info response")
	}
	if _, ok := data["build"]; !ok {
		t.Fatal("expected build in /_info response")
	}
}

func TestRoutesTextEndpoint(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, DevToolsRoutesPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain; charset=utf-8" {
		t.Fatalf("expected text/plain content type, got %q", ct)
	}
}

func TestMetricsEndpointNilDevMetrics(t *testing.T) {
	c := NewComponent(Options{Debug: true})
	c.devMetrics = nil // force nil
	r := router.NewRouter()
	c.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, DevToolsMetricsPath, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["enabled"] != false {
		t.Fatalf("expected enabled=false with nil devMetrics, got %v", data["enabled"])
	}
}
