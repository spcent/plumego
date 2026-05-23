// Example: with-ai
//
// This demo wires the stable-tier x/ai subpackages with offline behavior:
// provider, session, streaming, and tool. It follows the canonical 4-step
// bootstrap pattern: load config → build deps → register routes → start server.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"with-ai/internal/app"
	"with-ai/internal/config"
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
