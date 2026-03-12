package health

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// MetricsHandler creates a handler that exposes health metrics.
func MetricsHandler(tracker *MetricsTracker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics := tracker.GetMetrics()
		_ = contract.WriteJSON(w, http.StatusOK, metrics)
	})
}

// HealthReportHandler creates a handler that exposes comprehensive health report.
func HealthReportHandler(tracker *MetricsTracker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		report := tracker.GenerateReport()
		_ = contract.WriteJSON(w, httpStatusForHealth(report.HealthStatus.Status), report)
	})
}
