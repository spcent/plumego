// Package handler contains the HTTP handlers for the production reference application.
// Each handler is a struct with only the dependencies it needs, wired via constructor
// injection in internal/app/routes.go. This keeps handlers independently testable
// and avoids coupling every handler to the full App struct.
package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// ServiceHandler serves the root, liveness, and readiness endpoints.
// These endpoints carry no dependency on external state; they need only
// the service name and a static features list.
type ServiceHandler struct {
	ServiceName string
	Features    []string
}

type serviceResponse struct {
	Service   string   `json:"service"`
	Mode      string   `json:"mode"`
	Timestamp string   `json:"timestamp"`
	Features  []string `json:"features"`
}

type healthResponse struct {
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
	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   h.ServiceName,
		Check:     "liveness",
		Timestamp: utcNow(),
	}, nil)
}

// Ready reports that the service is ready to handle traffic.
func (h ServiceHandler) Ready(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ready",
		Service:   h.ServiceName,
		Check:     "readiness",
		Timestamp: utcNow(),
	}, nil)
}

func utcNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}
