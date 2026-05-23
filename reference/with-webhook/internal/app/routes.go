package app

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

type healthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
}

// RegisterRoutes wires all HTTP routes for the with-webhook demo.
func (a *App) RegisterRoutes() error {
	if err := a.Core.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeHealthResponse(w, r, "with-webhook")
	})); err != nil {
		return err
	}

	// Register inbound webhook routes (/webhooks/github, /webhooks/stripe).
	if err := a.Inbound.RegisterRoutes(a.Core); err != nil {
		return err
	}

	return nil
}

func writeHealthResponse(w http.ResponseWriter, r *http.Request, service string) {
	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   service,
		Timestamp: time.Now().UTC(),
	}, nil)
}
