package core

import (
	"github.com/spcent/plumego/core/components/observability"
	"github.com/spcent/plumego/metrics"
	mwobs "github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/router"
)

type MetricsConfig = observability.MetricsConfig
type TracingConfig = observability.TracingConfig
type ObservabilityConfig = observability.ObservabilityConfig

func DefaultMetricsConfig() MetricsConfig {
	return observability.DefaultMetricsConfig()
}

func DefaultTracingConfig() TracingConfig {
	return observability.DefaultTracingConfig()
}

func DefaultObservabilityConfig() ObservabilityConfig {
	return observability.DefaultObservabilityConfig()
}

// ConfigureObservability wires built-in metrics and tracing with structured logging.
func (a *App) ConfigureObservability(cfg ObservabilityConfig) error {
	return observability.Configure(observability.Hooks{
		EnsureMutable: a.ensureMutable,
		EnsureRouter: func() *router.Router {
			return a.ensureRouter()
		},
		EnableLogging: a.enableLogging,
		GetMetricsCollector: func() metrics.MetricsCollector {
			a.mu.RLock()
			defer a.mu.RUnlock()
			return a.metricsCollector
		},
		SetMetricsCollector: func(c metrics.MetricsCollector) {
			a.mu.Lock()
			a.metricsCollector = c
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
