// Example: production-service
//
// Production-oriented reference application. The service keeps Plumego's
// canonical bootstrap shape while making security and observability middleware
// choices explicit in app-local wiring.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"production-service/internal/app"
	"production-service/internal/config"
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
