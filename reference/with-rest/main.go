// Example: with-rest
//
// This demo adds x/rest resource controllers to a service that keeps the
// standard Plumego bootstrap and explicit route registration style.
package main

import (
	"log"

	"github.com/spcent/plumego/reference/with-rest/internal/app"
	"github.com/spcent/plumego/reference/with-rest/internal/config"
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

	log.Printf("Starting with-rest demo on %s", cfg.Core.Addr)
	if err := a.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
