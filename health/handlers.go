package health

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

const (
	healthHandlerTimeout       = 10 * time.Second
	componentHealthTimeout     = 5 * time.Second
	allComponentsHealthTimeout = 15 * time.Second
)

// HealthResponse is the JSON body for the main health endpoint.
type HealthResponse struct {
	HealthStatus
	BuildInfo BuildInfo    `json:"build_info,omitempty"`
	Runtime   *RuntimeInfo `json:"runtime,omitempty"`
}

// ComponentsListResponse is the JSON body for the component list endpoint.
type ComponentsListResponse struct {
	Components []string `json:"components"`
	Count      int      `json:"count"`
}

// SummaryHandler returns aggregate component health only.
func SummaryHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if manager == nil {
			contract.WriteError(w, r, contract.APIError{
				Status:   http.StatusServiceUnavailable,
				Code:     "HEALTH_MANAGER_UNAVAILABLE",
				Message:  "health manager is not configured",
				Category: contract.CategoryForStatus(http.StatusServiceUnavailable),
			})
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), healthHandlerTimeout)
		defer cancel()

		health := manager.CheckAllComponents(ctx)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_ = contract.WriteJSON(w, httpStatusForHealth(health.Status), health)
	})
}

// DetailedHandler returns aggregate health plus build metadata.
func DetailedHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), healthHandlerTimeout)
		defer cancel()

		health := manager.CheckAllComponents(ctx)
		resp := HealthResponse{
			HealthStatus: health,
			BuildInfo:    GetBuildInfo(),
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_ = contract.WriteJSON(w, httpStatusForHealth(health.Status), resp)
	})
}

// HealthHandler creates a comprehensive health check handler.
// Pass debug=true to include runtime diagnostics in the response.
func HealthHandler(manager HealthManager, debug bool) http.Handler {
	if !debug {
		return DetailedHandler(manager)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), healthHandlerTimeout)
		defer cancel()

		health := manager.CheckAllComponents(ctx)
		resp := HealthResponse{
			HealthStatus: health,
			BuildInfo:    GetBuildInfo(),
			Runtime:      getRuntimeInfo(),
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_ = contract.WriteJSON(w, httpStatusForHealth(health.Status), resp)
	})
}

// ComponentHealthHandler creates a handler for checking a specific component's health.
func ComponentHealthHandler(manager HealthManager, componentName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), componentHealthTimeout)
		defer cancel()

		_ = manager.CheckComponent(ctx, componentName)

		health, exists := manager.GetComponentHealth(componentName)
		if !exists {
			contract.WriteError(w, r, contract.APIError{
				Status:   http.StatusNotFound,
				Code:     "COMPONENT_NOT_FOUND",
				Message:  "component not found",
				Category: contract.CategoryForStatus(http.StatusNotFound),
			})
			return
		}

		_ = contract.WriteJSON(w, httpStatusForHealth(health.Status), health)
	})
}

// AllComponentsHealthHandler creates a handler for checking all components' health.
func AllComponentsHealthHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), allComponentsHealthTimeout)
		defer cancel()

		manager.CheckAllComponents(ctx)
		_ = contract.WriteJSON(w, http.StatusOK, manager.GetAllHealth())
	})
}

// LiveHandler creates a liveness probe handler that always returns 200.
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
		names := make([]string, 0, len(allHealth))
		for name := range allHealth {
			names = append(names, name)
		}

		_ = contract.WriteJSON(w, http.StatusOK, ComponentsListResponse{
			Components: names,
			Count:      len(names),
		})
	})
}
