package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

// BenchmarkOptimizedRouteMatching tests optimized route matching performance
func BenchmarkOptimizedRouteMatching(b *testing.B) {
	// Test routes of different complexities
	tests := []struct {
		name   string
		routes []struct{ method, path string }
	}{
		{
			name: "Simple Static Routes",
			routes: []struct{ method, path string }{
				{"GET", "/users"},
				{"POST", "/users"},
				{"GET", "/posts"},
			},
		},
		{
			name: "Parameterized Routes",
			routes: []struct{ method, path string }{
				{"GET", "/users/:id"},
				{"GET", "/users/:id/posts/:postId"},
				{"POST", "/users/:id/posts"},
			},
		},
		{
			name: "Mixed Routes",
			routes: []struct{ method, path string }{
				{"GET", "/"},
				{"GET", "/api/v1/users"},
				{"GET", "/api/v1/users/:id"},
				{"GET", "/api/v1/users/:id/posts/:postId"},
				{"POST", "/api/v1/posts"},
				{"GET", "/static/*filepath"},
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			r := NewRouter()

			// Register routes
			for _, route := range tt.routes {
				r.AddRoute(route.method, route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			}

			// Create test requests
			var testPaths []string
			switch tt.name {
			case "Simple Static Routes":
				testPaths = []string{"/users", "/posts"}
			case "Parameterized Routes":
				testPaths = []string{"/users/123", "/users/456/posts/789"}
			case "Mixed Routes":
				testPaths = []string{"/", "/api/v1/users", "/api/v1/users/123", "/api/v1/users/123/posts/456"}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				path := testPaths[i%len(testPaths)]
				req := httptest.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
			}
		})
	}
}

// TestOptimizedMatcher tests the optimized matcher
func TestOptimizedMatcher(t *testing.T) {
	r := NewRouter()

	// Register various types of routes
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

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"static path", "/static/path", "static"},
		{"single param", "/users/123", "user-123"},
		{"multiple params", "/users/456/posts/789", "user-456-post-789"},
		{"wildcard", "/files/docs/readme.txt", "file-docs/readme.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}

			if body := rec.Body.String(); body != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, body)
			}
		})
	}
}

// TestSecurityEnhancements tests security enhancements
func TestSecurityEnhancements(t *testing.T) {
	r := NewRouter()
	r.Static("/static", t.TempDir())

	tests := []struct {
		name         string
		path         string
		expectStatus int
	}{
		{"normal file", "/static/test.txt", http.StatusNotFound},
		{"directory traversal 1", "/static/../../../etc/passwd", http.StatusNotFound},
		{"directory traversal 2", "/static/../../etc/passwd", http.StatusNotFound},
		{"null byte", "/static/test%00.txt", http.StatusNotFound},
		{"absolute path", "/static//etc/passwd", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectStatus {
				t.Errorf("path %s: expected status %d, got %d", tt.path, tt.expectStatus, rec.Code)
			}
		})
	}
}

// TestConcurrentSafety tests concurrent safety
func TestConcurrentSafety(t *testing.T) {
	r := NewRouter()

	// Register routes concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			path := fmt.Sprintf("/concurrent/%d", id)
			r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(fmt.Sprintf("handler-%d", id)))
			}))
			done <- true
		}(i)
	}

	// Wait for all registrations
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent requests
	for i := 0; i < 100; i++ {
		go func(id int) {
			path := fmt.Sprintf("/concurrent/%d", id%10)
			req := httptest.NewRequest("GET", path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("concurrent request failed: status %d", rec.Code)
			}
		}(i)
	}
}
