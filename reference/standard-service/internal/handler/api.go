// Package handler contains the HTTP handlers for the reference application.
package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// APIHandler handles the core JSON API endpoints.
type APIHandler struct{}

// Hello responds with service metadata and available endpoints.
func (h APIHandler) Hello(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"message":   "hello from plumego standard-service",
		"service":   "plumego-reference",
		"mode":      "canonical",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"features": []string{
			"stable_root_only",
			"explicit_routes",
			"stdlib_handlers",
			"minimal_bootstrap",
		},
		"endpoints": map[string]string{
			"root":       "/",
			"healthz":    "/healthz",
			"readyz":     "/readyz",
			"api_hello":  "/api/hello",
			"api_status": "/api/status",
		},
	}
	if err := contract.WriteResponse(w, r, http.StatusOK, resp, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// Status responds with a summary of system health and component state.
func (h APIHandler) Status(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"status":    "healthy",
		"service":   "plumego-reference",
		"version":   "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
		"structure": map[string]any{
			"bootstrap":  "explicit",
			"extensions": "excluded_from_canonical_path",
			"handlers":   "net/http",
			"routes":     "one_method_one_path_one_handler",
		},
		"modules": []string{
			"core",
			"router",
			"contract",
			"middleware",
		},
	}
	if err := contract.WriteResponse(w, r, http.StatusOK, resp, nil); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
