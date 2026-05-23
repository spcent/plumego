// Package handler contains the HTTP handlers for the reference application.
package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

const codeGreetNameRequired = "greet.name.required"

// APIHandler handles the core JSON API endpoints.
// Logger is optional: when non-nil it emits structured log entries on each request.
// Pass a.Core.Logger() from routes.go to demonstrate structured logging.
// ServiceName carries the service identity string; set it from config so every
// response reflects the actual service name rather than a hardcoded placeholder.
// Version carries the build-time version string injected via main.go ldflags.
type APIHandler struct {
	Logger      plumelog.StructuredLogger
	ServiceName string
	Version     string
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
		Service: h.ServiceName,
		Version: h.Version,
		Docs:    "/api/hello",
	}, nil)
}

// Hello responds with service metadata and the canonical endpoint list.
//
// The list is explicit rather than auto-generated so that each entry carries a
// meaningful name and description. app_test.go TestHelloEndpointListMatchesRegisteredRoutes
// validates that this list stays in sync with routes.go.
// When you add a route in routes.go, add a matching entry here.
//
// Ordered by method (DELETE < GET < POST < PUT) then path — matching
// the sort order returned by a.Core.Routes().
func (h APIHandler) Hello(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, helloResponse{
		Message:   "hello from plumego standard-service",
		Service:   h.ServiceName,
		Mode:      "canonical",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   h.Version,
		Features: []string{
			"stable_root_only",
			"explicit_routes",
			"stdlib_handlers",
			"minimal_bootstrap",
		},
		Endpoints: []endpointInfo{
			{Name: "items_delete", Method: http.MethodDelete, Path: "/api/v1/items/:id", Description: "delete an item"},
			{Name: "root", Method: http.MethodGet, Path: "/", Description: "service identity"},
			{Name: "api_hello", Method: http.MethodGet, Path: "/api/hello", Description: "service discovery — this endpoint"},
			{Name: "api_status", Method: http.MethodGet, Path: "/api/status", Description: "runtime status"},
			{Name: "api_greet", Method: http.MethodGet, Path: "/api/v1/greet", Description: "greeting (demonstrates TypeRequired error)"},
			{Name: "items_list", Method: http.MethodGet, Path: "/api/v1/items", Description: "list items with limit/offset pagination"},
			{Name: "items_get", Method: http.MethodGet, Path: "/api/v1/items/:id", Description: "get item by id"},
			{Name: "healthz", Method: http.MethodGet, Path: "/healthz", Description: "liveness probe"},
			{Name: "readyz", Method: http.MethodGet, Path: "/readyz", Description: "readiness probe"},
			{Name: "items_create", Method: http.MethodPost, Path: "/api/v1/items", Description: "create an item"},
			{Name: "items_update", Method: http.MethodPut, Path: "/api/v1/items/:id", Description: "update an item"},
		},
	}, nil)
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
			Code(codeGreetNameRequired).
			Detail("field", "name").
			Message("name is required").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, greetResponse{Message: "hello, " + name}, nil)
}

// Status responds with a summary of system health and component state.
// It demonstrates two structured-logging patterns:
//   - WithFields attaches fixed fields to every subsequent log call.
//   - contract.RequestIDFromContext extracts the correlation ID stamped by
//     the requestid middleware so log lines are linkable to the inbound request.
func (h APIHandler) Status(w http.ResponseWriter, r *http.Request) {
	if h.Logger != nil {
		requestID := contract.RequestIDFromContext(r.Context())
		h.Logger.WithFields(plumelog.Fields{
			"endpoint":   "status",
			"request_id": requestID,
		}).Info("status request handled")
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, statusResponse{
		Status:    "healthy",
		Service:   h.ServiceName,
		Version:   h.Version,
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
	}, nil)
}
