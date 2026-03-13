package httpmetrics

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
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
			traceID := internalobs.EnsureTraceID(r)
			w.Header().Set(contract.RequestIDHeader, traceID)
			recorder := internalobs.NewResponseRecorder(w)
			start := time.Now()

			next.ServeHTTP(recorder, r)

			metricsData := internalobs.BuildRequestMetrics(r, recorder, start, traceID)
			path := metricsData.Path
			if metricsData.Route != "" {
				path = metricsData.Route
			}
			collector.ObserveHTTP(r.Context(), metricsData.Method, path, metricsData.Status, metricsData.Bytes, metricsData.Duration)
		})
	}
}
