// Example: with-frontend
//
// This demo mounts a static SPA alongside a JSON API using x/frontend.
// Assets can be served from a filesystem directory or an embedded fs.FS.
// Follows the canonical 4-step bootstrap: load config → build deps →
// register routes → start server.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"with-frontend/internal/app"
	"with-frontend/internal/config"
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

	log.Printf("Starting with-frontend demo on %s", cfg.Core.Addr)
	return a.Start(ctx)
}
