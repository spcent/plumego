package router

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
)

// TestAdvancedRouteMatching tests complex route matching scenarios
func TestAdvancedRouteMatching(t *testing.T) {
	r := NewRouter()

	// Register various route types
	mustAddRoute(r, http.MethodGet, "/static/path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("static"))
	}))

	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		w.Write([]byte("user-" + id))
	}))

	mustAddRoute(r, http.MethodGet, "/users/:id/posts/:postId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		postId := Param(r, "postId")
		w.Write([]byte(fmt.Sprintf("user-%s-post-%s", id, postId)))
	}))

	mustAddRoute(r, http.MethodGet, "/files/*filepath", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filepath := Param(r, "filepath")
		w.Write([]byte("file-" + filepath))
	}))

	mustAddRoute(r, http.MethodPost, "/any/*path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := Param(r, "path")
		w.Write([]byte("any-" + path))
	}))

	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{"static path", "GET", "/static/path", "static"},
		{"single param", "GET", "/users/123", "user-123"},
		{"multiple params", "GET", "/users/456/posts/789", "user-456-post-789"},
		{"wildcard at end", "GET", "/files/docs/readme.txt", "file-docs/readme.txt"},
		{"wildcard with multiple segments", "GET", "/files/a/b/c.txt", "file-a/b/c.txt"},
		{"any method with wildcard", "POST", "/any/anything/here", "any-anything/here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}

			if body := strings.TrimSpace(rec.Body.String()); body != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, body)
			}
		})
	}
}

// TestContextPropagation tests that context is properly propagated
func TestContextPropagation(t *testing.T) {
	r := NewRouter()

	mustAddRoute(r, http.MethodGet, "/test/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check params in context
		params := contract.RequestContextFromContext(r.Context()).Params
		if params["id"] != "123" {
			t.Errorf("param id in context: expected 123, got %s", params["id"])
		}

		// Check request context
		rc := contract.RequestContextFromContext(r.Context())
		if rc.Params["id"] != "123" {
			t.Errorf("param id in request context: expected 123, got %s", rc.Params["id"])
		}

		// Check Param helper
		id := Param(r, "id")
		if id != "123" {
			t.Errorf("Param helper: expected 123, got %s", id)
		}

		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/test/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// TestRouteConflict tests duplicate route detection
func TestRouteConflict(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Should return error on duplicate
	err := r.AddRoute("GET", "/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	if err == nil {
		t.Errorf("expected error on duplicate route")
	}
}

// TestRouterGroupNesting tests nested groups
func TestRouterGroupNesting(t *testing.T) {
	r := NewRouter()

	v1 := r.Group("/api/v1")
	v2 := r.Group("/api/v2")

	mustAddRoute(v1, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v1-users"))
	}))

	mustAddRoute(v2, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v2-users"))
	}))

	// Nested group
	v1Admin := v1.Group("/admin")
	mustAddRoute(v1Admin, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v1-admin-users"))
	}))

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/users", "v1-users"},
		{"/api/v2/users", "v2-users"},
		{"/api/v1/admin/users", "v1-admin-users"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if body := strings.TrimSpace(rec.Body.String()); body != tt.expected {
				t.Errorf("path %s: expected %q, got %q", tt.path, tt.expected, body)
			}
		})
	}
}

// TestStaticFileSecurity tests security features of static file serving
func TestStaticFileSecurity(t *testing.T) {
	r := NewRouter()
	r.Static("/static", t.TempDir())

	// Test directory traversal attempt
	req := httptest.NewRequest("GET", "/static/../../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("directory traversal should return 404, got %d", rec.Code)
	}

	// Test path with special characters (skip null byte as it causes NewRequest to panic)
	req = httptest.NewRequest("GET", "/static/test%00.txt", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should handle gracefully
	if rec.Code != http.StatusNotFound {
		t.Logf("null byte handling: status %d", rec.Code)
	}
}

// TestConcurrentRequests tests router concurrency safety
func TestConcurrentRequests(t *testing.T) {
	r := NewRouter()

	mustAddRoute(r, http.MethodGet, "/concurrent/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		w.Write([]byte(id))
	}))

	done := make(chan bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		go func(id int) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/concurrent/%d", id), nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("concurrent request failed: status %d", rec.Code)
			}

			if strings.TrimSpace(rec.Body.String()) != fmt.Sprintf("%d", id) {
				t.Errorf("concurrent request failed: wrong response")
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < iterations; i++ {
		<-done
	}
}

// TestRouterPrint tests the Print method
func TestRouterPrint(t *testing.T) {
	r := NewRouter()

	mustAddRoute(r, http.MethodGet, "/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	mustAddRoute(r, http.MethodPost, "/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	mustAddRoute(r, methodAny, "/any/*path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	var buf bytes.Buffer
	r.Print(&buf)

	output := buf.String()

	assertOutputContains(t, output, "GET    /users")
	assertOutputContains(t, output, "POST   /users")
	assertOutputContains(t, output, "/users/:id")
	assertOutputContains(t, output, "[wildcard]")
}
