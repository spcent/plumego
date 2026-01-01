package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Response represents a standardized JSON response structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents a standardized error response structure
type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// JSONWriter provides consistent JSON response writing
type JSONWriter struct {
	resourceName string
}

// NewJSONWriter creates a new JSON writer with resource name
func NewJSONWriter(resourceName string) *JSONWriter {
	return &JSONWriter{resourceName: resourceName}
}

// WriteJSON writes a standardized JSON response
func (jw *JSONWriter) WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Code:    status,
		Message: http.StatusText(status),
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// WriteError writes a standardized error response
func (jw *JSONWriter) WriteError(w http.ResponseWriter, status int, message string, details interface{}) {
	resourceName := "resource"
	if jw != nil && jw.resourceName != "" {
		resourceName = jw.resourceName
	}

	// Log the error
	log.Printf("Error for resource %s: %s (status: %d, details: %v)", resourceName, message, status, details)

	// Return standardized error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResp := ErrorResponse{
		Code:    status,
		Message: message,
		Details: details,
	}

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
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

	// Batch operations
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
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusNoContent)
}

// Head handles HEAD /resource requests (returns headers only, no body)
func (c *BaseResourceController) Head(w http.ResponseWriter, r *http.Request) {
	// Default implementation: just return 200 OK with no body
	w.WriteHeader(http.StatusOK)
}

// BatchCreate handles batch creation of resources
func (c *BaseResourceController) BatchCreate(w http.ResponseWriter, r *http.Request) {
	c.notImplemented(w, r, "BatchCreate")
}

// BatchDelete handles batch deletion of resources
func (c *BaseResourceController) BatchDelete(w http.ResponseWriter, r *http.Request) {
	c.notImplemented(w, r, "BatchDelete")
}

// JSON is a helper method to write JSON responses
func (c *BaseResourceController) JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(fmt.Sprintf(`{"code": %d, "message": "%s", "data": %v}`, status, http.StatusText(status), data)))
}

// Error is a helper method to write standardized error responses
func (c *BaseResourceController) Error(w http.ResponseWriter, status int, message string, details any) {
	// Log the error
	log.Printf("Error for resource %s: %s (status: %d, details: %v)", c.ResourceName, message, status, details)

	// Return standardized error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(fmt.Sprintf(`{"code": %d, "message": "%s", "details": %v}`, status, message, details)))
}

// notImplemented is a helper method to return a standardized "Not Implemented" response
func (c *BaseResourceController) notImplemented(w http.ResponseWriter, _ *http.Request, method string) {
	c.Error(w, http.StatusNotImplemented, "Not Implemented",
		fmt.Sprintf("The %s method is not implemented for the %s resource", method, c.ResourceName))
}
