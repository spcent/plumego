package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/app"
	"guardus/internal/config"
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

	logger := plumelog.NewLogger()
	cfg, err := config.Load(logger)
	if err != nil {
		return err
	}
	cfg.Version = version

	a, err := app.New(ctx, &cfg)
	if err != nil {
		return err
	}
	if err := a.RegisterRoutes(); err != nil {
		return err
	}
	return a.Start(ctx)
}
