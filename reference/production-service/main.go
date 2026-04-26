// Example: production-service
//
// Production-oriented reference application. The service keeps Plumego's
// canonical bootstrap shape while making security and observability middleware
// choices explicit in app-local wiring.
package main

import (
	"log"

	"github.com/spcent/plumego/reference/production-service/internal/app"
	"github.com/spcent/plumego/reference/production-service/internal/config"
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

	log.Printf("Starting Plumego Production Reference on %s", cfg.Core.Addr)
	if err := a.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
