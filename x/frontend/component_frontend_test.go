package frontend

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

// TestFrontendComponentFromFS tests the NewFrontendComponentFromFS function
func TestFrontendComponentFromFS(t *testing.T) {
	// Create a simple test filesystem
	fs := http.Dir(".")

	// Create component
	component := NewFrontendComponentFromFS(fs)

	if component == nil {
		t.Fatal("Expected component to be created")
	}

	// Test component name and health
	name, healthStatus := component.Health()
	if name != "frontend_fs" {
		t.Errorf("Expected name 'frontend_fs', got '%s'", name)
	}
	if healthStatus.Status != health.StatusHealthy {
		t.Errorf("Expected healthy status, got %v", healthStatus.Status)
	}
}

// TestFrontendComponentFromDir tests the NewFrontendComponentFromDir function
func TestFrontendComponentFromDir(t *testing.T) {
	// Create component
	component := NewFrontendComponentFromDir("/nonexistent")

	if component == nil {
		t.Fatal("Expected component to be created")
	}

	// Test component name
	name, _ := component.Health()
	if name != "frontend_dir" {
		t.Errorf("Expected name 'frontend_dir', got '%s'", name)
	}
}

// TestFrontendComponentRegisterRoutes tests route registration
func TestFrontendComponentRegisterRoutes(t *testing.T) {
	// Create a mock router
	r := router.NewRouter()

	// Test with nil register (should not panic)
	component := &frontendComponent{
		register: nil,
		name:     "test",
	}

	component.RegisterRoutes(r)

	// Test with successful registration
	successCalled := false
	component2 := &frontendComponent{
		register: func(r *router.Router) error {
			successCalled = true
			return nil
		},
		name: "test2",
	}

	component2.RegisterRoutes(r)
	if !successCalled {
		t.Error("Expected register function to be called")
	}

	// Verify health status
	_, healthStatus := component2.Health()
	if healthStatus.Status != health.StatusHealthy {
		t.Errorf("Expected healthy status, got %v", healthStatus.Status)
	}
}

// TestFrontendComponentRegisterRoutesError tests error handling in route registration
func TestFrontendComponentRegisterRoutesError(t *testing.T) {
	r := router.NewRouter()
	expectedErr := errors.New("registration failed")

	component := &frontendComponent{
		register: func(r *router.Router) error {
			return expectedErr
		},
		name: "test_error",
	}

	component.RegisterRoutes(r)

	// Verify health status shows error
	_, healthStatus := component.Health()
	if healthStatus.Status != health.StatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %v", healthStatus.Status)
	}
	if healthStatus.Message != expectedErr.Error() {
		t.Errorf("Expected error message '%s', got '%s'", expectedErr.Error(), healthStatus.Message)
	}
}

// TestFrontendComponentLifecycle tests Start and Stop methods
func TestFrontendComponentLifecycle(t *testing.T) {
	component := NewFrontendComponentFromDir("/test")

	ctx := context.Background()

	// Test Start
	if err := component.Start(ctx); err != nil {
		t.Errorf("Start should return nil, got %v", err)
	}

	// Test Stop
	if err := component.Stop(ctx); err != nil {
		t.Errorf("Stop should return nil, got %v", err)
	}
}

// TestFrontendComponentRegisterMiddleware tests RegisterMiddleware
func TestFrontendComponentRegisterMiddleware(t *testing.T) {
	component := NewFrontendComponentFromDir("/test")

	// Should not panic
	reg := middleware.NewRegistry()
	component.RegisterMiddleware(reg)
}
