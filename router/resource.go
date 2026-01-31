package router

import (
	"fmt"
	"net/http"

	"github.com/spcent/plumego/contract"
)

// Deprecated: use contract.Response or contract.WriteResponse instead.
// Response represents a standardized JSON response structure.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Deprecated: use contract.APIError or contract.WriteError instead.
// ErrorResponse represents a standardized error response structure.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// JSONWriter provides consistent JSON response writing.
// Deprecated: prefer contract.WriteResponse/WriteError for new code.
type JSONWriter struct {
	resourceName string
}

// NewJSONWriter creates a new JSON writer with resource name
func NewJSONWriter(resourceName string) *JSONWriter {
	return &JSONWriter{resourceName: resourceName}
}

// WriteJSON writes a standardized JSON response
func (jw *JSONWriter) WriteJSON(w http.ResponseWriter, status int, data any) {
	if err := contract.WriteResponse(w, nil, status, data, nil); err != nil {
		if jw != nil && jw.resourceName != "" {
			fmt.Printf("[router/resource] JSON encoding failed for %s: %v\n", jw.resourceName, err)
		} else {
			fmt.Printf("[router/resource] JSON encoding failed: %v\n", err)
		}
	}
}

// WriteError writes a standardized error response
func (jw *JSONWriter) WriteError(w http.ResponseWriter, status int, message string, details any) {
	apiErr := contract.APIError{
		Status:   status,
		Code:     http.StatusText(status),
		Message:  message,
		Category: contract.CategoryForStatus(status),
	}
	if details != nil {
		apiErr.Details = map[string]any{"details": details}
	}
	contract.WriteError(w, nil, apiErr)
}

// WriteResponse writes a contract response payload and includes trace id when available.
func (jw *JSONWriter) WriteResponse(w http.ResponseWriter, r *http.Request, status int, data any, meta map[string]any) error {
	return contract.WriteResponse(w, r, status, data, meta)
}

// WriteAPIError writes a contract error payload.
func (jw *JSONWriter) WriteAPIError(w http.ResponseWriter, r *http.Request, err contract.APIError) {
	contract.WriteError(w, r, err)
}

// WriteNotImplemented writes a standardized "Not Implemented" response
func (jw *JSONWriter) WriteNotImplemented(w http.ResponseWriter, method string) {
	resourceName := "resource"
	if jw != nil && jw.resourceName != "" {
		resourceName = jw.resourceName
	}

	jw.WriteError(w, http.StatusNotImplemented, "Not Implemented",
		fmt.Sprintf("The %s method is not implemented for the %s resource", method, resourceName))
}

// ResourceController defines the interface for RESTful resource controllers
type ResourceController interface {
	// Basic CRUD operations
	Index(http.ResponseWriter, *http.Request)  // GET /resource
	Show(http.ResponseWriter, *http.Request)   // GET /resource/:id
	Create(http.ResponseWriter, *http.Request) // POST /resource
	Update(http.ResponseWriter, *http.Request) // PUT /resource/:id
	Delete(http.ResponseWriter, *http.Request) // DELETE /resource/:id
	Patch(http.ResponseWriter, *http.Request)  // PATCH /resource/:id

	// Additional HTTP methods
	Options(http.ResponseWriter, *http.Request) // OPTIONS /resource
	Head(http.ResponseWriter, *http.Request)    // HEAD /resource

	// Batch operations (optional)
	BatchCreate(http.ResponseWriter, *http.Request) // POST /resource/batch
	BatchDelete(http.ResponseWriter, *http.Request) // DELETE /resource/batch
}

// BaseResourceController provides a default implementation of ResourceController
// All methods return "Not Implemented" error by default
type BaseResourceController struct {
	ResourceName string      // Resource name used for response messages
	jsonWriter   *JSONWriter // JSON writer for consistent responses
}

// NewBaseResourceController creates a new base resource controller with JSON writer
func NewBaseResourceController(resourceName string) *BaseResourceController {
	return &BaseResourceController{
		ResourceName: resourceName,
		jsonWriter:   NewJSONWriter(resourceName),
	}
}

// Index handles GET /resource requests
func (c *BaseResourceController) Index(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Index")
}

// Show handles GET /resource/:id requests
func (c *BaseResourceController) Show(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Show")
}

// Create handles POST /resource requests
func (c *BaseResourceController) Create(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Create")
}

// Update handles PUT /resource/:id requests
func (c *BaseResourceController) Update(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Update")
}

// Delete handles DELETE /resource/:id requests
func (c *BaseResourceController) Delete(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Delete")
}

// Patch handles PATCH /resource/:id requests
func (c *BaseResourceController) Patch(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Patch")
}

// Options handles OPTIONS /resource requests for CORS and method negotiation
func (c *BaseResourceController) Options(w http.ResponseWriter, r *http.Request) {
	// Set default CORS headers
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
	w.WriteHeader(http.StatusNoContent)
}

// Head handles HEAD /resource requests (returns headers only, no body)
func (c *BaseResourceController) Head(w http.ResponseWriter, r *http.Request) {
	// Default implementation: just return 200 OK with no body
	w.WriteHeader(http.StatusOK)
}

// BatchCreate handles batch creation of resources
func (c *BaseResourceController) BatchCreate(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "BatchCreate")
}

// BatchDelete handles batch deletion of resources
func (c *BaseResourceController) BatchDelete(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "BatchDelete")
}

// JSON is a helper method to write JSON responses
func (c *BaseResourceController) JSON(w http.ResponseWriter, status int, data any) {
	c.jsonWriter.WriteJSON(w, status, data)
}

// Error is a helper method to write standardized error responses
func (c *BaseResourceController) Error(w http.ResponseWriter, status int, message string, details any) {
	c.jsonWriter.WriteError(w, status, message, details)
}

// Response is a helper method to write contract-based JSON responses.
func (c *BaseResourceController) Response(w http.ResponseWriter, r *http.Request, status int, data any, meta map[string]any) {
	_ = c.jsonWriter.WriteResponse(w, r, status, data, meta)
}

// APIError is a helper method to write contract-based error responses.
func (c *BaseResourceController) APIError(w http.ResponseWriter, r *http.Request, err contract.APIError) {
	c.jsonWriter.WriteAPIError(w, r, err)
}
