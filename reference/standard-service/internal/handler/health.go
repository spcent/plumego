package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
)

// HealthHandler serves the canonical liveness and readiness endpoints.
//
// Liveness (/healthz) always returns 200 as long as the process is alive.
// It uses a simple local response struct — no dependency probing needed.
//
// Readiness (/readyz) probes each health.ComponentChecker in order and
// surfaces component names in both success and error responses so operators
// immediately know which dependency is unhealthy without inspecting logs.
//
// Use health.ComponentChecker (from github.com/spcent/plumego/health) as the
// checker type. It carries a stable Name() for labeling plus Check() for the
// probe. Implement health.ComponentChecker directly or adapt an existing
// client's health method with a small wrapper struct in routes.go.
type HealthHandler struct {
	ServiceName string
	// Checkers is an optional list of dependency readiness probes.
	// Each checker's Name() labels its result in the readiness response.
	// When nil or empty the readiness endpoint reports ready immediately.
	Checkers []health.ComponentChecker
}

// livenessResponse is the body returned by the liveness probe.
// Liveness only confirms the process is up; it carries no component state.
type livenessResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// Live reports that the process is serving HTTP traffic.
// Kubernetes liveness probes call this endpoint; it must never be gated on
// dependency health — if a dependency is down the process is still alive.
func (h HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, livenessResponse{
		Status:    "ok",
		Service:   h.ServiceName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil)
}

// Ready probes each registered health.ComponentChecker in order.
//
// On success the response body is a health.ReadinessStatus with a per-component
// map so load balancers and dashboards can see which dependencies passed.
// On the first failure a 503 TypeUnavailable error includes the component
// name and the probe error, making triage actionable without log access.
//
//	GET /readyz (all pass) → 200 health.ReadinessStatus{Ready:true, Components:{…}}
//	GET /readyz (db fails) → 503 TypeUnavailable detail.component="database"
func (h HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	components := make(map[string]bool, len(h.Checkers))
	for _, checker := range h.Checkers {
		if err := checker.Check(r.Context()); err != nil {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeUnavailable).
				Code("health.component.unhealthy").
				Detail("component", checker.Name()).
				Detail("reason", err.Error()).
				Message("service not ready").
				Build())
			return
		}
		components[checker.Name()] = true
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, health.ReadinessStatus{
		Ready:      true,
		Timestamp:  time.Now().UTC(),
		Components: components,
	}, nil)
}
