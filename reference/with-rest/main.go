// Example: with-rest
//
// This demo adds x/rest resource controllers to a service that keeps the
// standard Plumego bootstrap and explicit route registration style.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"with-rest/internal/app"
	"with-rest/internal/config"
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
