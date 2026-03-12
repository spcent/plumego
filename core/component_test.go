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

// TestDeclaredComponents tests explicit component assembly.
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
