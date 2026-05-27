package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/core"
	workerapp "workerfleet/internal/app"
	workerfleetmetrics "workerfleet/internal/platform/metrics"
)

func TestRegisterRoutesWiresMetricsWhenProvided(t *testing.T) {
	app := core.New(core.DefaultConfig(), core.AppDependencies{})
	metrics := workerfleetmetrics.NewCollector()
	if err := metrics.SetGauge(workerfleetmetrics.MetricWorkersTotal, map[string]string{
		workerfleetmetrics.LabelStatus: "online",
	}, 1); err != nil {
		t.Fatalf("set gauge: %v", err)
	}

	if err := RegisterRoutes(app, RouteDependencies{
		Workers: New(nil),
		Health:  NewHealthHandler(nil, discardLogger()),
		Metrics: metrics.Handler(),
	}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "workerfleet_workers") {
		t.Fatalf("body missing metrics output: %s", rec.Body.String())
	}
}

func TestRegisterRoutesWiresHealthAndReadiness(t *testing.T) {
	app := core.New(core.DefaultConfig(), core.AppDependencies{})
	if err := RegisterRoutes(app, RouteDependencies{Workers: New(nil), Health: NewHealthHandler(nil, discardLogger())}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	for _, path := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		app.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, rec.Code)
		}
	}
}

func TestRegisterRoutesWiresDrilldownRoutes(t *testing.T) {
	app := core.New(core.DefaultConfig(), core.AppDependencies{})
	if err := RegisterRoutes(app, RouteDependencies{Workers: New(nil), Health: NewHealthHandler(nil, discardLogger())}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	for _, path := range []string{"/v1/tasks/case-1/timeline", "/v1/exec-plans/plan-1/cases"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		app.ServeHTTP(rec, req)
		if rec.Code == http.StatusNotFound {
			t.Fatalf("%s was not routed", path)
		}
	}
}

func TestQueryRoutesRequireAdminAuthWhenConfigured(t *testing.T) {
	app := core.New(core.DefaultConfig(), core.AppDependencies{})
	workers := New(nil, WithAdminAuth(workerapp.AdminAuthConfig{
		Token:    "admin-secret",
		Required: true,
	}))
	if err := RegisterRoutes(app, RouteDependencies{Workers: workers, Health: NewHealthHandler(nil, discardLogger())}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	queryPaths := []string{
		"/v1/workers",
		"/v1/workers/worker-1",
		"/v1/tasks/task-1/timeline",
		"/v1/tasks/task-1",
		"/v1/exec-plans/plan-1/cases",
		"/v1/fleet/summary",
		"/v1/alerts",
	}
	for _, path := range queryPaths {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		app.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s status = %d, want 401", path, rec.Code)
		}

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer admin-secret")
		app.ServeHTTP(rec, req)
		if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusNotFound {
			t.Fatalf("%s authorized status = %d, want routed past auth", path, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	app.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", rec.Code)
	}
}
