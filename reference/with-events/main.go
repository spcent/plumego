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

	a, err := app.New(cfg, app.Deps{})
	if err != nil {
		return err
	}

	if err := a.RegisterRoutes(); err != nil {
		return err
	}

	return a.Start(ctx)
}
