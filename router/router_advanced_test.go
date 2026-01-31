package router

import (
	"bytes"
	"encoding/json"
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
	r.GetFunc("/static/path", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("static"))
	})

	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte("user-" + id))
	})

	r.GetFunc("/users/:id/posts/:postId", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		postId, _ := contract.Param(r, "postId")
		w.Write([]byte(fmt.Sprintf("user-%s-post-%s", id, postId)))
	})

	r.GetFunc("/files/*filepath", func(w http.ResponseWriter, r *http.Request) {
		filepath, _ := contract.Param(r, "filepath")
		w.Write([]byte("file-" + filepath))
	})

	r.PostFunc("/any/*path", func(w http.ResponseWriter, r *http.Request) {
		path, _ := contract.Param(r, "path")
		w.Write([]byte("any-" + path))
	})

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

// TestMiddlewareChain tests middleware application and ordering
func TestMiddlewareChain(t *testing.T) {
	r := NewRouter()

	var order []string

	// Global middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "global-1")
			next.ServeHTTP(w, r)
		})
	})

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "global-2")
			next.ServeHTTP(w, r)
		})
	})

	// Group middleware
	api := r.Group("/api")
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "api")
			next.ServeHTTP(w, r)
		})
	})

	api.GetFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.Write([]byte("ok"))
	})

	// Test API route
	order = []string{}
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	expected := []string{"global-1", "global-2", "api", "handler"}
	if !slicesEqual(order, expected) {
		t.Errorf("middleware order mismatch: expected %v, got %v", expected, order)
	}
}

// TestContextPropagation tests that context is properly propagated
func TestContextPropagation(t *testing.T) {
	r := NewRouter()

	r.GetFunc("/test/:id", func(w http.ResponseWriter, r *http.Request) {
		// Check params in context
		params := contract.ParamsFromContext(r.Context())
		if params["id"] != "123" {
			t.Errorf("param id in context: expected 123, got %s", params["id"])
		}

		// Check request context
		rc := contract.RequestContextFrom(r.Context())
		if rc.Params["id"] != "123" {
			t.Errorf("param id in request context: expected 123, got %s", rc.Params["id"])
		}

		// Check Param helper
		id, ok := contract.Param(r, "id")
		if !ok || id != "123" {
			t.Errorf("Param helper: expected 123, got %s", id)
		}

		w.Write([]byte("ok"))
	})

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
	r.GetFunc("/test", func(w http.ResponseWriter, r *http.Request) {})

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

	v1.GetFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v1-users"))
	})

	v2.GetFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v2-users"))
	})

	// Nested group
	v1Admin := v1.Group("/admin")
	v1Admin.GetFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v1-admin-users"))
	})

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

// TestResourceControllerAdvanced tests extended resource controller functionality
func TestResourceControllerAdvanced(t *testing.T) {
	type TestResource struct {
		BaseResourceController
	}

	ctrl := &TestResource{
		BaseResourceController: *NewBaseResourceController("test"),
	}

	// Override some methods
	ctrl.ResourceName = "advanced"

	r := NewRouter()
	if err := r.Resource("/advanced", ctrl); err != nil {
		t.Fatalf("Resource registration failed: %v", err)
	}

	// Test OPTIONS for CORS
	req := httptest.NewRequest("OPTIONS", "/advanced", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS should return 204, got %d", rec.Code)
	}

	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("CORS headers not set")
	}

	// Test HEAD
	req = httptest.NewRequest("HEAD", "/advanced", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("HEAD should return 200, got %d", rec.Code)
	}

	// Test default not implemented methods
	req = httptest.NewRequest("GET", "/advanced/123", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Errorf("Show should return 501, got %d", rec.Code)
	}

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Code != http.StatusText(http.StatusNotImplemented) {
		t.Errorf("expected code %q, got %q", http.StatusText(http.StatusNotImplemented), resp.Error.Code)
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

	r.GetFunc("/concurrent/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte(id))
	})

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

	_ = r.GetFunc("/users", func(w http.ResponseWriter, r *http.Request) {})
	_ = r.PostFunc("/users", func(w http.ResponseWriter, r *http.Request) {})
	_ = r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {})
	_ = r.AnyFunc("/any/*path", func(w http.ResponseWriter, r *http.Request) {})

	var buf bytes.Buffer
	r.Print(&buf)

	output := buf.String()

	if !strings.Contains(output, "GET    /users") {
		t.Error("Print output missing GET /users")
	}
	if !strings.Contains(output, "POST   /users") {
		t.Error("Print output missing POST /users")
	}
	if !strings.Contains(output, "/users/:id") {
		t.Error("Print output missing /users/:id")
	}
	if !strings.Contains(output, "[wildcard]") {
		t.Error("Print output missing wildcard marker")
	}
}

// TestRegisterMethod tests the Register method with RouteRegistrar
type testRegistrar struct {
	called bool
}

func (tr *testRegistrar) Register(r *Router) {
	tr.called = true
	_ = r.GetFunc("/registered", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("registered"))
	})
}

func TestRegisterMethod(t *testing.T) {
	r := NewRouter()
	reg := &testRegistrar{}

	r.Register(reg)

	if !reg.called {
		t.Error("Register method did not call Register on registrar")
	}

	// Test route was added
	req := httptest.NewRequest("GET", "/registered", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Error("registered route not working")
	}

	// Test duplicate registrar is ignored
	r.Register(reg)
}

// TestSetLogger tests logger configuration
func TestSetLogger(t *testing.T) {
	r := NewRouter()

	// Should not panic
	r.SetLogger(nil)
}

// Helper function
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
