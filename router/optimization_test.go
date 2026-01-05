package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

// BenchmarkOptimizedRouteMatching 测试优化后的路由匹配性能
func BenchmarkOptimizedRouteMatching(b *testing.B) {
	// 测试不同复杂度的路由
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

			// 注册路由
			for _, route := range tt.routes {
				r.AddRoute(route.method, route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			}

			// 创建测试请求
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

// BenchmarkWithCaching 测试带缓存的路由性能
func BenchmarkWithCaching(b *testing.B) {
	b.Run("WithoutCache", func(b *testing.B) {
		r := NewRouter()
		// Register routes only once
		for i := 0; i < 10; i++ {
			path := fmt.Sprintf("/api/v%d/users/:id", i)
			r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path := fmt.Sprintf("/api/v%d/users/%d", i%10, i)
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		r := NewRouterWithCache(100)
		// Register routes only once
		for i := 0; i < 10; i++ {
			path := fmt.Sprintf("/api/v%d/users/:id", i)
			r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path := fmt.Sprintf("/api/v%d/users/%d", i%10, i)
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// TestOptimizedMatcher 测试优化后的匹配器
func TestOptimizedMatcher(t *testing.T) {
	r := NewRouter()

	// 注册各种类型的路由
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

// TestSecurityEnhancements 测试安全增强
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

// TestRouterOptions 测试路由器选项
func TestRouterOptions(t *testing.T) {
	// Test default router
	r1 := NewRouter()
	if r1.enableCache {
		t.Error("default router should not have caching enabled")
	}

	// Test router with cache
	r2 := NewRouterWithCache(50)
	if !r2.enableCache {
		t.Error("router with cache should have caching enabled")
	}
	if r2.cacheSize != 50 {
		t.Errorf("expected cache size 50, got %d", r2.cacheSize)
	}

	// Test custom logger
	r3 := NewRouter(WithCache(10))
	if !r3.enableCache {
		t.Error("router should have caching enabled")
	}
}

// TestConcurrentSafety 测试并发安全性
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
