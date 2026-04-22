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
	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}

// Greet demonstrates the canonical error response pattern.
// It requires the "name" query parameter and returns a structured
// TypeBadRequest error if it is absent.
//
// Example:
//
//	GET /api/v1/greet?name=Alice  → 200 {"message":"hello, Alice"}
//	GET /api/v1/greet             → 400 structured APIError
func (h APIHandler) Greet(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Detail("field", "name").
			Message("name is required").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"message": "hello, " + name,
	}, nil)
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
	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}
