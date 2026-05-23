// Example: canonical
//
// This is the Plumego reference application. It demonstrates the canonical
// minimal application structure described in the style guide:
//   - main.go does only: load config, build deps, register routes, start server
//   - business/HTTP logic lives in internal sub-packages
//   - all routes use the standard http.HandlerFunc signature
//   - the reference depends only on stable root packages
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"standard-service/internal/app"
	"standard-service/internal/config"
)

// version is set at build time via -ldflags "-X main.version=1.0.0".
// It is threaded into config so handlers receive it through constructor injection
// rather than reading a package-level variable directly.
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
