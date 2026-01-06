package core

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/spcent/plumego/health"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	webhookin "github.com/spcent/plumego/net/webhookin"
	webhookout "github.com/spcent/plumego/net/webhookout"
	"github.com/spcent/plumego/pubsub"
	"github.com/spcent/plumego/router"
)

func TestWithRouter(t *testing.T) {
	app := &App{}
	r := router.NewRouter()
	opt := WithRouter(r)
	opt(app)
	if app.router != r {
		t.Errorf("expected router to be set")
	}
}

func TestWithAddr(t *testing.T) {
	app := &App{config: &AppConfig{}}
	opt := WithAddr(":8080")
	opt(app)
	if app.config.Addr != ":8080" {
		t.Errorf("expected addr to be :8080, got %s", app.config.Addr)
	}
}

func TestWithEnvPath(t *testing.T) {
	app := &App{config: &AppConfig{}}
	opt := WithEnvPath("/path/to/.env")
	opt(app)
	if app.config.EnvFile != "/path/to/.env" {
		t.Errorf("expected env file to be /path/to/.env, got %s", app.config.EnvFile)
	}
}

func TestWithShutdownTimeout(t *testing.T) {
	app := &App{config: &AppConfig{}}
	timeout := 30 * time.Second
	opt := WithShutdownTimeout(timeout)
	opt(app)
	if app.config.ShutdownTimeout != timeout {
		t.Errorf("expected shutdown timeout to be %v, got %v", timeout, app.config.ShutdownTimeout)
	}
}

func TestWithServerTimeouts(t *testing.T) {
	app := &App{config: &AppConfig{}}
	read := 5 * time.Second
	readHeader := 2 * time.Second
	write := 10 * time.Second
	idle := 120 * time.Second
	opt := WithServerTimeouts(read, readHeader, write, idle)
	opt(app)
	if app.config.ReadTimeout != read {
		t.Errorf("expected read timeout to be %v, got %v", read, app.config.ReadTimeout)
	}
	if app.config.ReadHeaderTimeout != readHeader {
		t.Errorf("expected read header timeout to be %v, got %v", readHeader, app.config.ReadHeaderTimeout)
	}
	if app.config.WriteTimeout != write {
		t.Errorf("expected write timeout to be %v, got %v", write, app.config.WriteTimeout)
	}
	if app.config.IdleTimeout != idle {
		t.Errorf("expected idle timeout to be %v, got %v", idle, app.config.IdleTimeout)
	}
}

func TestWithMaxHeaderBytes(t *testing.T) {
	app := &App{config: &AppConfig{}}
	bytes := 8192
	opt := WithMaxHeaderBytes(bytes)
	opt(app)
	if app.config.MaxHeaderBytes != bytes {
		t.Errorf("expected max header bytes to be %d, got %d", bytes, app.config.MaxHeaderBytes)
	}
}

func TestWithMaxBodyBytes(t *testing.T) {
	app := &App{config: &AppConfig{}}
	bytes := int64(1024 * 1024)
	opt := WithMaxBodyBytes(bytes)
	opt(app)
	if app.config.MaxBodyBytes != bytes {
		t.Errorf("expected max body bytes to be %d, got %d", bytes, app.config.MaxBodyBytes)
	}
}

func TestWithPubSub(t *testing.T) {
	app := &App{}
	ps := pubsub.New()
	opt := WithPubSub(ps)
	opt(app)
	if app.pub != ps {
		t.Errorf("expected pubsub to be set")
	}
}

func TestWithPubSubDebug(t *testing.T) {
	app := &App{config: &AppConfig{}}
	cfg := PubSubConfig{Enabled: true, Path: "/debug/pubsub"}
	opt := WithPubSubDebug(cfg)
	opt(app)
	if app.config.PubSub != cfg {
		t.Errorf("expected pubsub config to be set")
	}
}

func TestWithWebhookOut(t *testing.T) {
	app := &App{config: &AppConfig{}}
	svc := &webhookout.Service{}
	cfg := WebhookOutConfig{Enabled: true, BasePath: "/webhooks", Service: svc}
	opt := WithWebhookOut(cfg)
	opt(app)
	if app.config.WebhookOut != cfg {
		t.Errorf("expected webhook out config to be set")
	}
}

func TestWithWebhookIn(t *testing.T) {
	app := &App{config: &AppConfig{}}
	deduper := webhookin.NewDeduper(10 * time.Minute)
	cfg := WebhookInConfig{Enabled: true, GitHubPath: "/github", Deduper: deduper}
	opt := WithWebhookIn(cfg)
	opt(app)
	if app.config.WebhookIn != cfg {
		t.Errorf("expected webhook in config to be set")
	}
}

func TestWithConcurrencyLimits(t *testing.T) {
	app := &App{config: &AppConfig{}}
	maxConcurrent := 100
	queueDepth := 1000
	queueTimeout := 5 * time.Second
	opt := WithConcurrencyLimits(maxConcurrent, queueDepth, queueTimeout)
	opt(app)
	if app.config.MaxConcurrency != maxConcurrent {
		t.Errorf("expected max concurrency to be %d, got %d", maxConcurrent, app.config.MaxConcurrency)
	}
	if app.config.QueueDepth != queueDepth {
		t.Errorf("expected queue depth to be %d, got %d", queueDepth, app.config.QueueDepth)
	}
	if app.config.QueueTimeout != queueTimeout {
		t.Errorf("expected queue timeout to be %v, got %v", queueTimeout, app.config.QueueTimeout)
	}
}

