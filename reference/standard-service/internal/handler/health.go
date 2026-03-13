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

// Live reports that the process is serving HTTP traffic.
func (h HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	service := h.ServiceName
	if service == "" {
		service = "plumego-reference"
	}

	if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"status":    "ok",
		"service":   service,
		"check":     "liveness",
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// Ready reports that the reference service is ready to accept requests.
func (h HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	service := h.ServiceName
	if service == "" {
		service = "plumego-reference"
	}

	if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"status":    "ready",
		"service":   service,
		"check":     "readiness",
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
