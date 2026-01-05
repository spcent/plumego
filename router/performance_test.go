package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

// BenchmarkRouterComparison compares performance between old and new router implementations
func BenchmarkRouterComparison(b *testing.B) {
	tests := []struct {
		name   string
		routes []struct{ method, path string }
		paths  []string
	}{
		{
			name: "Simple Static Routes",
			routes: []struct{ method, path string }{
				{"GET", "/users"},
				{"POST", "/users"},
				{"GET", "/posts"},
				{"GET", "/comments"},
			},
			paths: []string{"/users", "/posts", "/comments"},
		},
		{
			name: "Parameterized Routes",
			routes: []struct{ method, path string }{
				{"GET", "/users/:id"},
				{"GET", "/users/:id/posts/:postId"},
				{"POST", "/users/:id/posts"},
				{"GET", "/posts/:id/comments/:commentId"},
			},
			paths: []string{
				"/users/123",
				"/users/456/posts/789",
				"/posts/100/comments/200",
			},
		},
		{
			name: "Mixed Complex Routes",
			routes: []struct{ method, path string }{
				{"GET", "/"},
				{"GET", "/api/v1/users"},
				{"GET", "/api/v1/users/:id"},
				{"GET", "/api/v1/users/:id/posts/:postId"},
				{"POST", "/api/v1/posts"},
				{"GET", "/static/*filepath"},
				{"GET", "/health"},
				{"POST", "/webhook/:service"},
			},
			paths: []string{
				"/",
				"/api/v1/users",
				"/api/v1/users/123",
				"/api/v1/users/123/posts/456",
				"/static/css/main.css",
				"/health",
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Test with current optimized router
			b.Run("Optimized", func(b *testing.B) {
				r := NewRouter()
				for _, route := range tt.routes {
					r.AddRoute(route.method, route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}))
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					path := tt.paths[i%len(tt.paths)]
					req := httptest.NewRequest("GET", path, nil)
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
				}
			})

			// Test with caching enabled
			b.Run("WithCache", func(b *testing.B) {
				r := NewRouterWithCache(100)
				for _, route := range tt.routes {
					r.AddRoute(route.method, route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}))
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					path := tt.paths[i%len(tt.paths)]
					req := httptest.NewRequest("GET", path, nil)
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
				}
			})
		})
	}
}

// BenchmarkRadixTreePerformance tests radix tree routing performance
func BenchmarkRadixTreePerformance(b *testing.B) {
	// Create a radix tree directly
	rt := NewRadixTree()

	// Register routes
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/users/:id"},
		{"GET", "/users/:id/posts/:postId"},
		{"POST", "/users/:id/posts"},
		{"GET", "/posts/:id"},
		{"GET", "/static/*filepath"},
		{"GET", "/api/v1/users/:id/profile"},
		{"POST", "/api/v1/webhook/:service"},
	}

	for _, route := range routes {
		rt.Insert(route.method, route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), []string{}, nil)
	}

	paths := []string{
		"/users/123",
		"/users/456/posts/789",
		"/posts/100",
		"/static/css/main.css",
		"/api/v1/users/123/profile",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		result := rt.Find("GET", path)
		if result == nil {
			b.Fatal("Route not found")
		}
	}
}

// BenchmarkCachePerformance tests cache performance
func BenchmarkCachePerformance(b *testing.B) {
	cache := NewRouteCache(100)

	// Pre-populate cache
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("GET /users/%d", i)
		result := &MatchResult{
			Handler:     http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			ParamValues: []string{fmt.Sprintf("%d", i)},
			ParamKeys:   []string{"id"},
		}
		cache.Set(key, result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("GET /users/%d", i%50)
		if _, found := cache.Get(key); !found {
			b.Fatal("Cache miss")
		}
	}
}

