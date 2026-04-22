package app

import (
	"errors"
	"net/http"

	"github.com/spcent/plumego/core"
	"workerfleet/internal/handler"
)

func RegisterRoutes(app *core.App, workers *handler.Handler, health *handler.HealthHandler, metrics http.Handler) error {
	if app == nil {
		return errors.New("core app is required")
	}
	if workers == nil {
		return errors.New("workerfleet handler is required")
	}
	if health == nil {
		return errors.New("workerfleet health handler is required")
	}

	if err := app.Get("/healthz", http.HandlerFunc(health.Live)); err != nil {
		return err
	}
	if err := app.Get("/readyz", http.HandlerFunc(health.Ready)); err != nil {
		return err
	}
	if err := app.Post("/v1/workers/register", http.HandlerFunc(workers.RegisterWorker)); err != nil {
		return err
	}
	if err := app.Post("/v1/workers/heartbeat", http.HandlerFunc(workers.HeartbeatWorker)); err != nil {
		return err
	}
	if err := app.Get("/v1/workers", http.HandlerFunc(workers.ListWorkers)); err != nil {
		return err
	}
	if err := app.Get("/v1/workers/:worker_id", http.HandlerFunc(workers.GetWorker)); err != nil {
		return err
	}
	if err := app.Get("/v1/tasks/:task_id/timeline", http.HandlerFunc(workers.GetCaseTimeline)); err != nil {
		return err
	}
	if err := app.Get("/v1/tasks/:task_id", http.HandlerFunc(workers.GetTask)); err != nil {
		return err
	}
	if err := app.Get("/v1/exec-plans/:exec_plan_id/cases", http.HandlerFunc(workers.ListExecPlanCases)); err != nil {
		return err
	}
	if err := app.Get("/v1/fleet/summary", http.HandlerFunc(workers.FleetSummary)); err != nil {
		return err
	}
	if err := app.Get("/v1/alerts", http.HandlerFunc(workers.ListAlerts)); err != nil {
		return err
	}
	if metrics != nil {
		if err := app.Get("/metrics", metrics); err != nil {
			return err
		}
	}
	return nil
}
