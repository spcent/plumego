package app

import (
	"net/http"

	"with-gateway/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the with-gateway demo.
func (a *App) RegisterRoutes() error {
	// HealthHandler receives no Checkers because the gateway has no application-level
	// dependencies to probe. In production, pass one health.ComponentChecker per
	// downstream dependency so /readyz reflects real backend availability.
	h := handler.HealthHandler{
		ServiceName: "with-gateway",
		Logger:      a.Core.Logger(),
	}

	if err := a.Core.Get("/healthz", http.HandlerFunc(h.Live)); err != nil {
		return err
	}
	if err := a.Core.Get("/readyz", http.HandlerFunc(h.Ready)); err != nil {
		return err
	}

	// Proxy all /proxy/* requests to the configured backend.
	if err := a.Core.Any("/proxy/*path", http.HandlerFunc(a.Proxy.ServeHTTP)); err != nil {
		return err
	}

	return nil
}
