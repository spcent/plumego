// Example: canonical
//
// This is the Plumego reference application. It demonstrates the canonical
// application structure described in the style guide:
//   - main.go does only: load config, build deps, register routes, start server
//   - business/HTTP logic lives in internal sub-packages
//   - all routes use the standard http.HandlerFunc signature
package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/spcent/plumego/reference/standard-service/internal/app"
	"github.com/spcent/plumego/reference/standard-service/internal/config"
)

//go:embed ui/*
var rawFS embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Expose the ui/ subtree so internal packages see "templates/" directly.
	uiFS, err := fs.Sub(rawFS, "ui")
	if err != nil {
		log.Fatalf("failed to open embedded ui: %v", err)
	}

	a, err := app.New(cfg, uiFS)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := a.RegisterRoutes(); err != nil {
		log.Fatalf("failed to register routes: %v", err)
	}

	log.Printf("Starting Plumego Reference on %s", cfg.Core.Addr)
	if err := a.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
