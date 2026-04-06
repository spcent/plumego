package httpmetrics

import (
	"net/http"

	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
)

type Observer = metrics.HTTPObserver

func Middleware(collector Observer) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		if collector == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			prepared := internalobs.PrepareRequest(w, r)
			r = prepared.Request
			recorder := prepared.Recorder

			next.ServeHTTP(recorder, r)

			metricsData := internalobs.BuildRequestMetrics(r, recorder, prepared.StartedAt, prepared.RequestID)
			path := metricsData.Path
			if metricsData.Route != "" {
				path = metricsData.Route
			}
			collector.ObserveHTTP(r.Context(), metricsData.Method, path, metricsData.Status, metricsData.Bytes, metricsData.Duration)
		})
	}
}
