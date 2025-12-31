package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	manager := NewHealthManager()
	
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
	manager := NewHealthManager()
	
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
	manager := NewHealthManager()
	
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
	manager := NewHealthManager()
	
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
	manager := NewHealthManager()
	
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

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	components, ok := response["components"].([]interface{})
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
	manager := NewHealthManager()
	
	// Register a healthy component
	mockHealthy := &MockChecker{name: "healthy", healthy: true}
	manager.RegisterComponent(mockHealthy)

	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	rr := httptest.NewRecorder()

	ReadinessHandlerWithManager(manager).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when ready, got %d", rr.Code)
	}

	var response map[string]interface{}
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
