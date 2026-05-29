// Example: with-messaging
//
// Shows how to integrate x/messaging (pub/sub, queues, webhooks) into an
// existing service that follows the standard-service layout.
//
// Use reference/with-events for a service built around an event-driven
// architecture from the ground up.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"with-messaging/internal/app"
	"with-messaging/internal/config"
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
