package app

import (
	"errors"
	"net/http"

	"github.com/spcent/plumego/core"
	"workerfleet/internal/handler"
)

func RegisterRoutes(app *core.App, workers *handler.Handler, metricsHandlers ...http.Handler) error {
	if app == nil {
		return errors.New("core app is required")
	}
	if workers == nil {
		return errors.New("workerfleet handler is required")
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
	if err := app.Get("/v1/tasks/:task_id", http.HandlerFunc(workers.GetTask)); err != nil {
		return err
	}
	if err := app.Get("/v1/fleet/summary", http.HandlerFunc(workers.FleetSummary)); err != nil {
		return err
	}
	if err := app.Get("/v1/alerts", http.HandlerFunc(workers.ListAlerts)); err != nil {
		return err
	}
	if len(metricsHandlers) > 0 && metricsHandlers[0] != nil {
		if err := app.Get("/metrics", metricsHandlers[0]); err != nil {
			return err
		}
	}
	return nil
}
