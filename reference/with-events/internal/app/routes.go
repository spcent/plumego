package app

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

type placeholderResponse struct {
	Status string `json:"status"`
	Group  string `json:"group"`
}

// RegisterRoutes wires placeholder route groups for later event-flow cards.
func (a *App) RegisterRoutes() error {
	if err := a.Core.Post("/orders", http.HandlerFunc(a.Orders.Create)); err != nil {
		return err
	}
	if err := a.Core.Get("/orders/:id", http.HandlerFunc(a.Orders.Get)); err != nil {
		return err
	}
	if err := a.Core.Post("/scheduler/retry", placeholder("scheduler")); err != nil {
		return err
	}
	if err := a.Core.Post("/webhook/send", placeholder("webhook")); err != nil {
		return err
	}
	return nil
}

func placeholder(group string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = contract.WriteResponse(w, r, http.StatusOK, placeholderResponse{
			Status: "placeholder",
			Group:  group,
		}, nil)
	}
}
