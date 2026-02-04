package observability

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spcent/plumego/metrics"
	mwobs "github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/router"
)

// MetricsConfig configures the built-in metrics endpoint and collector wiring.
type MetricsConfig struct {
	Enabled   bool
	Path      string
	Namespace string
	MaxSeries int
	Collector metrics.MetricsCollector
	Handler   http.Handler
}

// TracingConfig configures the built-in tracing hook.
type TracingConfig struct {
	Enabled     bool
	ServiceName string
	Tracer      mwobs.Tracer
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

// Configure wires built-in metrics and tracing with structured logging.
func Configure(hooks Hooks, cfg ObservabilityConfig) error {
	if hooks.EnsureMutable == nil {
		return fmt.Errorf("observability hooks missing EnsureMutable")
	}
	if err := hooks.EnsureMutable("configure_observability", "configure observability"); err != nil {
		return err
	}

	if cfg.Metrics.Enabled {
		if err := configureMetrics(hooks, cfg.Metrics); err != nil {
			return err
		}
	}

	if cfg.Tracing.Enabled {
		if err := configureTracing(hooks, cfg.Tracing); err != nil {
			return err
		}
	}

	if cfg.Metrics.Enabled || cfg.Tracing.Enabled {
		if hooks.EnableLogging == nil {
			return fmt.Errorf("observability hooks missing EnableLogging")
		}
		if err := hooks.EnableLogging(); err != nil {
			return err
		}
	}

	return nil
}

func configureMetrics(hooks Hooks, cfg MetricsConfig) error {
	if hooks.EnsureRouter == nil {
		return fmt.Errorf("observability hooks missing EnsureRouter")
	}
	router := hooks.EnsureRouter()

	path := normalizeObservabilityPath(cfg.Path)
	if path == "" {
		path = "/metrics"
	}

	collector := cfg.Collector
	if collector == nil && hooks.GetMetricsCollector != nil {
		collector = hooks.GetMetricsCollector()
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

	if err := router.AddRoute(http.MethodGet, path, handler); err != nil {
		return err
	}

	if hooks.SetMetricsCollector != nil {
		hooks.SetMetricsCollector(collector)
	}
	return nil
}

func configureTracing(hooks Hooks, cfg TracingConfig) error {
	tracer := cfg.Tracer
	if tracer == nil && hooks.GetTracer != nil {
		tracer = hooks.GetTracer()
	}
	if tracer == nil {
		tracer = metrics.NewOpenTelemetryTracer(cfg.ServiceName)
	}

	if hooks.SetTracer != nil {
		hooks.SetTracer(tracer)
	}
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

// Hooks provide access to the app wiring points needed by Configure.
type Hooks struct {
	EnsureMutable       func(op, desc string) error
	EnsureRouter        func() *router.Router
	EnableLogging       func() error
	GetMetricsCollector func() metrics.MetricsCollector
	SetMetricsCollector func(metrics.MetricsCollector)
	GetTracer           func() mwobs.Tracer
	SetTracer           func(mwobs.Tracer)
}
