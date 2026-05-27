package app

import (
	"net/http"

	"with-gateway/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the with-gateway demo.
func (a *App) RegisterRoutes() error {
	logger := a.Core.Logger()
	if err := a.Core.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.WriteHealthResponse(w, r, "with-gateway", logger)
	})); err != nil {
		return err
	}

	// Proxy all /proxy/* requests to the configured backend.
	if err := a.Core.Any("/proxy/*path", http.HandlerFunc(a.Proxy.ServeHTTP)); err != nil {
		return err
	}

	return nil
}
