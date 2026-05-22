// Example: with-tenant
//
// This demo adds x/tenant resolution, policy, quota, and rate limiting to a
// small API. The tenant middleware chain is applied per-route so only
// tenant-aware endpoints pay the cost. Follows the canonical 4-step bootstrap:
// load config → build deps → register routes → start server.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"with-tenant/internal/app"
	"with-tenant/internal/config"
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

	log.Printf("Starting with-tenant demo on %s", cfg.Core.Addr)
	return a.Start(ctx)
}
