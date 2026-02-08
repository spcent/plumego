package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Benchmarks for optimized router performance.
// Run with: go test -bench=BenchmarkOpt -benchmem -count=3 ./router/

// BenchmarkOptStaticRoute benchmarks static route matching (no params).
func BenchmarkOptStaticRoute(b *testing.B) {
	r := NewRouter()
	r.AddRoute(GET, "/about", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	r.AddRoute(GET, "/contact", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	r.AddRoute(GET, "/api/v1/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/about", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkOptSingleParam benchmarks single-parameter route matching.
func BenchmarkOptSingleParam(b *testing.B) {
	r := NewRouter()
	r.AddRoute(GET, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkOptMultiParam benchmarks multi-parameter route matching.
func BenchmarkOptMultiParam(b *testing.B) {
	r := NewRouter()
	r.AddRoute(GET, "/users/:id/posts/:postId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkOptWildcard benchmarks wildcard route matching.
func BenchmarkOptWildcard(b *testing.B) {
	r := NewRouter()
	r.AddRoute(GET, "/static/*filepath", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/static/css/main.css", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkOptManyRoutes benchmarks matching in a router with many registered routes.
func BenchmarkOptManyRoutes(b *testing.B) {
	r := NewRouter()

	// Register 50 static routes
	for i := 0; i < 50; i++ {
		path := fmt.Sprintf("/api/v1/resource%d", i)
		r.AddRoute(GET, path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	}
	// Register 20 parameterized routes
	for i := 0; i < 20; i++ {
		path := fmt.Sprintf("/api/v1/entity%d/:id", i)
		r.AddRoute(GET, path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	}

	paths := []string{
		"/api/v1/resource0",
		"/api/v1/resource25",
		"/api/v1/resource49",
		"/api/v1/entity0/123",
		"/api/v1/entity10/456",
		"/api/v1/entity19/789",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkOptDeepPath benchmarks deeply nested path matching.
func BenchmarkOptDeepPath(b *testing.B) {
	r := NewRouter()
	r.AddRoute(GET, "/api/v1/users/:userId/orgs/:orgId/teams/:teamId/members/:memberId",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

	req := httptest.NewRequest("GET", "/api/v1/users/1/orgs/2/teams/3/members/4", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkOptPatternCache benchmarks pattern cache lookups specifically.
func BenchmarkOptPatternCache(b *testing.B) {
	cache := NewRouteCache(100)

	// Pre-populate pattern cache
	patterns := []string{
		"/users/:id",
		"/users/:id/posts/:postId",
		"/api/v1/users/:id/profile",
		"/teams/:teamId/members/:memberId",
	}
	for _, p := range patterns {
		cache.SetPattern(GET, p, &MatchResult{
			Handler:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			ParamKeys: []string{"id"},
		})
	}

	lookupPaths := []string{
		"/users/123",
		"/users/456/posts/789",
		"/api/v1/users/100/profile",
		"/teams/t1/members/m2",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := lookupPaths[i%len(lookupPaths)]
		cache.GetByPattern(GET, path)
	}
}

// BenchmarkOptNormalizePath benchmarks path normalization.
func BenchmarkOptNormalizePath(b *testing.B) {
	paths := []string{
		"/",
		"/users",
		"/users/123",
		"/api/v1/users/123/posts/456",
		"/static/css/main.css",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := paths[i%len(paths)]
		fastNormalizePath(p)
	}
}

// BenchmarkOptCacheKey benchmarks cache key construction.
func BenchmarkOptCacheKey(b *testing.B) {
	methods := []string{"GET", "POST", "PUT"}
	paths := []string{"/users", "/users/123", "/api/v1/health"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := methods[i%len(methods)]
		p := paths[i%len(paths)]
		fastBuildCacheKey(m, p)
	}
}

// BenchmarkOptSplitPath benchmarks fast path splitting.
func BenchmarkOptSplitPath(b *testing.B) {
	paths := []string{
		"/users/123",
		"/api/v1/users/123/posts/456",
		"/static/css/main.css",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := paths[i%len(paths)]
		bufPtr := pathBufPool.Get().(*[]string)
		buf := *bufPtr
		fastSplitPath(p, buf)
		*bufPtr = buf
		pathBufPool.Put(bufPtr)
	}
}

// BenchmarkOptParallelStatic benchmarks static route matching under contention.
func BenchmarkOptParallelStatic(b *testing.B) {
	r := NewRouter()
	r.AddRoute(GET, "/about", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/about", nil)
		w := httptest.NewRecorder()
		for pb.Next() {
			w.Body.Reset()
			r.ServeHTTP(w, req)
		}
	})
}

// BenchmarkOptParallelParam benchmarks parameterized route matching under contention.
func BenchmarkOptParallelParam(b *testing.B) {
	r := NewRouter()
	r.AddRoute(GET, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/users/123", nil)
		w := httptest.NewRecorder()
		for pb.Next() {
			w.Body.Reset()
			r.ServeHTTP(w, req)
		}
	})
}
