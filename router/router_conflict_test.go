package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRouteConflictDetection tests that routes with different parameter names
// at the same position are detected as conflicts
func TestRouteConflictDetection(t *testing.T) {
	r := NewRouter()

	// Register first route with parameter :id
	err := r.AddRoute("GET", "/user/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler1"))
	}))
	if err != nil {
		t.Fatalf("Failed to register first route: %v", err)
	}

	// Try to register second route with different parameter name :name at same position
	// This should fail because it conflicts with the first route
	err = r.AddRoute("GET", "/user/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler2"))
	}))

	// The error should be non-nil because this is a conflict
	if err == nil {
		t.Error("Expected error when registering conflicting route, but got nil")
	}

	// Verify the error message mentions conflict
	if err != nil && err.Error() != "" {
		t.Logf("Got expected error: %v", err)
	}
}

// TestSameParamNameAllowed tests that routes with the same parameter name
// at the same position are allowed (not a conflict) - but duplicate routes should still fail
func TestSameParamNameAllowed(t *testing.T) {
	r := NewRouter()

	// Register first route with parameter :id
	err := r.AddRoute("GET", "/user/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler1"))
	}))
	if err != nil {
		t.Fatalf("Failed to register first route: %v", err)
	}

	// Register second route with same parameter name :id at same position
	// This should fail because it's a duplicate route (same path and method)
	err = r.AddRoute("GET", "/user/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler2"))
	}))

	// This should fail because it's a duplicate route (same path and method)
	if err == nil {
		t.Error("Expected error when registering duplicate route, but got nil")
	}
}

// TestDifferentParamPositions tests that routes with parameters at different
// positions are allowed (not a conflict)
func TestDifferentParamPositions(t *testing.T) {
	r := NewRouter()

	// Register route with parameter at first position
	err := r.AddRoute("GET", "/:id/profile", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler1"))
	}))
	if err != nil {
		t.Fatalf("Failed to register first route: %v", err)
	}

	// Register route with parameter at second position
	// This should succeed because the parameter is at a different position
	err = r.AddRoute("GET", "/user/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler2"))
	}))

	if err != nil {
		t.Errorf("Expected no error for routes with parameters at different positions, but got: %v", err)
	}
}

// TestWildcardConflictDetection tests that wildcard routes are properly handled
func TestWildcardConflictDetection(t *testing.T) {
	r := NewRouter()

	// Register first route with wildcard
	err := r.AddRoute("GET", "/files/*path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler1"))
	}))
	if err != nil {
		t.Fatalf("Failed to register first route: %v", err)
	}

	// Try to register second wildcard at same position
	// This should fail because only one wildcard is allowed per parent
	err = r.AddRoute("GET", "/files/*filepath", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler2"))
	}))

	// This should fail because it's a duplicate route (same path and method)
	if err == nil {
		t.Error("Expected error when registering duplicate wildcard route, but got nil")
	}
}

// TestParamAfterStatic tests that parameter routes after static routes work correctly
func TestParamAfterStatic(t *testing.T) {
	r := NewRouter()

	// Register static route
	err := r.AddRoute("GET", "/user/list", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("list"))
	}))
	if err != nil {
		t.Fatalf("Failed to register static route: %v", err)
	}

	// Register parameter route
	err = r.AddRoute("GET", "/user/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("detail"))
	}))
	if err != nil {
		t.Fatalf("Failed to register parameter route: %v", err)
	}

	// Test that both routes work
	req1 := httptest.NewRequest("GET", "/user/list", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("Expected status 200 for static route, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest("GET", "/user/123", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status 200 for parameter route, got %d", rec2.Code)
	}
}

// TestMultipleParamsInPath tests routes with multiple parameters
func TestMultipleParamsInPath(t *testing.T) {
	r := NewRouter()

	// Register route with multiple parameters
	err := r.AddRoute("GET", "/users/:userId/posts/:postId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("post"))
	}))
	if err != nil {
		t.Fatalf("Failed to register route: %v", err)
	}

	// Try to register conflicting route with different parameter names
	err = r.AddRoute("GET", "/users/:uid/posts/:pid", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("conflict"))
	}))

	// This should fail because parameter names are different
	if err == nil {
		t.Error("Expected error when registering conflicting route with different param names, but got nil")
	}
}

// TestGroupWithParams tests parameter routes in groups
func TestGroupWithParams(t *testing.T) {
	r := NewRouter()
	api := r.Group("/api")

	// Register route in group with parameter
	err := api.AddRoute("GET", "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("user"))
	}))
	if err != nil {
		t.Fatalf("Failed to register route in group: %v", err)
	}

	// Try to register conflicting route in same group
	err = api.AddRoute("GET", "/users/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("conflict"))
	}))

	// This should fail
	if err == nil {
		t.Error("Expected error when registering conflicting route in group, but got nil")
	}
}

// TestConcurrentRouteRegistration tests concurrent route registration
func TestConcurrentRouteRegistration(t *testing.T) {
	r := NewRouter()

	// Register a base route
	err := r.AddRoute("GET", "/user/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("base"))
	}))
	if err != nil {
		t.Fatalf("Failed to register base route: %v", err)
	}

	// Try to register conflicting routes concurrently
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			err := r.AddRoute("GET", "/user/:param", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("conflict"))
			}))
			done <- err
		}(i)
	}

	// All should fail
	for i := 0; i < 10; i++ {
		err := <-done
		if err == nil {
			t.Errorf("Expected error for concurrent conflicting route registration, but got nil")
		}
	}
}
