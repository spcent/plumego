// Package handler contains the HTTP handlers for the reference application.
package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// version is set at build time via -ldflags "-X standard-service/internal/handler.version=1.0.0"
// During local development, it defaults to "dev" to signal an unversioned build.
var version = "dev"

// APIHandler handles the core JSON API endpoints.
type APIHandler struct {
	Logger plumelog.StructuredLogger // Demonstrates structured logging capability.
	Routes []RouteInfo                // Populated after all routes are registered via core.Routes()
}

// RouteInfo mirrors router.RouteInfo for response documentation.
type RouteInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type rootResponse struct {
	Service string `json:"service"`
	Version string `json:"version"`
	Docs    string `json:"docs"`
}

// endpointInfo describes a single HTTP endpoint in the service discovery response.
type endpointInfo struct {
	Name        string `json:"name"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type helloResponse struct {
	Message   string         `json:"message"`
	Service   string         `json:"service"`
	Mode      string         `json:"mode"`
	Timestamp string         `json:"timestamp"`
	Version   string         `json:"version"`
	Features  []string       `json:"features"`
	Endpoints []endpointInfo `json:"endpoints"`
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

// Root responds with minimal service identity. For the full endpoint listing use GET /api/hello.
func (h APIHandler) Root(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, rootResponse{
		Service: "plumego-reference",
		Version: version,
		Docs:    "/api/hello",
	}, nil)
}

// Hello responds with service metadata and available endpoints in a deterministic,
// method-aware listing. Endpoints are generated from core.Routes() at initialization,
// so the list automatically includes all registered routes without manual maintenance.
func (h APIHandler) Hello(w http.ResponseWriter, r *http.Request) {
	// Build endpoint list from registered routes.
	endpoints := make([]endpointInfo, 0, len(h.Routes))
	for _, ri := range h.Routes {
		endpoints = append(endpoints, endpointInfo{
			Method: ri.Method,
			Path:   ri.Path,
			// Description is intentionally minimal since it's auto-generated.
			Description: "endpoint",
		})
	}

	resp := helloResponse{
		Message:   "hello from plumego standard-service",
		Service:   "plumego-reference",
		Mode:      "canonical",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   version,
		Features: []string{
			"stable_root_only",
			"explicit_routes",
			"stdlib_handlers",
			"minimal_bootstrap",
		},
		Endpoints: endpoints,
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}

// Greet demonstrates the canonical error response pattern.
// It requires the "name" query parameter and returns a structured
// TypeRequired error if it is absent.
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
			Code("greet.name.required").
			Detail("field", "name").
			Message("name is required").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, greetResponse{Message: "hello, " + name}, nil)
}

// Status responds with a summary of system health and component state.
func (h APIHandler) Status(w http.ResponseWriter, r *http.Request) {
	// Demonstrates structured logging: logger.WithFields() adds fields to subsequent entries.
	// In production, this helps correlate log lines by request ID, tenant, or feature.
	if h.Logger != nil {
		h.Logger.WithFields(plumelog.Fields{
			"endpoint":    "status",
			"route_count": len(h.Routes),
		}).Info("status request handled")
	}

	resp := statusResponse{
		Status:    "healthy",
		Service:   "plumego-reference",
		Version:   version,
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
			"log",
		},
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, resp, nil)
}
