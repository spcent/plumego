package health

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

const (
	readinessHandlerTimeout = 5 * time.Second
)

// ReadinessHandler exposes the current readiness state as JSON.
// It returns HTTP 200 when ready and 503 otherwise.
func ReadinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		status := GetReadiness()
		code := http.StatusOK
		if !status.Ready {
			code = http.StatusServiceUnavailable
		}

		_ = contract.WriteJSON(w, code, status)
	})
}

// ReadinessHandlerWithManager exposes the current readiness state based on component health.
// It returns HTTP 200 when ready and 503 otherwise.
func ReadinessHandlerWithManager(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if manager == nil {
			sendErrorResponse(w, r, http.StatusServiceUnavailable, "HEALTH_MANAGER_UNAVAILABLE",
				"Health manager is not configured", "")
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), readinessHandlerTimeout)
		defer cancel()

		// Perform health check
		overallHealth := manager.CheckAllComponents(ctx)

		code := http.StatusOK
		if !overallHealth.Status.isReady() {
			code = http.StatusServiceUnavailable
		}

		response := map[string]any{
			"ready":     overallHealth.Status.isReady(),
			"status":    overallHealth.Status,
			"message":   overallHealth.Message,
			"timestamp": overallHealth.Timestamp,
		}

		_ = contract.WriteJSON(w, code, response)
	})
}

// BuildInfoHandler exposes build metadata as JSON for diagnostics and release verification.
func BuildInfoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = contract.WriteJSON(w, http.StatusOK, GetBuildInfo())
	})
}
