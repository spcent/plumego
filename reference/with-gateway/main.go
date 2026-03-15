// Example: non-canonical
//
// This is a feature demo showing how to add x/gateway (reverse proxy)
// to a service that follows the standard-service layout.
//
// It is NOT the canonical app layout. See reference/standard-service for that.
package main

import (
	"log"

	"github.com/spcent/plumego/reference/with-gateway/internal/app"
	"github.com/spcent/plumego/reference/with-gateway/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := a.RegisterRoutes(); err != nil {
		log.Fatalf("failed to register routes: %v", err)
	}

	log.Printf("Starting with-gateway demo on %s (backend: %s)", cfg.Core.Addr, cfg.GatewayBackend)
	if err := a.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
