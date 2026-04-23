package handler

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/contract"
)

type ReadinessFunc func(context.Context) error

type HealthHandler struct {
	ready ReadinessFunc
}

type HealthStatus struct {
	Status string `json:"status"`
}

func NewHealthHandler(ready ReadinessFunc) *HealthHandler {
	return &HealthHandler{ready: ready}
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, HealthStatus{Status: "ok"}, nil)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.ready != nil {
		if err := h.ready(r.Context()); err != nil {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeUnavailable).
				Code(contract.CodeUnavailable).
				Message("workerfleet readiness check failed").
				Build())
			return
		}
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, HealthStatus{Status: "ready"}, nil)
}
