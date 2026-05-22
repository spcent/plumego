package handler

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/metrics"
)

// MetricsCollector is the minimal interface that OpsHandler depends on.
// *metrics.BaseMetricsCollector satisfies this interface; tests may pass a stub.
type MetricsCollector interface {
	GetStats() metrics.CollectorStats
}

// OpsHandler serves GET /ops/metrics.
// It returns in-process HTTP request metrics collected by the httpmetrics middleware.
type OpsHandler struct {
	Metrics MetricsCollector
}

// MetricStats handles GET /ops/metrics.
// This route is protected by bearer-token auth wired in routes.go.
func (h OpsHandler) MetricStats(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, h.Metrics.GetStats(), nil)
}
