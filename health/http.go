package health

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Error      string    `json:"error"`
	Code       string    `json:"code,omitempty"`
	Message    string    `json:"message,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	RequestID  string    `json:"request_id,omitempty"`
	StackTrace string    `json:"stack_trace,omitempty"`
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
			sendErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
				"Only GET and HEAD methods are allowed", requestID)
			return
		}

		// Perform health check with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
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
			sendErrorResponse(w, http.StatusGatewayTimeout, "HEALTH_CHECK_TIMEOUT",
				"Health check timed out", requestID)
			return
		}

		// Determine response code based on health status
		code := http.StatusOK
		switch health.Status {
		case StatusUnhealthy:
			code = http.StatusServiceUnavailable
		case StatusDegraded:
			code = http.StatusPartialContent
		}

		// Create enhanced response
		response := HealthResponse{
			HealthStatus: health,
			ResponseTime: time.Since(startTime),
			RequestID:    requestID,
			BuildInfo:    GetBuildInfo(),
			Readiness:    GetReadiness(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(code)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			// Log encoding error but don't fail the request
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})
}

// ComponentHealthHandler creates a handler for checking specific component health.
func ComponentHealthHandler(manager HealthManager, componentName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Check the specific component
		_ = manager.CheckComponent(ctx, componentName)

		health, exists := manager.GetComponentHealth(componentName)
		if !exists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "component not found",
			})
			return
		}

		code := http.StatusOK
		switch health.Status {
		case StatusUnhealthy:
			code = http.StatusServiceUnavailable
		case StatusDegraded:
			code = http.StatusPartialContent
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(health)
	})
}

// AllComponentsHealthHandler creates a handler for checking all components health.
func AllComponentsHealthHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Check all components
		manager.CheckAllComponents(ctx)
		allHealth := manager.GetAllHealth()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(allHealth)
	})
}

// HealthHistoryHandler creates a handler that returns health check history.
func HealthHistoryHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		history := manager.GetHealthHistory()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(history)
	})
}

// HealthHistoryExportHandler creates a handler that exports health check history in various formats.
func HealthHistoryExportHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters for filtering and format
		query := HealthHistoryQuery{}

		// Parse time range
		if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
			if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
				query.StartTime = &startTime
			}
		}

		if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
			if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
				query.EndTime = &endTime
			}
		}

		// Parse state filter
		if stateStr := r.URL.Query().Get("state"); stateStr != "" {
			state := HealthState(stateStr)
			query.State = &state
		}

		// Parse component filter
		if component := r.URL.Query().Get("component"); component != "" {
			query.Component = component
		}

		// Parse limit
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				query.Limit = limit
			}
		}

		// Parse offset
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil {
				query.Offset = offset
			}
		}

		// Get format parameter
		format := r.URL.Query().Get("format")
		if format == "" {
			format = "json"
		}

		// Query history
		result := manager.QueryHealthHistory(query)

		switch strings.ToLower(format) {
		case "csv":
			exportHistoryToCSV(w, result.Entries)
		case "json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(result)
		default:
			sendErrorResponse(w, http.StatusBadRequest, "INVALID_FORMAT",
				"Supported formats: json, csv", "")
			return
		}
	})
}

// exportHistoryToCSV exports health history entries to CSV format.
func exportHistoryToCSV(w http.ResponseWriter, entries []HealthHistoryEntry) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=health_history.csv")

	writer := csv.NewWriter(w)

	// Write header
	header := []string{"Timestamp", "State", "Message", "Components", "Duration"}
	_ = writer.Write(header)

	// Write data
	for _, entry := range entries {
		record := []string{
			entry.Timestamp.Format(time.RFC3339),
			string(entry.State),
			entry.Message,
			strings.Join(entry.Components, ";"),
			entry.Duration.String(),
		}
		_ = writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		http.Error(w, "Failed to write CSV", http.StatusInternalServerError)
	}
}

// HealthHistoryStatsHandler returns statistics about health history.
func HealthHistoryStatsHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats := manager.GetHealthHistoryStats()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(stats)
	})
}

// ReadinessHandler exposes the current readiness state as JSON.
// It returns HTTP 200 when ready and 503 otherwise.
func ReadinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		status := GetReadiness()
		code := http.StatusOK
		if !status.Ready {
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(status)
	})
}

// ReadinessHandlerWithManager exposes the current readiness state based on component health.
// It returns HTTP 200 when ready and 503 otherwise.
func ReadinessHandlerWithManager(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Perform health check
		overallHealth := manager.CheckAllComponents(ctx)

		code := http.StatusOK
		if !overallHealth.Status.isReady() {
			code = http.StatusServiceUnavailable
		}

		response := map[string]interface{}{
			"ready":     overallHealth.Status.isReady(),
			"status":    overallHealth.Status,
			"message":   overallHealth.Message,
			"timestamp": overallHealth.Timestamp,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(response)
	})
}

// BuildInfoHandler exposes build metadata as JSON for diagnostics and release verification.
func BuildInfoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GetBuildInfo())
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
		allHealth := manager.GetAllHealth()
		components := make([]string, 0, len(allHealth))
		for name := range allHealth {
			components = append(components, name)
		}

		response := map[string]interface{}{
			"components": components,
			"count":      len(components),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	})
}

// isReady checks if the health status indicates the service is ready to serve traffic.
func (hs HealthState) isReady() bool {
	return hs == StatusHealthy || hs == StatusDegraded
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
func sendErrorResponse(w http.ResponseWriter, statusCode int, code, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Error:     http.StatusText(statusCode),
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		RequestID: requestID,
	}

	if isDevelopment() {
		errorResp.StackTrace = string(debug.Stack())
	}

	_ = json.NewEncoder(w).Encode(errorResp)
}

// handlePanic handles panics in HTTP handlers gracefully.
func handlePanic(w http.ResponseWriter, _ *http.Request, panicValue interface{}, requestID string) {
	// Log the panic (in a real application, you'd use proper logging)
	fmt.Printf("Panic in health handler: %v\nStack: %s\n", panicValue, debug.Stack())

	sendErrorResponse(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR",
		"Internal server error occurred", requestID)
}

// isDevelopment checks if the application is running in development mode.
func isDevelopment() bool {
	// Simple check - in real applications, this might check environment variables
	return strings.Contains(strings.ToLower(GetBuildInfo().Version), "dev")
}
