package benchmark_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/spcent/plumego/router"
)

func BenchmarkRouterStatic(b *testing.B) {
	benchmarkRouteMatch(b, routeBenchCase{
		name:        "static",
		requestPath: "/assets/health",
		plumegoPath: "/assets/health",
		chiPath:     "/assets/health",
	})
}

func BenchmarkRouterSingleParam(b *testing.B) {
	benchmarkRouteMatch(b, routeBenchCase{
		name:        "single-param",
		requestPath: "/users/12345",
		plumegoPath: "/users/:id",
		chiPath:     "/users/{id}",
	})
}

func BenchmarkRouterMultiParam(b *testing.B) {
	benchmarkRouteMatch(b, routeBenchCase{
		name:        "multi-param",
		requestPath: "/orgs/acme/users/42/repos/plumego",
		plumegoPath: "/orgs/:org/users/:userID/repos/:repo",
		chiPath:     "/orgs/{org}/users/{userID}/repos/{repo}",
	})
}

type routeBenchCase struct {
	name        string
	requestPath string
	plumegoPath string
	chiPath     string
}

func benchmarkRouteMatch(b *testing.B, tc routeBenchCase) {
	b.Helper()

	b.Run("plumego", func(b *testing.B) {
		r := router.NewRouter()
		mustAddBenchmarkRoute(b, r, http.MethodGet, tc.plumegoPath, benchmarkNoOpHandler())
		runHandlerBenchmark(b, r, tc.requestPath)
	})

	b.Run("chi", func(b *testing.B) {
		r := chi.NewRouter()
		r.Method(http.MethodGet, tc.chiPath, benchmarkNoOpHandler())
		runHandlerBenchmark(b, r, tc.requestPath)
	})
}

func runHandlerBenchmark(b *testing.B, h http.Handler, requestPath string) {
	b.Helper()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, requestPath, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			b.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusNoContent)
		}
	}
}

func benchmarkNoOpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

func mustAddBenchmarkRoute(b *testing.B, r *router.Router, method, path string, h http.Handler) {
	b.Helper()
	if err := r.AddRoute(method, path, h); err != nil {
		b.Fatalf("add route %s %s: %v", method, path, err)
	}
}
