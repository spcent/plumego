// Package main is the process entrypoint for mini-saas-api.
// It does exactly four things: load config, construct the app, register routes, start the server.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mini-saas-api/internal/app"
	"mini-saas-api/internal/config"
)

// version is set at build time via -ldflags "-X main.version=1.0.0".
var version = "dev"

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
	cfg.App.Version = version

	a, err := app.New(cfg)
	if err != nil {
		return err
	}

	if err := a.RegisterRoutes(); err != nil {
		return err
	}

	return a.Start(ctx)
}
