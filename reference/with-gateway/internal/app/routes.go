package app

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// RegisterRoutes wires all HTTP routes for the with-gateway demo.
func (a *App) RegisterRoutes() error {
	a.Core.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
			"status":    "ok",
			"service":   "with-gateway",
			"timestamp": time.Now().Format(time.RFC3339),
		}, nil); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})

	// Proxy all /proxy/* requests to the configured backend.
	a.Core.Any("/proxy/*", a.Proxy.ServeHTTP)

	return nil
}
