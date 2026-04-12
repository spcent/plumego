package app

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// RegisterRoutes wires all HTTP routes for the with-gateway demo.
func (a *App) RegisterRoutes() error {
	if err := a.Core.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
			"status":    "ok",
			"service":   "with-gateway",
			"timestamp": time.Now().Format(time.RFC3339),
		}, nil)
	})); err != nil {
		return err
	}

	// Proxy all /proxy/* requests to the configured backend.
	if err := a.Core.Any("/proxy/*", http.HandlerFunc(a.Proxy.ServeHTTP)); err != nil {
		return err
	}

	return nil
}
