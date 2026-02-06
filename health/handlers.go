package health

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/spcent/plumego/contract"
)

const (
	healthHandlerTimeout       = 10 * time.Second
	componentHealthTimeout     = 5 * time.Second
	allComponentsHealthTimeout = 15 * time.Second
)

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      string    `json:"code,omitempty"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

// HealthResponse represents a standardized health response.
type HealthResponse struct {
	HealthStatus `json:",inline"`
	ResponseTime time.Duration   `json:"response_time"`
	RequestID    string          `json:"request_id,omitempty"`
	BuildInfo    BuildInfo       `json:"build_info,omitempty"`
	Readiness    ReadinessStatus `json:"readiness,omitempty"`
}

// HealthHandler creates a comprehensive health check handler with enhanced error handling.
func HealthHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		requestID := extractRequestID(r)

		// Enhanced error handling with recovery
		defer func() {
			if rvr := recover(); rvr != nil {
				handlePanic(w, r, rvr, requestID)
			}
		}()

		// Validate request method
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			sendErrorResponse(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
				"Only GET and HEAD methods are allowed", requestID)
			return
		}

		if manager == nil {
			sendErrorResponse(w, r, http.StatusServiceUnavailable, "HEALTH_MANAGER_UNAVAILABLE",
				"Health manager is not configured", requestID)
			return
		}

		// Perform health check with timeout
		ctx, cancel := withCheckTimeout(r.Context(), healthHandlerTimeout)
		defer cancel()

		var health HealthStatus

		// Use goroutine with channel for better timeout handling
		done := make(chan bool, 1)
		go func() {
			defer func() { done <- true }()
			health = manager.CheckAllComponents(ctx)
		}()

		// Wait for completion or timeout
		select {
		case <-done:
			// Health check completed successfully
		case <-ctx.Done():
			// Health check timed out
			sendErrorResponse(w, r, http.StatusGatewayTimeout, "HEALTH_CHECK_TIMEOUT",
				"Health check timed out", requestID)
			return
		}

		code := httpStatusForHealth(health.Status)

		// Create enhanced response
		response := HealthResponse{
			HealthStatus: health,
			ResponseTime: time.Since(startTime),
			RequestID:    requestID,
			BuildInfo:    GetBuildInfo(),
			Readiness:    GetReadiness(),
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		if err := contract.WriteJSON(w, code, response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})
}

// ComponentHealthHandler creates a handler for checking specific component health.
func ComponentHealthHandler(manager HealthManager, componentName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), componentHealthTimeout)
		defer cancel()

		// Check the specific component
		_ = manager.CheckComponent(ctx, componentName)

		health, exists := manager.GetComponentHealth(componentName)
		if !exists {
			_ = contract.WriteJSON(w, http.StatusNotFound, map[string]string{
				"error": "component not found",
			})
			return
		}

		_ = contract.WriteJSON(w, httpStatusForHealth(health.Status), health)
	})
}

// AllComponentsHealthHandler creates a handler for checking all components health.
func AllComponentsHealthHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), allComponentsHealthTimeout)
		defer cancel()

		// Check all components
		manager.CheckAllComponents(ctx)
		allHealth := manager.GetAllHealth()

		_ = contract.WriteJSON(w, http.StatusOK, allHealth)
	})
}

// LiveHandler creates a liveness probe handler that always returns 200.
// This is useful for Kubernetes liveness probes.
func LiveHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	})
}

// ComponentsListHandler creates a handler that lists all registered components.
func ComponentsListHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		allHealth := manager.GetAllHealth()
		components := make([]string, 0, len(allHealth))
		for name := range allHealth {
			components = append(components, name)
		}

		response := map[string]any{
			"components": components,
			"count":      len(components),
		}

		_ = contract.WriteJSON(w, http.StatusOK, response)
	})
}

// extractRequestID extracts request ID from headers for tracing.
func extractRequestID(r *http.Request) string {
	// Try common headers for request ID
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	if id := r.Header.Get("X-Correlation-ID"); id != "" {
		return id
	}
	if id := r.Header.Get("Request-ID"); id != "" {
		return id
	}
	return ""
}

// sendErrorResponse sends a standardized error response.
func sendErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, code, message, requestID string) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	details := map[string]any{
		"error":     http.StatusText(statusCode),
		"timestamp": time.Now(),
	}
	if requestID != "" {
		details["request_id"] = requestID
	}

	apiErr := contract.APIError{
		Status:   statusCode,
		Code:     code,
		Message:  message,
		Category: contract.CategoryForStatus(statusCode),
		Details:  details,
	}
	if requestID != "" {
		apiErr.TraceID = requestID
	}
	contract.WriteError(w, r, apiErr)
}

// handlePanic handles panics in HTTP handlers gracefully.
func handlePanic(w http.ResponseWriter, r *http.Request, panicValue any, requestID string) {
	log.Printf("[PANIC] health handler panic: %v\n%s", panicValue, debug.Stack())

	sendErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR",
		"Internal server error occurred", requestID)
}

// isDevelopment checks if the application is running in development mode.
func isDevelopment() bool {
	return os.Getenv("APP_ENV") == "development"
}
