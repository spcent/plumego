// Example: with-ops
//
// This demo mounts protected x/observability/ops routes alongside stable request
// observability middleware. It follows the canonical 4-step bootstrap pattern:
// load config → build deps → register routes → start server.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"with-ops/internal/app"
	"with-ops/internal/config"
)

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

	a, err := app.New(cfg)
	if err != nil {
		return err
	}

	if err := a.RegisterRoutes(); err != nil {
		return err
	}

	return a.Start(ctx)
}
