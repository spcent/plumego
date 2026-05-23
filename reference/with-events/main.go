// Example: with-events
//
// This scenario reference shows an in-process event-driven service built from
// core.App and x/messaging.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"with-events/internal/app"
	"with-events/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	a, err := app.New(cfg, app.Deps{})
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := a.RegisterRoutes(); err != nil {
		log.Fatalf("failed to register routes: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.Start(ctx); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
