package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"
)

// MockChecker is a simple implementation of ComponentChecker for testing.
type MockChecker struct {
	name    string
	healthy bool
	delay   time.Duration
}

func (mc *MockChecker) Name() string {
	return mc.name
}

func (mc *MockChecker) Check(ctx context.Context) error {
	if mc.delay > 0 {
		select {
		case <-time.After(mc.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if mc.healthy {
		return nil
	}
	return MockError("mock component failure")
}

type MockError string

func (e MockError) Error() string {
	return string(e)
}

func TestHealthHandler(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Register a healthy component
	mockHealthy := &MockChecker{name: "healthy", healthy: true}
	manager.RegisterComponent(mockHealthy)

	// Register an unhealthy component
	mockUnhealthy := &MockChecker{name: "unhealthy", healthy: false}
	manager.RegisterComponent(mockUnhealthy)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(manager).ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when components are unhealthy, got %d", rr.Code)
	}

	var health HealthStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &health); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if health.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy status, got %v", health.Status)
	}
}

func TestComponentHealthHandler(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	mockHealthy := &MockChecker{name: "healthy", healthy: true}
	manager.RegisterComponent(mockHealthy)

	// Test existing component
	req := httptest.NewRequest(http.MethodGet, "/health/component/healthy", nil)
	rr := httptest.NewRecorder()

	ComponentHealthHandler(manager, "healthy").ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for healthy component, got %d", rr.Code)
	}

	var health ComponentHealth
	if err := json.Unmarshal(rr.Body.Bytes(), &health); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if health.Status != StatusHealthy {
		t.Fatalf("expected healthy status, got %v", health.Status)
	}

	// Test non-existent component
	req = httptest.NewRequest(http.MethodGet, "/health/component/nonexistent", nil)
	rr = httptest.NewRecorder()

	ComponentHealthHandler(manager, "nonexistent").ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent component, got %d", rr.Code)
	}
}

func TestAllComponentsHealthHandler(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	mock1 := &MockChecker{name: "component1", healthy: true}
	mock2 := &MockChecker{name: "component2", healthy: false}

	manager.RegisterComponent(mock1)
	manager.RegisterComponent(mock2)

	req := httptest.NewRequest(http.MethodGet, "/health/all", nil)
	rr := httptest.NewRecorder()

	AllComponentsHealthHandler(manager).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var allHealth map[string]*ComponentHealth
	if err := json.Unmarshal(rr.Body.Bytes(), &allHealth); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(allHealth) != 2 {
		t.Fatalf("expected 2 components, got %d", len(allHealth))
	}

	if _, exists := allHealth["component1"]; !exists {
		t.Fatalf("component1 not found in response")
	}

	if _, exists := allHealth["component2"]; !exists {
		t.Fatalf("component2 not found in response")
	}
}

func TestHealthHistoryHandler(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
		EnableHistory:      true, // Enable history for this test
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	mock := &MockChecker{name: "test", healthy: true}
	manager.RegisterComponent(mock)

	// Trigger some health checks
	ctx := context.Background()
	manager.CheckAllComponents(ctx)

	req := httptest.NewRequest(http.MethodGet, "/health/history", nil)
	rr := httptest.NewRecorder()

	HealthHistoryHandler(manager).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var history []HealthHistoryEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &history); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(history) == 0 {
		t.Fatalf("expected history entries, got none")
	}
}

func TestLiveHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rr := httptest.NewRecorder()

	LiveHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if body := rr.Body.String(); body != "alive" {
		t.Fatalf("expected 'alive', got '%s'", body)
	}
}

func TestComponentsListHandler(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	mock1 := &MockChecker{name: "comp1", healthy: true}
	mock2 := &MockChecker{name: "comp2", healthy: true}

	manager.RegisterComponent(mock1)
	manager.RegisterComponent(mock2)

	req := httptest.NewRequest(http.MethodGet, "/health/components", nil)
	rr := httptest.NewRecorder()

	ComponentsListHandler(manager).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	components, ok := response["components"].([]any)
	if !ok {
		t.Fatalf("components field not found or not an array")
	}

	if len(components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(components))
	}

	count, ok := response["count"].(float64)
	if !ok || int(count) != 2 {
		t.Fatalf("expected count to be 2, got %v", count)
	}
}

func TestReadinessHandlerWithManager(t *testing.T) {
	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Register a healthy component
	mockHealthy := &MockChecker{name: "healthy", healthy: true}
	manager.RegisterComponent(mockHealthy)

	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	rr := httptest.NewRecorder()

	ReadinessHandlerWithManager(manager).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when ready, got %d", rr.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if ready, ok := response["ready"].(bool); !ok || !ready {
		t.Fatalf("expected ready=true, got %v", response["ready"])
	}
}

