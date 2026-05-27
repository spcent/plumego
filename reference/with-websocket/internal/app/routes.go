package app

import (
	"net/http"

	"with-websocket/internal/handler"
)

// RegisterRoutes wires all HTTP routes for the with-websocket demo.
func (a *App) RegisterRoutes() error {
	logger := a.Core.Logger()
	if err := a.Core.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.WriteHealthResponse(w, r, "with-websocket", logger)
	})); err != nil {
		return err
	}

	// Register WebSocket upgrade and broadcast routes via the ws.Server helper.
	if err := a.WS.RegisterRoutes(a.Core); err != nil {
		return err
	}

	return nil
}
