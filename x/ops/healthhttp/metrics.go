package healthhttp

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// MetricsHandler creates a handler that exposes tracked health metrics.
func MetricsHandler(tracker *Tracker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireTracker(tracker, w, r) {
			return
		}
		metrics := tracker.GetMetrics()
		_ = contract.WriteJSON(w, http.StatusOK, metrics)
	})
}

// HealthReportHandler creates a handler that exposes comprehensive health report.
func HealthReportHandler(tracker *Tracker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireTracker(tracker, w, r) {
			return
		}
		report := tracker.GenerateReport()
		_ = contract.WriteJSON(w, httpStatusForHealth(report.HealthStatus.Status), report)
	})
}
