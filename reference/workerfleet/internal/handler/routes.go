package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/spcent/plumego/core"
	workerapp "workerfleet/internal/app"
)

func RegisterServiceRoutes(app *core.App, service *workerapp.Service, ready func(context.Context) error, metrics http.Handler, workerAuth workerapp.WorkerIngressAuthConfig, adminAuth workerapp.AdminAuthConfig) error {
	return RegisterRoutes(app, New(service, WithWorkerIngressAuth(workerAuth), WithAdminAuth(adminAuth)), NewHealthHandler(ready), metrics)
}

func RegisterRoutes(app *core.App, workers *Handler, health *HealthHandler, metrics http.Handler) error {
	if err := validateRouteDependencies(app, workers, health); err != nil {
		return err
	}

	routes := []routeSpec{
		{method: http.MethodGet, path: "/healthz", handler: http.HandlerFunc(health.Live)},
		{method: http.MethodGet, path: "/readyz", handler: http.HandlerFunc(health.Ready)},
		{method: http.MethodPost, path: "/v1/workers/register", handler: http.HandlerFunc(workers.RegisterWorker)},
		{method: http.MethodPost, path: "/v1/workers/heartbeat", handler: http.HandlerFunc(workers.HeartbeatWorker)},
		{method: http.MethodGet, path: "/v1/workers", handler: http.HandlerFunc(workers.ListWorkers)},
		{method: http.MethodGet, path: "/v1/workers/:worker_id", handler: http.HandlerFunc(workers.GetWorker)},
		{method: http.MethodGet, path: "/v1/tasks/:task_id/timeline", handler: http.HandlerFunc(workers.GetCaseTimeline)},
		{method: http.MethodGet, path: "/v1/tasks/:task_id", handler: http.HandlerFunc(workers.GetTask)},
		{method: http.MethodGet, path: "/v1/exec-plans/:exec_plan_id/cases", handler: http.HandlerFunc(workers.ListExecPlanCases)},
		{method: http.MethodGet, path: "/v1/fleet/summary", handler: http.HandlerFunc(workers.FleetSummary)},
		{method: http.MethodGet, path: "/v1/alerts", handler: http.HandlerFunc(workers.ListAlerts)},
	}
	if metrics != nil {
		routes = append(routes, routeSpec{method: http.MethodGet, path: "/metrics", handler: metrics})
	}
	return registerRouteSpecs(app, routes)
}

type routeSpec struct {
	method  string
	path    string
	handler http.Handler
}

func validateRouteDependencies(app *core.App, workers *Handler, health *HealthHandler) error {
	if app == nil {
		return errors.New("core app is required")
	}
	if workers == nil {
		return errors.New("workerfleet handler is required")
	}
	if health == nil {
		return errors.New("workerfleet health handler is required")
	}
	return nil
}

func registerRouteSpecs(app *core.App, routes []routeSpec) error {
	for _, route := range routes {
		if err := app.AddRoute(route.method, route.path, route.handler); err != nil {
			return err
		}
	}
	return nil
}
