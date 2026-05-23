// Package handler contains the HTTP handlers for the production reference application.
// Each handler is a struct with only the dependencies it needs, wired via constructor
// injection in internal/app/routes.go. This keeps handlers independently testable
// and avoids coupling every handler to the full App struct.
package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
)

const codeComponentUnhealthy = "health.component.unhealthy"

// ServiceHandler serves the root, liveness, and readiness endpoints.
// These endpoints need only the service name and a static features list.
// Checkers is an optional list of dependency readiness probes for /readyz.
// Each checker's Name() labels its result in the readiness response.
// When nil or empty the readiness endpoint reports ready immediately.
type ServiceHandler struct {
	ServiceName string
	Features    []string
	Checkers    []health.ComponentChecker
}

type serviceResponse struct {
	Service   string   `json:"service"`
	Mode      string   `json:"mode"`
	Timestamp string   `json:"timestamp"`
	Features  []string `json:"features"`
}

type livenessResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Check     string `json:"check"`
	Timestamp string `json:"timestamp"`
}

// Root responds with service identity and enabled features.
func (h ServiceHandler) Root(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, serviceResponse{
		Service:   h.ServiceName,
		Mode:      "production-reference",
		Timestamp: utcNow(),
		Features:  h.Features,
	}, nil)
}

// Live reports that the process is serving HTTP traffic.
// Kubernetes liveness probes call this endpoint; it must never be gated on
// dependency health — if a dependency is down the process is still alive.
func (h ServiceHandler) Live(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, livenessResponse{
		Status:    "ok",
		Service:   h.ServiceName,
		Check:     "liveness",
		Timestamp: utcNow(),
	}, nil)
}

// Ready probes each registered health.ComponentChecker in order.
// On success it returns health.ReadinessStatus with a per-component map.
// On the first failure it returns 503 TypeUnavailable naming the failing
// component so operators can triage without log access.
//
//	GET /readyz (all pass)  → 200 health.ReadinessStatus{Ready:true, Components:{…}}
//	GET /readyz (db fails)  → 503 TypeUnavailable detail.component="database"
func (h ServiceHandler) Ready(w http.ResponseWriter, r *http.Request) {
	components := make(map[string]bool, len(h.Checkers))
	for _, checker := range h.Checkers {
		if err := checker.Check(r.Context()); err != nil {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeUnavailable).
				Code(codeComponentUnhealthy).
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

func utcNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}
