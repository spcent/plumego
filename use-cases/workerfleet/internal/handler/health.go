package handler

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

type ReadinessFunc func(context.Context) error

// HealthHandler serves the liveness and readiness probes.
// Logger must not be nil; pass app.Logger() from RegisterServiceRoutes.
type HealthHandler struct {
	ready  ReadinessFunc
	Logger plumelog.StructuredLogger
}

type HealthStatus struct {
	Status string `json:"status"`
}

func NewHealthHandler(ready ReadinessFunc, logger plumelog.StructuredLogger) *HealthHandler {
	return &HealthHandler{ready: ready, Logger: logger}
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, HealthStatus{Status: "ok"}, nil))
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.ready != nil {
		if err := h.ready(r.Context()); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeUnavailable).
				Code(contract.CodeUnavailable).
				Message("workerfleet readiness check failed").
				Build()))
			return
		}
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, HealthStatus{Status: "ready"}, nil))
}
