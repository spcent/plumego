package router

// Benchmarks comparing plumego router overhead against stdlib http.ServeMux.
//
// Go 1.22+ ServeMux supports method-prefixed patterns and {param} extraction
// via r.PathValue(), making it a valid apples-to-apples comparison target.
//
// Run with:
//
//	go test -bench=BenchmarkCmp -benchmem -count=5 ./router/

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func newStdMux(patterns []string) *http.ServeMux {
	mux := http.NewServeMux()
	for _, p := range patterns {
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}
	return mux
}

func newPlumeRouter(method string, routes []string) *Router {
	r := NewRouter()
	for _, path := range routes {
		mustAddRoute(r, method, path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	}
	return r
}

// ── 1. Static route ──────────────────────────────────────────────────────────

func BenchmarkCmpStaticRoute_Stdlib(b *testing.B) {
	mux := newStdMux([]string{
		"GET /about",
		"GET /contact",
		"GET /api/v1/health",
	})
	req := httptest.NewRequest("GET", "/about", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkCmpStaticRoute_Plumego(b *testing.B) {
	r := newPlumeRouter("GET", []string{"/about", "/contact", "/api/v1/health"})
	req := httptest.NewRequest("GET", "/about", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

// ── 2. Single path parameter ──────────────────────────────────────────────────

func BenchmarkCmpSingleParam_Stdlib(b *testing.B) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		_ = r.PathValue("id")
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkCmpSingleParam_Plumego(b *testing.B) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = Param(req, "id")
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

// ── 3. Two path parameters ────────────────────────────────────────────────────

func BenchmarkCmpTwoParams_Stdlib(b *testing.B) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}/posts/{postId}", func(w http.ResponseWriter, r *http.Request) {
		_ = r.PathValue("id")
		_ = r.PathValue("postId")
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkCmpTwoParams_Plumego(b *testing.B) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users/:id/posts/:postId", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = Param(req, "id")
		_ = Param(req, "postId")
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

// ── 4. Four path parameters (deep nested resource) ────────────────────────────

func BenchmarkCmpDeepParams_Stdlib(b *testing.B) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/users/{userId}/orgs/{orgId}/teams/{teamId}/members/{memberId}",
		func(w http.ResponseWriter, r *http.Request) {
			_ = r.PathValue("userId")
			_ = r.PathValue("orgId")
			_ = r.PathValue("teamId")
			_ = r.PathValue("memberId")
			w.WriteHeader(http.StatusOK)
		})
	req := httptest.NewRequest("GET", "/api/v1/users/1/orgs/2/teams/3/members/4", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkCmpDeepParams_Plumego(b *testing.B) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/api/v1/users/:userId/orgs/:orgId/teams/:teamId/members/:memberId",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_ = Param(req, "userId")
			_ = Param(req, "orgId")
			_ = Param(req, "teamId")
			_ = Param(req, "memberId")
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

// ── 5. Wildcard route ─────────────────────────────────────────────────────────

func BenchmarkCmpWildcard_Stdlib(b *testing.B) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /static/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/static/css/main.css", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkCmpWildcard_Plumego(b *testing.B) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/static/*filepath", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = Param(req, "filepath")
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

// ── 6. Large route table (50 static + 20 param routes) ───────────────────────

func BenchmarkCmpManyRoutes_Stdlib(b *testing.B) {
	mux := http.NewServeMux()
	for i := 0; i < 50; i++ {
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/resource%d", i),
			func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	}
	for i := 0; i < 20; i++ {
		mux.HandleFunc(fmt.Sprintf("GET /api/v1/entity%d/{id}", i),
			func(w http.ResponseWriter, r *http.Request) {
				_ = r.PathValue("id")
				w.WriteHeader(http.StatusOK)
			})
	}
	paths := []string{
		"/api/v1/resource0",
		"/api/v1/resource25",
		"/api/v1/resource49",
		"/api/v1/entity0/123",
		"/api/v1/entity10/456",
		"/api/v1/entity19/789",
	}
	reqs := make([]*http.Request, len(paths))
	for i, p := range paths {
		reqs[i] = httptest.NewRequest("GET", p, nil)
	}
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		mux.ServeHTTP(w, reqs[i%len(reqs)])
	}
}

func BenchmarkCmpManyRoutes_Plumego(b *testing.B) {
	r := NewRouter()
	for i := 0; i < 50; i++ {
		mustAddRoute(r, "GET", fmt.Sprintf("/api/v1/resource%d", i),
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(http.StatusOK) }))
	}
	for i := 0; i < 20; i++ {
		mustAddRoute(r, "GET", fmt.Sprintf("/api/v1/entity%d/:id", i),
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				_ = Param(req, "id")
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
	reqs := make([]*http.Request, len(paths))
	for i, p := range paths {
		reqs[i] = httptest.NewRequest("GET", p, nil)
	}
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, reqs[i%len(reqs)])
	}
}

// ── 7. Parallel static (concurrent throughput) ────────────────────────────────

func BenchmarkCmpParallelStatic_Stdlib(b *testing.B) {
	mux := newStdMux([]string{"GET /about"})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/about", nil)
		w := httptest.NewRecorder()
		for pb.Next() {
			w.Body.Reset()
			mux.ServeHTTP(w, req)
		}
	})
}

