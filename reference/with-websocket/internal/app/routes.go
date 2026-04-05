package app

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// RegisterRoutes wires all HTTP routes for the with-websocket demo.
func (a *App) RegisterRoutes() error {
	if err := a.Core.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
			"status":    "ok",
			"service":   "with-websocket",
			"timestamp": time.Now().Format(time.RFC3339),
		}, nil); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}); err != nil {
		return err
	}

	// Register WebSocket upgrade and broadcast routes via the ws.Server helper.
	if err := a.WS.RegisterRoutes(a.Core); err != nil {
		return err
	}

	return nil
}
