// Example: with-observability
//
// This is the Plumego observability reference application. It extends
// standard-service with production-grade HTTP metrics (Prometheus exposition
// format) and distributed tracing (OpenTelemetry span collection).
//
// Key differences from standard-service:
//   - httpmetrics.Middleware is wired with a real PrometheusCollector (not noop)
//   - tracing.Middleware is wired with an OpenTelemetryTracer
//   - GET /metrics exposes Prometheus counters, latency summaries, and uptime
//   - GET /api/v1/spans exposes recent spans for debugging (remove in production)
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"with-observability/internal/app"
	"with-observability/internal/config"
)

// version is set at build time via -ldflags "-X main.version=1.0.0".
var version = "dev"

func main() {
	if err := run(); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.App.Version = version

	a, err := app.New(cfg)
	if err != nil {
		return err
	}

	if err := a.RegisterRoutes(); err != nil {
		return err
	}

	return a.Start(ctx)
}
