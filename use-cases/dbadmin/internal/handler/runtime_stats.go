package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// RuntimeStatsHandler exposes runtime statistics for monitoring goroutine leaks
// and memory usage in production.
type RuntimeStatsHandler struct {
	Logger     plumelog.StructuredLogger
	StartTime  time.Time
}

// RuntimeStats represents runtime statistics.
type RuntimeStats struct {
	Goroutines     int           `json:"goroutines"`
	GCPauses       uint64        `json:"gc_pauses_ns"`
	HeapAlloc      uint64        `json:"heap_alloc_bytes"`
	HeapSys        uint64        `json:"heap_sys_bytes"`
	HeapInuse      uint64        `json:"heap_inuse_bytes"`
	HeapIdle       uint64        `json:"heap_idle_bytes"`
	HeapReleased   uint64        `json:"heap_released_bytes"`
	HeapObjects    uint64        `json:"heap_objects"`
	StackInuse     uint64        `json:"stack_inuse_bytes"`
	StackSys       uint64        `json:"stack_sys_bytes"`
	NumGC          uint32        `json:"num_gc"`
	LastGC         string        `json:"last_gc"`
	Uptime         string        `json:"uptime"`
	UptimeSeconds  int64         `json:"uptime_seconds"`
	GoVersion      string        `json:"go_version"`
	NumCPU         int           `json:"num_cpu"`
}

// GetStats returns current runtime statistics.
func (h RuntimeStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(h.StartTime)

	stats := RuntimeStats{
		Goroutines:    runtime.NumGoroutine(),
		GCPauses:      m.PauseTotalNs,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapInuse:     m.HeapInuse,
		HeapIdle:      m.HeapIdle,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		NumGC:         m.NumGC,
		LastGC:        time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
		Uptime:        uptime.String(),
		UptimeSeconds: int64(uptime.Seconds()),
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, stats, nil))
}