// BenchmarkParameterValidation tests parameter validation overhead
func BenchmarkParameterValidation(b *testing.B) {
	r := NewRouter()

	// Add route with validation
	validation := NewRouteValidation().
		AddParam("id", PositiveIntValidator).
		AddParam("email", EmailValidator)

	r.AddValidation("GET", "/users/:id", validation)

	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test requests
	tests := []struct {
		name string
		path string
	}{
		{"Valid", "/users/123"},
		{"Invalid", "/users/abc"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest("GET", tt.path, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
			}
		})
	}
}

// TestOptimizedRouterFeatures validates all optimization features work together
func TestOptimizedRouterFeatures(t *testing.T) {
	// Create router with all features
	r := NewRouterWithCache(50)

	// Add validation
	validation := NewRouteValidation().
		AddParam("id", PositiveIntValidator)
	r.AddValidation("GET", "/users/:id", validation)

	// Register routes
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte("user-" + id))
	})

	r.GetFunc("/posts/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte("post-" + id))
	})

	// Test 1: Valid parameter
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "user-123" {
		t.Errorf("Expected 'user-123', got '%s'", body)
	}

	// Test 2: Invalid parameter (should fail validation)
	req = httptest.NewRequest("GET", "/users/abc", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid param, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Test 3: Cache hit on second request
	req = httptest.NewRequest("GET", "/posts/456", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Second request should be faster due to cache
	req = httptest.NewRequest("GET", "/posts/456", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 on cached request, got %d", w.Code)
	}
}

// TestRadixTreeMatching validates radix tree routing
func TestRadixTreeMatching(t *testing.T) {
	rt := NewRadixTree()

	// Register routes
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/users/:id"},
		{"GET", "/users/:id/posts/:postId"},
		{"GET", "/static/*filepath"},
	}

	for _, route := range routes {
		rt.Insert(route.method, route.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), []string{}, nil)
	}

	// Test cases
	tests := []struct {
		method   string
		path     string
		expected bool
	}{
		{"GET", "/users/123", true},
		{"GET", "/users/456/posts/789", true},
		{"GET", "/static/css/main.css", true},
		{"GET", "/users", false},
		{"POST", "/users/123", false},
	}

	for _, tt := range tests {
		result := rt.Find(tt.method, tt.path)
		found := result != nil
		if found != tt.expected {
			t.Errorf("Find(%s, %s): expected %v, got %v", tt.method, tt.path, tt.expected, found)
		}
	}
}

// TestCacheEviction validates LRU cache eviction
func TestCacheEviction(t *testing.T) {
	cache := NewRouteCache(3) // Small capacity

	// Add 3 entries
	cache.Set("key1", &MatchResult{Handler: nil})
	cache.Set("key2", &MatchResult{Handler: nil})
	cache.Set("key3", &MatchResult{Handler: nil})

	// Access key1 to make it most recently used
	cache.Get("key1")

	// Add 4th entry, should evict key2 (least recently used)
	cache.Set("key4", &MatchResult{Handler: nil})

	// Verify key1 and key3 still exist
	if _, found := cache.Get("key1"); !found {
		t.Error("key1 should still exist")
	}
	if _, found := cache.Get("key3"); !found {
		t.Error("key3 should still exist")
	}
	if _, found := cache.Get("key4"); !found {
		t.Error("key4 should exist")
	}

	// key2 should be evicted
	if _, found := cache.Get("key2"); found {
		t.Error("key2 should have been evicted")
	}
}

// TestValidationRules validates parameter validation
func TestValidationRules(t *testing.T) {
	validation := NewRouteValidation().
		AddParam("id", PositiveIntValidator).
		AddParam("email", EmailValidator).
		AddParam("name", NewLengthValidator(1, 50))

	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{
			name:    "Valid all",
			params:  map[string]string{"id": "123", "email": "test@example.com", "name": "John"},
			wantErr: false,
		},
		{
			name:    "Invalid id",
			params:  map[string]string{"id": "abc", "email": "test@example.com"},
			wantErr: true,
		},
		{
			name:    "Invalid email",
			params:  map[string]string{"id": "123", "email": "invalid-email"},
			wantErr: true,
		},
		{
			name:    "Name too long",
			params:  map[string]string{"id": "123", "name": string(make([]byte, 51))},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.Validate(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
