package core

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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

// TestPubSubDebugComponent tests the pubsub debug component
func TestPubSubDebugComponent(t *testing.T) {
	// Create a mock pubsub for testing
	type mockPubSub struct{}

	cfg := PubSubConfig{
		Enabled: true,
		Path:    "/debug/pubsub",
	}

	// Test with nil pubsub
	component := newPubSubDebugComponent(cfg, nil)

	if component == nil {
		t.Fatal("Expected component to be created")
	}

	// Test lifecycle
	ctx := context.Background()
	if err := component.Start(ctx); err != nil {
		t.Errorf("Start should return nil, got %v", err)
	}
	if err := component.Stop(ctx); err != nil {
		t.Errorf("Stop should return nil, got %v", err)
	}

	// Test health
	name, healthStatus := component.Health()
	if name != "pubsub_debug" {
		t.Errorf("Expected name 'pubsub_debug', got '%s'", name)
	}
	if healthStatus.Status != health.StatusHealthy {
		t.Errorf("Expected healthy status, got %v", healthStatus.Status)
	}
}

// TestPubSubDebugComponentRegisterRoutes tests route registration
func TestPubSubDebugComponentRegisterRoutes(t *testing.T) {
	r := router.NewRouter()
	cfg := PubSubConfig{
		Enabled: true,
		Path:    "/debug/pubsub",
	}

	component := newPubSubDebugComponent(cfg, nil)

	// This should not panic even with nil pubsub
	component.RegisterRoutes(r)
}

// TestPubSubDebugComponentRegisterMiddleware tests middleware registration
func TestPubSubDebugComponentRegisterMiddleware(t *testing.T) {
	cfg := PubSubConfig{
		Enabled: true,
		Path:    "/debug/pubsub",
	}

	component := newPubSubDebugComponent(cfg, nil)
	reg := middleware.NewRegistry()

	// Should not panic
	component.RegisterMiddleware(reg)
}

// TestHasComponentType tests the hasComponentType method
func TestHasComponentType(t *testing.T) {
	app := &App{
		config: &AppConfig{},
	}

	// Test with nil components
	result := app.hasComponentType("test")
	if result {
		t.Error("Expected false for nil components")
	}

	// Test with matching component type
	app.components = []Component{
		&stubComponent{},
	}

	result = app.hasComponentType(&stubComponent{})
	if !result {
		t.Error("Expected true for matching component type")
	}

	// Test with non-matching component type
	type otherComponent struct{}
	result = app.hasComponentType(&otherComponent{})
	if result {
		t.Error("Expected false for non-matching component type")
	}
}

// TestAppLogger tests the Logger method
func TestAppLogger(t *testing.T) {
	app := New()

	// Logger should be initialized by New()
	logger := app.Logger()
	if logger == nil {
		t.Error("Expected logger to be returned")
	}
}

// TestBuiltInComponents tests the builtInComponents method
func TestBuiltInComponents(t *testing.T) {
	app := &App{
		config: &AppConfig{},
	}

	components := app.builtInComponents()

	// Should return empty slice for default config
	if len(components) != 0 {
		t.Errorf("Expected 0 built-in components, got %d", len(components))
	}

	// Test with pubsub debug enabled
	app.config.PubSub.Enabled = true
	app.config.PubSub.Path = "/debug"

	components = app.builtInComponents()
	if len(components) != 1 {
		t.Errorf("Expected 1 built-in component with pubsub debug, got %d", len(components))
	}

	// Test with webhook out enabled
	app.config.PubSub.Enabled = false
	app.config.WebhookOut.Enabled = true
	app.config.WebhookOut.Service = nil

	components = app.builtInComponents()
	if len(components) != 1 {
		t.Errorf("Expected 1 built-in component with webhook out, got %d", len(components))
	}

	// Test with webhook in enabled
	app.config.WebhookOut.Enabled = false
	app.config.WebhookIn.Enabled = true
	app.config.WebhookIn.Pub = nil

	components = app.builtInComponents()
	if len(components) != 1 {
		t.Errorf("Expected 1 built-in component with webhook in, got %d", len(components))
	}

	// Test with all enabled
	app.config.PubSub.Enabled = true
	app.config.WebhookOut.Enabled = true
	app.config.WebhookIn.Enabled = true

	components = app.builtInComponents()
	if len(components) != 3 {
		t.Errorf("Expected 3 built-in components, got %d", len(components))
	}
}

// TestServeHTTP tests the ServeHTTP method
func TestServeHTTP(t *testing.T) {
	app := &App{
		config: &AppConfig{},
	}

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// This should not panic even without handler setup
	app.ServeHTTP(w, req)
}

// TestEnsureHandler tests the ensureHandler method
func TestEnsureHandler(t *testing.T) {
	app := &App{
		config: &AppConfig{},
	}

	// This should not panic
	app.ensureHandler()
}
