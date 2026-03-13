package healthhttp

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
)

// HealthResponse is the JSON body for the main health endpoint.
type HealthResponse struct {
	health.HealthStatus
	BuildInfo health.BuildInfo `json:"build_info,omitempty"`
	Runtime   *RuntimeInfo     `json:"runtime,omitempty"`
}

// ComponentsListResponse is the JSON body for the component list endpoint.
type ComponentsListResponse struct {
	Components []string `json:"components"`
	Count      int      `json:"count"`
}

// SummaryHandler returns aggregate component health only.
func SummaryHandler(manager health.HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), healthHandlerTimeout)
		defer cancel()

		status := manager.CheckAllComponents(ctx)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_ = contract.WriteJSON(w, httpStatusForHealth(status.Status), status)
	})
}

// DetailedHandler returns aggregate health plus build metadata.
func DetailedHandler(manager health.HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), healthHandlerTimeout)
		defer cancel()

		status := manager.CheckAllComponents(ctx)
		resp := HealthResponse{
			HealthStatus: status,
			BuildInfo:    health.GetBuildInfo(),
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_ = contract.WriteJSON(w, httpStatusForHealth(status.Status), resp)
	})
}

// HealthHandler creates a comprehensive health check handler.
// Pass debug=true to include runtime diagnostics in the response.
func HealthHandler(manager health.HealthManager, debug bool) http.Handler {
	if !debug {
		return DetailedHandler(manager)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), healthHandlerTimeout)
		defer cancel()

		status := manager.CheckAllComponents(ctx)
		resp := HealthResponse{
			HealthStatus: status,
			BuildInfo:    health.GetBuildInfo(),
			Runtime:      getRuntimeInfo(),
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_ = contract.WriteJSON(w, httpStatusForHealth(status.Status), resp)
	})
}

// ComponentHealthHandler creates a handler for checking a specific component's health.
func ComponentHealthHandler(manager health.HealthManager, componentName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		ctx, cancel := withCheckTimeout(r.Context(), componentHealthTimeout)
		defer cancel()

		_ = manager.CheckComponent(ctx, componentName)

		componentHealth, exists := manager.GetComponentHealth(componentName)
		if !exists {
			contract.WriteError(w, r, contract.APIError{
				Status:   http.StatusNotFound,
				Code:     "COMPONENT_NOT_FOUND",
				Message:  "component not found",
				Category: contract.CategoryForStatus(http.StatusNotFound),
			})
			return
		}

		_ = contract.WriteJSON(w, httpStatusForHealth(componentHealth.Status), componentHealth)
	})
}

// AllComponentsHealthHandler creates a handler for checking all components' health.
func AllComponentsHealthHandler(manager health.HealthManager) http.Handler {
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
func ComponentsListHandler(manager health.HealthManager) http.Handler {
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
