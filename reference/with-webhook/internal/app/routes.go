package app

import (
	"net/http"

	"with-webhook/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the with-webhook demo.
func (a *App) RegisterRoutes() error {
	logger := a.Core.Logger()
	if err := a.Core.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.WriteHealthResponse(w, r, "with-webhook", logger)
	})); err != nil {
		return err
	}

	// Register inbound webhook routes (/webhooks/github, /webhooks/stripe).
	if err := a.Inbound.RegisterRoutes(a.Core); err != nil {
		return err
	}

	return nil
}
