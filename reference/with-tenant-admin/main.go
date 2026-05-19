package main

import (
	"log"

	"with-tenant-admin/internal/app"
	"with-tenant-admin/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	a, err := app.New(cfg, app.Deps{})
	if err != nil {
		log.Fatalf("initialize app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		log.Fatalf("register routes: %v", err)
	}

	log.Printf("Starting with-tenant-admin on %s", cfg.Addr)
	if err := a.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
