package core

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

// MetricsConfig configures the built-in metrics endpoint and collector wiring.
type MetricsConfig struct {
	Enabled   bool
	Path      string
	Namespace string
	MaxSeries int
	Collector middleware.MetricsCollector
	Handler   http.Handler
}

// TracingConfig configures the built-in tracing hook.
type TracingConfig struct {
	Enabled     bool
	ServiceName string
	Tracer      middleware.Tracer
}

// ObservabilityConfig combines metrics and tracing settings.
type ObservabilityConfig struct {
	Metrics MetricsConfig
	Tracing TracingConfig
}

// DefaultMetricsConfig returns baseline metrics defaults.
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Enabled:   false,
		Path:      "/metrics",
		Namespace: "plumego",
		MaxSeries: 10000,
	}
}

// DefaultTracingConfig returns baseline tracing defaults.
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		Enabled:     false,
		ServiceName: "plumego",
	}
}

// DefaultObservabilityConfig returns baseline observability defaults.
func DefaultObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		Metrics: DefaultMetricsConfig(),
		Tracing: DefaultTracingConfig(),
	}
}

// ConfigureObservability wires built-in metrics and tracing with structured logging.
func (a *App) ConfigureObservability(cfg ObservabilityConfig) error {
	if cfg.Metrics.Enabled {
		if err := a.configureMetrics(cfg.Metrics); err != nil {
			return err
		}
	}

	if cfg.Tracing.Enabled {
		if err := a.configureTracing(cfg.Tracing); err != nil {
			return err
		}
	}

	if cfg.Metrics.Enabled || cfg.Tracing.Enabled {
		if err := a.enableLogging(); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) configureMetrics(cfg MetricsConfig) error {
	if a.router == nil {
		a.router = router.NewRouter()
		a.router.SetLogger(a.logger)
	}

	path := normalizeObservabilityPath(cfg.Path)
	if path == "" {
		path = "/metrics"
	}

	collector := cfg.Collector
	if collector == nil {
		collector = a.metricsCollector
	}

	if collector == nil {
		prom := metrics.NewPrometheusCollector(cfg.Namespace)
		if cfg.MaxSeries > 0 {
			prom.WithMaxMemory(cfg.MaxSeries)
		}
		collector = prom
	}

	handler := cfg.Handler
	if handler == nil {
		if provider, ok := collector.(interface{ Handler() http.Handler }); ok {
			handler = provider.Handler()
		}
	}

	if handler == nil {
		return fmt.Errorf("metrics enabled but no handler available")
	}

	if err := a.router.AddRoute(http.MethodGet, path, handler); err != nil {
		return err
	}

	a.metricsCollector = collector
	return nil
}

func (a *App) configureTracing(cfg TracingConfig) error {
	tracer := cfg.Tracer
	if tracer == nil {
		tracer = a.tracer
	}
	if tracer == nil {
		tracer = metrics.NewOpenTelemetryTracer(cfg.ServiceName)
	}

	a.tracer = tracer
	return nil
}

func normalizeObservabilityPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}
