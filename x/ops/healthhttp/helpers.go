package healthhttp

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
)

const (
	healthHandlerTimeout       = 10 * time.Second
	componentHealthTimeout     = 5 * time.Second
	allComponentsHealthTimeout = 15 * time.Second
	readinessHandlerTimeout    = 5 * time.Second
)

func withCheckTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		return ctx, func() {}
	}

	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining <= timeout {
			return ctx, func() {}
		}
	}

	return context.WithTimeout(ctx, timeout)
}

func requireManager(manager Manager, w http.ResponseWriter, r *http.Request) bool {
	if manager == nil {
		contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusServiceUnavailable).
			Code("HEALTH_MANAGER_UNAVAILABLE").
			Message("health manager is not configured").
			Build())
		return false
	}
	return true
}

func requireTracker(tracker *Tracker, w http.ResponseWriter, r *http.Request) bool {
	if tracker == nil {
		contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusServiceUnavailable).
			Code("HEALTH_TRACKER_UNAVAILABLE").
			Message("health tracker is not configured").
			Build())
		return false
	}
	return true
}

func httpStatusForHealth(state health.HealthState) int {
	switch state {
	case health.StatusUnhealthy:
		return http.StatusServiceUnavailable
	case health.StatusDegraded:
		return http.StatusPartialContent
	default:
		return http.StatusOK
	}
}

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
