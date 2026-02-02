package health

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// MetricsHandler creates a handler that exposes health metrics.
func MetricsHandler(collector *MetricsCollector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics := collector.GetMetrics()
		_ = contract.WriteJSON(w, http.StatusOK, metrics)
	})
}

// HealthReportHandler creates a handler that exposes comprehensive health report.
func HealthReportHandler(collector *MetricsCollector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		report := collector.GenerateReport()

		code := http.StatusOK
		switch report.HealthStatus.Status {
		case StatusUnhealthy:
			code = http.StatusServiceUnavailable
		case StatusDegraded:
			code = http.StatusPartialContent
		}

		_ = contract.WriteJSON(w, code, report)
	})
}
