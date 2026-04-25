package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// HealthHandler serves the canonical liveness and readiness endpoints.
type HealthHandler struct {
	ServiceName string
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Check     string `json:"check"`
	Timestamp string `json:"timestamp"`
}

// Live reports that the process is serving HTTP traffic.
func (h HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	service := h.ServiceName
	if service == "" {
		service = "plumego-reference"
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   service,
		Check:     "liveness",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil)
}

// Ready reports that the reference service is ready to accept requests.
func (h HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	service := h.ServiceName
	if service == "" {
		service = "plumego-reference"
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, healthResponse{
		Status:    "ready",
		Service:   service,
		Check:     "readiness",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil)
}
