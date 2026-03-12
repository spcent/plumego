package health

import (
	"net/http"
	"runtime"
	"time"

	"github.com/spcent/plumego/contract"
)

// RuntimeInfo contains runtime diagnostics for development and debug endpoints.
type RuntimeInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	GOARCH       string `json:"goarch"`
	GOOS         string `json:"goos"`
	MemAlloc     uint64 `json:"mem_alloc_bytes"`
	MemSys       uint64 `json:"mem_sys_bytes"`
	NumGC        uint32 `json:"num_gc"`
}

// getRuntimeInfo collects current Go runtime diagnostics.
func getRuntimeInfo() *RuntimeInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &RuntimeInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		GOARCH:       runtime.GOARCH,
		GOOS:         runtime.GOOS,
		MemAlloc:     m.Alloc,
		MemSys:       m.Sys,
		NumGC:        m.NumGC,
	}
}

// RuntimeInfoHandler exposes runtime diagnostics only.
func RuntimeInfoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_ = contract.WriteJSON(w, http.StatusOK, getRuntimeInfo())
	})
}

// DebugHealthHandler returns comprehensive system diagnostics.
// It should only be mounted in development or internal environments —
// the caller (route registration) decides when to register this handler.
// Pass debug=false to return 404, preventing accidental exposure.
func DebugHealthHandler(manager HealthManager, debug bool) http.Handler {
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
