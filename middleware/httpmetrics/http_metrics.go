package httpmetrics

import (
	"net/http"

	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	internaltelemetry "github.com/spcent/plumego/middleware/internal/telemetry"
)

type Observer = metrics.HTTPObserver

func Middleware(collector Observer) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		if collector == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			prepared := internaltelemetry.PrepareRequest(w, r)
			r = prepared.Request
			recorder := prepared.Recorder

			defer internaltelemetry.FinishPreservingPanic(func() {
				metricsData := prepared.Complete(r)
				collector.ObserveHTTP(r.Context(), metricsData.Method, metricsData.ObservedPath(), metricsData.Status, metricsData.Bytes, metricsData.Duration)
			})

			next.ServeHTTP(recorder, r)
		})
	}
}
