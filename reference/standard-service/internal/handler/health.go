package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// ReadinessChecker reports whether a dependency is ready to serve traffic.
// Register one per dependency (database, cache, downstream service) in routes.go.
type ReadinessChecker interface {
	Ready(ctx context.Context) error
}

// HealthHandler serves the canonical liveness and readiness endpoints.
type HealthHandler struct {
	ServiceName string
	// Checkers is an optional list of dependency readiness probes.
	// When nil or empty the readiness endpoint reports ready immediately.
	Checkers []ReadinessChecker
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Check     string `json:"check"`
	Timestamp string `json:"timestamp"`
}

// Live reports that the process is serving HTTP traffic.
func (h HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   h.ServiceName,
		Check:     "liveness",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil)
}

// Ready probes each registered ReadinessChecker in order.
// The first failure returns 503 TypeUnavailable; all passing returns 200.
func (h HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	for _, checker := range h.Checkers {
		if err := checker.Ready(r.Context()); err != nil {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeUnavailable).
				Detail("reason", err.Error()).
				Message("service not ready").
				Build())
			return
		}
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ready",
		Service:   h.ServiceName,
		Check:     "readiness",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil)
}
