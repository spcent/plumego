package core

import (
	"net/http/httptest"
	"testing"
)

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

	result = app.hasComponentType(&BaseComponent{})
	if result {
		t.Error("Expected false because component tracking has been removed")
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

// TestDeclaredComponents tests that core no longer tracks declared components.
func TestDeclaredComponents(t *testing.T) {
	app := &App{
		config: &AppConfig{},
	}

	components := app.declaredComponents()

	if len(components) != 0 {
		t.Errorf("Expected 0 declared components, got %d", len(components))
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
