package healthhttp

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// RuntimeInfoHandler exposes runtime diagnostics only.
func RuntimeInfoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_ = contract.WriteJSON(w, http.StatusOK, getRuntimeInfo())
	})
}

// DebugHealthHandler returns comprehensive system diagnostics.
// It should only be mounted in development or internal environments.
func DebugHealthHandler(manager Manager, debug bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !debug {
			http.NotFound(w, r)
			return
		}

		response := map[string]any{
			"timestamp":  time.Now(),
			"build_info": GetBuildInfo(),
			"runtime":    getRuntimeInfo(),
		}

		if manager != nil {
			ctx, cancel := withCheckTimeout(r.Context(), healthHandlerTimeout)
			defer cancel()

			response["readiness"] = manager.Readiness()
			response["health"] = manager.CheckAllComponents(ctx)
			response["components"] = manager.GetAllHealth()
			response["config"] = manager.GetConfig()
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_ = contract.WriteJSON(w, http.StatusOK, response)
	})
}
