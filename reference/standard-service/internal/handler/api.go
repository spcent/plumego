// Package handler contains the HTTP handlers for the reference application.
package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// APIHandler handles the core JSON API endpoints.
type APIHandler struct{}

type helloResponse struct {
	Message   string            `json:"message"`
	Service   string            `json:"service"`
	Mode      string            `json:"mode"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Features  []string          `json:"features"`
	Endpoints map[string]string `json:"endpoints"`
}

type greetResponse struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status    string          `json:"status"`
	Service   string          `json:"service"`
	Version   string          `json:"version"`
	Timestamp string          `json:"timestamp"`
	Structure statusStructure `json:"structure"`
	Modules   []string        `json:"modules"`
}

type statusStructure struct {
	Bootstrap  string `json:"bootstrap"`
	Extensions string `json:"extensions"`
	Handlers   string `json:"handlers"`
	Routes     string `json:"routes"`
}

// Hello responds with service metadata and available endpoints.
func (h APIHandler) Hello(w http.ResponseWriter, r *http.Request) {
	resp := helloResponse{
		Message:   "hello from plumego standard-service",
		Service:   "plumego-reference",
		Mode:      "canonical",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Features: []string{
			"stable_root_only",
			"explicit_routes",
			"stdlib_handlers",
			"minimal_bootstrap",
		},
		Endpoints: map[string]string{
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
	_ = contract.WriteResponse(w, r, http.StatusOK, greetResponse{Message: "hello, " + name}, nil)
}

// Status responds with a summary of system health and component state.
func (h APIHandler) Status(w http.ResponseWriter, r *http.Request) {
	resp := statusResponse{
		Status:    "healthy",
		Service:   "plumego-reference",
		Version:   "1.0.0",
		Timestamp: time.Now().Format(time.RFC3339),
		Structure: statusStructure{
			Bootstrap:  "explicit",
			Extensions: "excluded_from_canonical_path",
			Handlers:   "net/http",
			Routes:     "one_method_one_path_one_handler",
		},
		Modules: []string{
			"core",
			"router",
			"contract",
			"middleware",
		},
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}
