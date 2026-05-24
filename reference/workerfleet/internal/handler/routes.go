package handler

import (
	"errors"
	"net/http"

	"github.com/spcent/plumego/core"
	workerapp "workerfleet/internal/app"
)

func RegisterServiceRoutes(app *core.App, deps workerapp.RouteDependencies) error {
	return RegisterRoutes(app, RouteDependencies{
		Workers: New(deps.Service, WithWorkerIngressAuth(deps.WorkerAuth), WithAdminAuth(deps.AdminAuth)),
		Health:  NewHealthHandler(deps.Ready),
		Metrics: deps.Metrics,
	})
}

type RouteDependencies struct {
	Workers *Handler
	Health  *HealthHandler
	Metrics http.Handler
}

func RegisterRoutes(app *core.App, deps RouteDependencies) error {
	if err := validateRouteDependencies(app, deps.Workers, deps.Health); err != nil {
		return err
	}

	routes := []routeSpec{
		{method: http.MethodGet, path: "/healthz", handler: http.HandlerFunc(deps.Health.Live)},
		{method: http.MethodGet, path: "/readyz", handler: http.HandlerFunc(deps.Health.Ready)},
		{method: http.MethodPost, path: "/v1/workers/register", handler: http.HandlerFunc(deps.Workers.RegisterWorker)},
		{method: http.MethodPost, path: "/v1/workers/heartbeat", handler: http.HandlerFunc(deps.Workers.HeartbeatWorker)},
		{method: http.MethodGet, path: "/v1/workers", handler: http.HandlerFunc(deps.Workers.ListWorkers)},
		{method: http.MethodGet, path: "/v1/workers/:worker_id", handler: http.HandlerFunc(deps.Workers.GetWorker)},
		{method: http.MethodGet, path: "/v1/tasks/:task_id/timeline", handler: http.HandlerFunc(deps.Workers.GetCaseTimeline)},
		{method: http.MethodGet, path: "/v1/tasks/:task_id", handler: http.HandlerFunc(deps.Workers.GetTask)},
		{method: http.MethodGet, path: "/v1/exec-plans/:exec_plan_id/cases", handler: http.HandlerFunc(deps.Workers.ListExecPlanCases)},
		{method: http.MethodGet, path: "/v1/fleet/summary", handler: http.HandlerFunc(deps.Workers.FleetSummary)},
		{method: http.MethodGet, path: "/v1/alerts", handler: http.HandlerFunc(deps.Workers.ListAlerts)},
	}
	if deps.Metrics != nil {
		routes = append(routes, routeSpec{method: http.MethodGet, path: "/metrics", handler: deps.Metrics})
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
