package core

import (
	"github.com/spcent/plumego/core/components/observability"
	"github.com/spcent/plumego/metrics"
	mwobs "github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/router"
)

// ConfigureObservability wires built-in metrics and tracing with structured logging.
func (a *App) ConfigureObservability(cfg observability.ObservabilityConfig) error {
	return observability.Configure(observability.Hooks{
		EnsureMutable: a.ensureMutable,
		EnsureRouter: func() *router.Router {
			return a.ensureRouter()
		},
		GetPrometheusCollector: func() *metrics.PrometheusCollector {
			a.mu.RLock()
			defer a.mu.RUnlock()
			return a.prometheusMetrics
		},
		SetPrometheusCollector: func(c *metrics.PrometheusCollector) {
			a.mu.Lock()
			a.prometheusMetrics = c
			a.httpMetrics = c
			a.mu.Unlock()
		},
		SetHTTPMetrics: func(observer metrics.HTTPObserver) {
			a.mu.Lock()
			a.httpMetrics = observer
			a.mu.Unlock()
		},
		GetTracer: func() mwobs.Tracer {
			a.mu.RLock()
			defer a.mu.RUnlock()
			return a.tracer
		},
		SetTracer: func(t mwobs.Tracer) {
			a.mu.Lock()
			a.tracer = t
			a.mu.Unlock()
		},
	}, cfg)
}
