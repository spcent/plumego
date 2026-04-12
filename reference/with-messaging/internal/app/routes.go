package app

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// RegisterRoutes wires all HTTP routes for the with-messaging demo.
func (a *App) RegisterRoutes() error {
	if err := a.Core.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
			"status":    "ok",
			"service":   "with-messaging",
			"timestamp": time.Now().Format(time.RFC3339),
		}, nil)
	})); err != nil {
		return err
	}

	if err := a.Core.Post("/events/publish", http.HandlerFunc(a.Handler.Publish)); err != nil {
		return err
	}

	return nil
}