func TestWithHTTP2(t *testing.T) {
	app := &App{config: &AppConfig{}}
	opt := WithHTTP2(true)
	opt(app)
	if !app.config.EnableHTTP2 {
		t.Errorf("expected HTTP2 to be enabled")
	}
}

func TestWithTLS(t *testing.T) {
	app := &App{config: &AppConfig{}}
	opt := WithTLS("/path/cert.pem", "/path/key.pem")
	opt(app)
	if !app.config.TLS.Enabled {
		t.Errorf("expected TLS to be enabled")
	}
	if app.config.TLS.CertFile != "/path/cert.pem" {
		t.Errorf("expected cert file to be /path/cert.pem, got %s", app.config.TLS.CertFile)
	}
	if app.config.TLS.KeyFile != "/path/key.pem" {
		t.Errorf("expected key file to be /path/key.pem, got %s", app.config.TLS.KeyFile)
	}
}

func TestWithTLSConfig(t *testing.T) {
	app := &App{config: &AppConfig{}}
	tlsConfig := TLSConfig{Enabled: true, CertFile: "/cert", KeyFile: "/key"}
	opt := WithTLSConfig(tlsConfig)
	opt(app)
	if app.config.TLS != tlsConfig {
		t.Errorf("expected TLS config to be set")
	}
}

func TestWithDebug(t *testing.T) {
	app := &App{config: &AppConfig{}}
	opt := WithDebug()
	opt(app)
	if !app.config.Debug {
		t.Errorf("expected debug to be enabled")
	}
}

func TestWithLogger(t *testing.T) {
	app := &App{}
	logger := log.NewGLogger()
	opt := WithLogger(logger)
	opt(app)
	if app.logger != logger {
		t.Errorf("expected logger to be set")
	}
}

func TestWithLoggerNil(t *testing.T) {
	app := &App{logger: log.NewGLogger()}
	originalLogger := app.logger
	opt := WithLogger(nil)
	opt(app)
	if app.logger != originalLogger {
		t.Errorf("expected logger to remain unchanged when nil is passed")
	}
}

func TestWithComponent(t *testing.T) {
	app := &App{}
	comp := &mockComponent{}
	opt := WithComponent(comp)
	opt(app)
	if len(app.components) != 1 {
		t.Errorf("expected 1 component, got %d", len(app.components))
	}
	if app.components[0] != comp {
		t.Errorf("expected component to be set")
	}
}

func TestWithComponents(t *testing.T) {
	app := &App{}
	comp1 := &mockComponent{}
	comp2 := &mockComponent{}
	opt := WithComponents(comp1, comp2)
	opt(app)
	if len(app.components) != 2 {
		t.Errorf("expected 2 components, got %d", len(app.components))
	}
}

func TestWithMetricsCollector(t *testing.T) {
	app := &App{}
	collector := &mockMetricsCollector{}
	opt := WithMetricsCollector(collector)
	opt(app)
	if app.metricsCollector != collector {
		t.Errorf("expected metrics collector to be set")
	}
}

func TestWithTracer(t *testing.T) {
	app := &App{}
	tracer := &mockTracer{}
	opt := WithTracer(tracer)
	opt(app)
	if app.tracer != tracer {
		t.Errorf("expected tracer to be set")
	}
}

// Mock implementations for testing
type mockComponent struct{}

func (m *mockComponent) RegisterRoutes(r *router.Router)             {}
func (m *mockComponent) RegisterMiddleware(reg *middleware.Registry) {}
func (m *mockComponent) Start(ctx context.Context) error             { return nil }
func (m *mockComponent) Stop(ctx context.Context) error              { return nil }
func (m *mockComponent) Health() (string, health.HealthStatus) {
	return "mock", health.HealthStatus{Status: health.StatusHealthy}
}

type mockMetricsCollector struct{}

func (m *mockMetricsCollector) IncCounter(name string, tags map[string]string)                      {}
func (m *mockMetricsCollector) ObserveHistogram(name string, value float64, tags map[string]string) {}
func (m *mockMetricsCollector) SetGauge(name string, value float64, tags map[string]string)         {}
func (m *mockMetricsCollector) Observe(ctx context.Context, metrics middleware.RequestMetrics)      {}

type mockTracer struct{}

func (m *mockTracer) Start(ctx context.Context, r *http.Request) (context.Context, middleware.TraceSpan) {
	return ctx, &mockSpan{}
}
func (m *mockTracer) StartSpan(name string) any              { return nil }
func (m *mockTracer) EndSpan(span any, err error)            {}
func (m *mockTracer) SetTag(span any, key string, value any) {}
func (m *mockTracer) Log(span any, fields map[string]any)    {}

type mockSpan struct{}

func (m *mockSpan) End(metrics middleware.RequestMetrics) {}
