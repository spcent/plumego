package observability

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/spcent/plumego/metrics"
	mwobs "github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/router"
)

func TestDefaultMetricsConfig(t *testing.T) {
	cfg := DefaultMetricsConfig()
	if cfg.Enabled {
		t.Fatal("expected metrics disabled by default")
	}
	if cfg.Path != "/metrics" {
		t.Fatalf("expected path /metrics, got %q", cfg.Path)
	}
	if cfg.Namespace != "plumego" {
		t.Fatalf("expected namespace plumego, got %q", cfg.Namespace)
	}
	if cfg.MaxSeries != 10000 {
		t.Fatalf("expected MaxSeries 10000, got %d", cfg.MaxSeries)
	}
}

func TestDefaultTracingConfig(t *testing.T) {
	cfg := DefaultTracingConfig()
	if cfg.Enabled {
		t.Fatal("expected tracing disabled by default")
	}
	if cfg.ServiceName != "plumego" {
		t.Fatalf("expected service name plumego, got %q", cfg.ServiceName)
	}
}

func TestDefaultObservabilityConfig(t *testing.T) {
	cfg := DefaultObservabilityConfig()
	if cfg.Metrics.Enabled || cfg.Tracing.Enabled {
		t.Fatal("expected both disabled by default")
	}
	if cfg.Metrics.Namespace != "plumego" {
		t.Fatalf("expected metrics namespace plumego, got %q", cfg.Metrics.Namespace)
	}
	if cfg.Tracing.ServiceName != "plumego" {
		t.Fatalf("expected tracing service name plumego, got %q", cfg.Tracing.ServiceName)
	}
}