func TestReadinessHandler(t *testing.T) {
	t.Cleanup(func() { SetNotReady("starting") })
	SetNotReady("booting")
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()

	ReadinessHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when not ready, got %d", rr.Code)
	}

	var payload ReadinessStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.Ready {
		t.Fatalf("expected ready=false, got true")
	}

	SetReady()
	rr = httptest.NewRecorder()
	ReadinessHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when ready, got %d", rr.Code)
	}
}

func TestBuildInfoHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/build", nil)
	rr := httptest.NewRecorder()

	BuildInfoHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var info BuildInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &info); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if info.Version == "" {
		t.Fatalf("version should be populated from defaults")
	}
}

func TestHealthStateIsReady(t *testing.T) {
	if !StatusHealthy.isReady() {
		t.Fatalf("StatusHealthy should be ready")
	}

	if !StatusDegraded.isReady() {
		t.Fatalf("StatusDegraded should be ready")
	}

	if StatusUnhealthy.isReady() {
		t.Fatalf("StatusUnhealthy should not be ready")
	}
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "no env vars set",
			envVars:  map[string]string{},
			expected: false,
		},
		{
			name:     "APP_ENV=development",
			envVars:  map[string]string{"APP_ENV": "development"},
			expected: true,
		},
		{
			name:     "APP_ENV=Development (case insensitive)",
			envVars:  map[string]string{"APP_ENV": "Development"},
			expected: true,
		},
		{
			name:     "APP_ENV=production",
			envVars:  map[string]string{"APP_ENV": "production"},
			expected: false,
		},
		{
			name:     "APP_DEBUG=true",
			envVars:  map[string]string{"APP_DEBUG": "true"},
			expected: true,
		},
		{
			name:     "APP_DEBUG=True (case insensitive)",
			envVars:  map[string]string{"APP_DEBUG": "True"},
			expected: true,
		},
		{
			name:     "APP_DEBUG=false",
			envVars:  map[string]string{"APP_DEBUG": "false"},
			expected: false,
		},
		{
			name:     "APP_ENV=production and APP_DEBUG=true",
			envVars:  map[string]string{"APP_ENV": "production", "APP_DEBUG": "true"},
			expected: true,
		},
		{
			name:     "APP_ENV=development and APP_DEBUG=false",
			envVars:  map[string]string{"APP_ENV": "development", "APP_DEBUG": "false"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear relevant env vars before each subtest
			t.Setenv("APP_ENV", "")
			t.Setenv("APP_DEBUG", "")

			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			got := isDevelopment()
			if got != tt.expected {
				t.Fatalf("isDevelopment() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHealthHandlerIncludesRuntimeInDevMode(t *testing.T) {
	t.Cleanup(func() { SetNotReady("starting") })

	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	mock := &MockChecker{name: "test", healthy: true}
	manager.RegisterComponent(mock)

	t.Run("runtime included in dev mode", func(t *testing.T) {
		t.Setenv("APP_DEBUG", "true")

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()

		HealthHandler(manager).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}

		var response HealthResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if response.Runtime == nil {
			t.Fatal("expected runtime info in dev mode, got nil")
		}
		if response.Runtime.GoVersion != runtime.Version() {
			t.Fatalf("expected go version %s, got %s", runtime.Version(), response.Runtime.GoVersion)
		}
		if response.Runtime.NumCPU != runtime.NumCPU() {
			t.Fatalf("expected num_cpu %d, got %d", runtime.NumCPU(), response.Runtime.NumCPU)
		}
		if response.Runtime.GOOS != runtime.GOOS {
			t.Fatalf("expected goos %s, got %s", runtime.GOOS, response.Runtime.GOOS)
		}
	})

	t.Run("runtime omitted in production", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("APP_DEBUG", "")

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()

		HealthHandler(manager).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}

		var response HealthResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		if response.Runtime != nil {
			t.Fatal("expected runtime info to be nil in production mode")
		}
	})
}

