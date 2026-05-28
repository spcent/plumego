package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dbadmin/internal/app"
	"dbadmin/internal/config"
)

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
