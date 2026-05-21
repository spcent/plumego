// Example: non-canonical
//
// This is a feature demo showing how to add x/messaging/webhook inbound receivers
// to a service that follows the standard-service layout.
//
// It is NOT the canonical app layout. See reference/standard-service for that.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"with-webhook/internal/app"
	"with-webhook/internal/config"
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

	log.Printf("Starting with-webhook demo on %s", cfg.Core.Addr)
	return a.Start(ctx)
}
