package health

import (
	"encoding/json"
	"net/http"
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(status)
	})
}

// BuildInfoHandler exposes build metadata as JSON for diagnostics and release verification.
func BuildInfoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GetBuildInfo())
	})
}
