package app

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// RegisterRoutes wires all HTTP routes for the with-messaging demo.
func (a *App) RegisterRoutes() error {
	a.Core.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
			"status":    "ok",
			"service":   "with-messaging",
			"timestamp": time.Now().Format(time.RFC3339),
		}, nil); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})

	a.Core.Post("/events/publish", a.Handler.Publish)

	return nil
}
