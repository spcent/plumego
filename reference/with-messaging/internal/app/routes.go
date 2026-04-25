package app

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// RegisterRoutes wires all HTTP routes for the with-messaging demo.
func (a *App) RegisterRoutes() error {
	if err := a.Core.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeHealthResponse(w, r, "with-messaging")
	})); err != nil {
		return err
	}

	if err := a.Core.Post("/events/publish", http.HandlerFunc(a.Handler.Publish)); err != nil {
		return err
	}

	return nil
}

func writeHealthResponse(w http.ResponseWriter, r *http.Request, service string) {
	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   service,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil)
}