func TestHandlePanicDevMode(t *testing.T) {
	t.Run("includes stack trace in dev mode", func(t *testing.T) {
		t.Setenv("APP_DEBUG", "true")

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()

		handlePanic(rr, req, "test panic value", "req-123")

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rr.Code)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "test panic value") {
			t.Fatalf("dev mode panic response should contain panic value, got: %s", body)
		}
		if !strings.Contains(body, "goroutine") {
			t.Fatalf("dev mode panic response should contain stack trace, got: %s", body)
		}
	})

	t.Run("hides details in production", func(t *testing.T) {
		t.Setenv("APP_ENV", "")
		t.Setenv("APP_DEBUG", "")

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()

		handlePanic(rr, req, "secret panic value", "req-456")

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rr.Code)
		}

		body := rr.Body.String()
		if strings.Contains(body, "secret panic value") {
			t.Fatalf("production panic response must NOT contain panic value, got: %s", body)
		}
		if strings.Contains(body, "goroutine") {
			t.Fatalf("production panic response must NOT contain stack trace, got: %s", body)
		}
	})
}

func TestDebugHealthHandler(t *testing.T) {
	t.Cleanup(func() { SetNotReady("starting") })

	config := HealthCheckConfig{
		MaxHistoryEntries:  100,
		HistoryRetention:   24 * time.Hour,
		AutoCleanupEnabled: false,
	}
	manager, err := NewHealthManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	mock := &MockChecker{name: "db", healthy: true}
	manager.RegisterComponent(mock)

	t.Run("returns 404 in production", func(t *testing.T) {
		t.Setenv("APP_ENV", "")
		t.Setenv("APP_DEBUG", "")

		req := httptest.NewRequest(http.MethodGet, "/health/debug", nil)
		rr := httptest.NewRecorder()

		DebugHealthHandler(manager).ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 in production, got %d", rr.Code)
		}
	})

	t.Run("returns diagnostics in dev mode via APP_ENV", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")

		req := httptest.NewRequest(http.MethodGet, "/health/debug", nil)
		rr := httptest.NewRecorder()

		DebugHealthHandler(manager).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 in dev mode, got %d", rr.Code)
		}

		var response map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		// Verify runtime info is present
		if _, ok := response["runtime"]; !ok {
			t.Fatal("expected runtime field in debug response")
		}
		// Verify build info is present
		if _, ok := response["build_info"]; !ok {
			t.Fatal("expected build_info field in debug response")
		}
		// Verify readiness is present
		if _, ok := response["readiness"]; !ok {
			t.Fatal("expected readiness field in debug response")
		}
		// Verify health info is present (manager was provided)
		if _, ok := response["health"]; !ok {
			t.Fatal("expected health field in debug response")
		}
		// Verify components info is present
		if _, ok := response["components"]; !ok {
			t.Fatal("expected components field in debug response")
		}
		// Verify config is present
		if _, ok := response["config"]; !ok {
			t.Fatal("expected config field in debug response")
		}
	})

	t.Run("returns diagnostics in dev mode via APP_DEBUG", func(t *testing.T) {
		t.Setenv("APP_ENV", "")
		t.Setenv("APP_DEBUG", "true")

		req := httptest.NewRequest(http.MethodGet, "/health/debug", nil)
		rr := httptest.NewRecorder()

		DebugHealthHandler(manager).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 with APP_DEBUG=true, got %d", rr.Code)
		}
	})

	t.Run("works with nil manager", func(t *testing.T) {
		t.Setenv("APP_DEBUG", "true")

		req := httptest.NewRequest(http.MethodGet, "/health/debug", nil)
		rr := httptest.NewRecorder()

		DebugHealthHandler(nil).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 even with nil manager, got %d", rr.Code)
		}

		var response map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}

		// Should still have runtime and build info
		if _, ok := response["runtime"]; !ok {
			t.Fatal("expected runtime field even with nil manager")
		}
		// Should not have health/components/config when manager is nil
		if _, ok := response["health"]; ok {
			t.Fatal("expected no health field with nil manager")
		}
	})
}

func TestGetRuntimeInfo(t *testing.T) {
	info := getRuntimeInfo()

	if info.GoVersion != runtime.Version() {
		t.Fatalf("expected go version %s, got %s", runtime.Version(), info.GoVersion)
	}
	if info.NumCPU != runtime.NumCPU() {
		t.Fatalf("expected num_cpu %d, got %d", runtime.NumCPU(), info.NumCPU)
	}
	if info.GOARCH != runtime.GOARCH {
		t.Fatalf("expected goarch %s, got %s", runtime.GOARCH, info.GOARCH)
	}
	if info.GOOS != runtime.GOOS {
		t.Fatalf("expected goos %s, got %s", runtime.GOOS, info.GOOS)
	}
	if info.NumGoroutine <= 0 {
		t.Fatalf("expected positive num_goroutine, got %d", info.NumGoroutine)
	}
	if info.MemSys == 0 {
		t.Fatal("expected non-zero mem_sys_bytes")
	}
}