func TestConfigureNilEnsureMutable(t *testing.T) {
	hooks := Hooks{}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true

	err := Configure(hooks, cfg)
	if err == nil {
		t.Fatal("expected error for nil EnsureMutable")
	}
	if err.Error() != "observability hooks missing EnsureMutable" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureEnsureMutableFails(t *testing.T) {
	hooks := Hooks{
		EnsureMutable: func(op, desc string) error {
			return fmt.Errorf("app is already booted")
		},
	}
	cfg := DefaultObservabilityConfig()

	err := Configure(hooks, cfg)
	if err == nil {
		t.Fatal("expected error from EnsureMutable")
	}
	if err.Error() != "app is already booted" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureBothDisabled(t *testing.T) {
	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
	}
	cfg := DefaultObservabilityConfig()

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestConfigureMetricsNilEnsureRouter(t *testing.T) {
	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true

	err := Configure(hooks, cfg)
	if err == nil {
		t.Fatal("expected error for nil EnsureRouter")
	}
	if err.Error() != "observability hooks missing EnsureRouter" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureMetricsNoHandler(t *testing.T) {
	r := router.NewRouter()
	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
		EnsureRouter:  func() *router.Router { return r },
		EnableLogging: func() error { return nil },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.Collector = &noHandlerCollector{}

	err := Configure(hooks, cfg)
	if err == nil {
		t.Fatal("expected error for no handler available")
	}
	if err.Error() != "metrics enabled but no handler available" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureMetricsWithPrometheus(t *testing.T) {
	r := router.NewRouter()
	var setCollector metrics.MetricsCollector

	hooks := Hooks{
		EnsureMutable:       func(op, desc string) error { return nil },
		EnsureRouter:        func() *router.Router { return r },
		EnableLogging:       func() error { return nil },
		SetMetricsCollector: func(c metrics.MetricsCollector) { setCollector = c },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if setCollector == nil {
		t.Fatal("expected collector to be set")
	}
}

func TestConfigureMetricsWithExplicitCollector(t *testing.T) {
	r := router.NewRouter()
	prom := metrics.NewPrometheusCollector("test")
	var setCollector metrics.MetricsCollector

	hooks := Hooks{
		EnsureMutable:       func(op, desc string) error { return nil },
		EnsureRouter:        func() *router.Router { return r },
		EnableLogging:       func() error { return nil },
		SetMetricsCollector: func(c metrics.MetricsCollector) { setCollector = c },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.Collector = prom

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if setCollector != prom {
		t.Fatal("expected the explicit collector to be set")
	}
}

func TestConfigureMetricsWithExplicitHandler(t *testing.T) {
	r := router.NewRouter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
		EnsureRouter:  func() *router.Router { return r },
		EnableLogging: func() error { return nil },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.Handler = handler
	cfg.Metrics.Collector = &noHandlerCollector{}

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureMetricsCollectorFromHooks(t *testing.T) {
	r := router.NewRouter()
	prom := metrics.NewPrometheusCollector("hookns")

	hooks := Hooks{
		EnsureMutable:       func(op, desc string) error { return nil },
		EnsureRouter:        func() *router.Router { return r },
		EnableLogging:       func() error { return nil },
		GetMetricsCollector: func() metrics.MetricsCollector { return prom },
		SetMetricsCollector: func(c metrics.MetricsCollector) {},
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureTracingOnly(t *testing.T) {
	var setTracer mwobs.Tracer

	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
		EnableLogging: func() error { return nil },
		SetTracer:     func(t mwobs.Tracer) { setTracer = t },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Tracing.Enabled = true

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if setTracer == nil {
		t.Fatal("expected tracer to be set")
	}
}

func TestConfigureTracingWithExplicitTracer(t *testing.T) {
	explicit := &stubTracer{}
	var setTracer mwobs.Tracer

	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
		EnableLogging: func() error { return nil },
		SetTracer:     func(t mwobs.Tracer) { setTracer = t },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Tracing.Enabled = true
	cfg.Tracing.Tracer = explicit

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if setTracer != explicit {
		t.Fatal("expected explicit tracer")
	}
}

func TestConfigureTracingFromHooks(t *testing.T) {
	hookTracer := &stubTracer{}
	var setTracer mwobs.Tracer

	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
		EnableLogging: func() error { return nil },
		GetTracer:     func() mwobs.Tracer { return hookTracer },
		SetTracer:     func(t mwobs.Tracer) { setTracer = t },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Tracing.Enabled = true

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if setTracer != hookTracer {
		t.Fatal("expected hook tracer to be used")
	}
}

func TestConfigureEnableLoggingCalledWhenMetricsEnabled(t *testing.T) {
	r := router.NewRouter()
	loggingEnabled := false

	hooks := Hooks{
		EnsureMutable:       func(op, desc string) error { return nil },
		EnsureRouter:        func() *router.Router { return r },
		EnableLogging:       func() error { loggingEnabled = true; return nil },
		SetMetricsCollector: func(c metrics.MetricsCollector) {},
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true

	_ = Configure(hooks, cfg)
	if !loggingEnabled {
		t.Fatal("expected EnableLogging to be called")
	}
}

func TestConfigureEnableLoggingCalledWhenTracingEnabled(t *testing.T) {
	loggingEnabled := false

	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
		EnableLogging: func() error { loggingEnabled = true; return nil },
		SetTracer:     func(t mwobs.Tracer) {},
	}
	cfg := DefaultObservabilityConfig()
	cfg.Tracing.Enabled = true

	_ = Configure(hooks, cfg)
	if !loggingEnabled {
		t.Fatal("expected EnableLogging to be called")
	}
}

func TestConfigureEnableLoggingNotCalledWhenBothDisabled(t *testing.T) {
	loggingCalled := false

	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
		EnableLogging: func() error { loggingCalled = true; return nil },
	}
	cfg := DefaultObservabilityConfig()

	_ = Configure(hooks, cfg)
	if loggingCalled {
		t.Fatal("EnableLogging should not be called when both disabled")
	}
}

func TestConfigureEnableLoggingNilHook(t *testing.T) {
	r := router.NewRouter()

	hooks := Hooks{
		EnsureMutable:       func(op, desc string) error { return nil },
		EnsureRouter:        func() *router.Router { return r },
		SetMetricsCollector: func(c metrics.MetricsCollector) {},
		// EnableLogging is nil
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true

	err := Configure(hooks, cfg)
	if err == nil {
		t.Fatal("expected error for nil EnableLogging")
	}
	if err.Error() != "observability hooks missing EnableLogging" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureEnableLoggingError(t *testing.T) {
	r := router.NewRouter()

	hooks := Hooks{
		EnsureMutable:       func(op, desc string) error { return nil },
		EnsureRouter:        func() *router.Router { return r },
		EnableLogging:       func() error { return fmt.Errorf("logging failed") },
		SetMetricsCollector: func(c metrics.MetricsCollector) {},
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true

	err := Configure(hooks, cfg)
	if err == nil {
		t.Fatal("expected error from EnableLogging")
	}
	if err.Error() != "logging failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeObservabilityPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"/metrics", "/metrics"},
		{"metrics", "/metrics"},
		{"  /custom  ", "/custom"},
		{"  custom  ", "/custom"},
		{"/", "/"},
	}
	for _, tt := range tests {
		got := normalizeObservabilityPath(tt.input)
		if got != tt.want {
			t.Errorf("normalizeObservabilityPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestConfigureMetricsEmptyPath(t *testing.T) {
	r := router.NewRouter()

	hooks := Hooks{
		EnsureMutable:       func(op, desc string) error { return nil },
		EnsureRouter:        func() *router.Router { return r },
		EnableLogging:       func() error { return nil },
		SetMetricsCollector: func(c metrics.MetricsCollector) {},
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.Path = ""

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureMetricsZeroMaxSeries(t *testing.T) {
	r := router.NewRouter()

	hooks := Hooks{
		EnsureMutable:       func(op, desc string) error { return nil },
		EnsureRouter:        func() *router.Router { return r },
		EnableLogging:       func() error { return nil },
		SetMetricsCollector: func(c metrics.MetricsCollector) {},
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.MaxSeries = 0

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureTracingWithNilSetTracer(t *testing.T) {
	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
		EnableLogging: func() error { return nil },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Tracing.Enabled = true

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureMetricsWithNilSetCollector(t *testing.T) {
	r := router.NewRouter()

	hooks := Hooks{
		EnsureMutable: func(op, desc string) error { return nil },
		EnsureRouter:  func() *router.Router { return r },
		EnableLogging: func() error { return nil },
	}
	cfg := DefaultObservabilityConfig()
	cfg.Metrics.Enabled = true

	err := Configure(hooks, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// noHandlerCollector implements MetricsCollector but does NOT have a Handler() method.
type noHandlerCollector struct{}

func (c *noHandlerCollector) Record(_ context.Context, _ metrics.MetricRecord)            {}
func (c *noHandlerCollector) ObserveHTTP(_ context.Context, _, _ string, _, _ int, _ time.Duration) {
}
func (c *noHandlerCollector) ObservePubSub(_ context.Context, _, _ string, _ time.Duration, _ error) {
}
func (c *noHandlerCollector) ObserveMQ(_ context.Context, _, _ string, _ time.Duration, _ error, _ bool) {
}
func (c *noHandlerCollector) ObserveKV(_ context.Context, _, _ string, _ time.Duration, _ error, _ bool) {
}
func (c *noHandlerCollector) ObserveIPC(_ context.Context, _, _, _ string, _ int, _ time.Duration, _ error) {
}
func (c *noHandlerCollector) ObserveDB(_ context.Context, _, _, _ string, _ int, _ time.Duration, _ error) {
}
func (c *noHandlerCollector) GetStats() metrics.CollectorStats { return metrics.CollectorStats{} }
func (c *noHandlerCollector) Clear()                           {}

// stubTracer satisfies mwobs.Tracer.
type stubTracer struct{}

func (s *stubTracer) Start(ctx context.Context, _ *http.Request) (context.Context, mwobs.TraceSpan) {
	return ctx, &stubSpan{}
}

type stubSpan struct{}

func (s *stubSpan) End(_ mwobs.RequestMetrics) {}