func BenchmarkCmpParallelStatic_Plumego(b *testing.B) {
	r := newPlumeRouter("GET", []string{"/about"})
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

// ── 8. Parallel single param (concurrent throughput) ─────────────────────────

func BenchmarkCmpParallelParam_Stdlib(b *testing.B) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		_ = r.PathValue("id")
		w.WriteHeader(http.StatusOK)
	})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/users/42", nil)
		w := httptest.NewRecorder()
		for pb.Next() {
			w.Body.Reset()
			mux.ServeHTTP(w, req)
		}
	})
}

func BenchmarkCmpParallelParam_Plumego(b *testing.B) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_ = Param(req, "id")
		w.WriteHeader(http.StatusOK)
	}))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/users/42", nil)
		w := httptest.NewRecorder()
		for pb.Next() {
			w.Body.Reset()
			r.ServeHTTP(w, req)
		}
	})
}

// ── 9. Method routing overhead (plumego has explicit per-method trees) ─────────

func BenchmarkCmpMethodRouting_Stdlib(b *testing.B) {
	mux := http.NewServeMux()
	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
		m := method
		mux.HandleFunc(fmt.Sprintf("%s /api/resource/{id}", m),
			func(w http.ResponseWriter, r *http.Request) {
				_ = r.PathValue("id")
				w.WriteHeader(http.StatusOK)
			})
	}
	reqs := []*http.Request{
		httptest.NewRequest("GET", "/api/resource/1", nil),
		httptest.NewRequest("POST", "/api/resource/2", nil),
		httptest.NewRequest("PUT", "/api/resource/3", nil),
		httptest.NewRequest("DELETE", "/api/resource/4", nil),
		httptest.NewRequest("PATCH", "/api/resource/5", nil),
	}
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		mux.ServeHTTP(w, reqs[i%len(reqs)])
	}
}

func BenchmarkCmpMethodRouting_Plumego(b *testing.B) {
	r := NewRouter()
	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
		mustAddRoute(r, method, "/api/resource/:id",
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				_ = Param(req, "id")
				w.WriteHeader(http.StatusOK)
			}))
	}
	reqs := []*http.Request{
		httptest.NewRequest("GET", "/api/resource/1", nil),
		httptest.NewRequest("POST", "/api/resource/2", nil),
		httptest.NewRequest("PUT", "/api/resource/3", nil),
		httptest.NewRequest("DELETE", "/api/resource/4", nil),
		httptest.NewRequest("PATCH", "/api/resource/5", nil),
	}
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, reqs[i%len(reqs)])
	}
}
