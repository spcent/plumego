package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/requestid"
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
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when nil logger is passed")
		}
	}()

	WithLogger(nil)(&App{})
}

func TestNewDefaultsToNoOpLogger(t *testing.T) {
	app := New()
	if app.logger == nil {
		t.Fatal("expected logger to be initialized")
	}
	if _, ok := app.logger.(*log.NoOpLogger); !ok {
		t.Fatalf("expected NoOpLogger by default, got %T", app.logger)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	app := New()
	app.Use(requestid.Middleware())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	app.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatalf("expected X-Request-ID to be set")
	}
}

func TestWithMethodNotAllowed(t *testing.T) {
	app := New(WithMethodNotAllowed(true))
	app.Get("/only", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("expected Allow header to include GET")
	}
}

func TestWithHTTPMetrics(t *testing.T) {
	app := &App{}
	collector := &mockMetricsCollector{
		NoopCollector: metrics.NewNoopCollector(),
	}
	opt := WithHTTPMetrics(collector)
	opt(app)
	if app.httpMetrics == nil {
		t.Errorf("expected HTTP metrics observer to be set")
	}
}

// mockMetricsCollector embeds NoopCollector for cleaner mock implementation.
// NoopCollector already satisfies HTTPObserver, so this mock stays small.
type mockMetricsCollector struct {
	*metrics.NoopCollector
}
