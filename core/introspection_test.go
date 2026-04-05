package core

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRuntimeSnapshotIncludesStableFields(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = ":9090"
	cfg.EnvFile = ".env.test"
	cfg.Debug = true
	cfg.ShutdownTimeout = 7 * time.Second
	cfg.ReadTimeout = 11 * time.Second
	cfg.ReadHeaderTimeout = 3 * time.Second
	cfg.WriteTimeout = 13 * time.Second
	cfg.IdleTimeout = 17 * time.Second
	cfg.MaxHeaderBytes = 2048
	cfg.EnableHTTP2 = false
	cfg.DrainInterval = 250 * time.Millisecond
	cfg.TLS = TLSConfig{
		Enabled:  true,
		CertFile: "cert.pem",
		KeyFile:  "key.pem",
	}
	app := New(cfg)
	app.started = true
	app.configFrozen = true
	app.httpServer = &http.Server{}

	snapshot := app.RuntimeSnapshot()

	if snapshot.Addr != ":9090" {
		t.Fatalf("addr = %q, want %q", snapshot.Addr, ":9090")
	}
	if snapshot.EnvFile != ".env.test" {
		t.Fatalf("env file = %q, want %q", snapshot.EnvFile, ".env.test")
	}
	if !snapshot.Debug {
		t.Fatal("expected debug=true")
	}
	if snapshot.ShutdownTimeout != 7*time.Second {
		t.Fatalf("shutdown timeout = %v, want %v", snapshot.ShutdownTimeout, 7*time.Second)
	}
	if snapshot.ReadTimeout != 11*time.Second {
		t.Fatalf("read timeout = %v, want %v", snapshot.ReadTimeout, 11*time.Second)
	}
	if snapshot.ReadHeaderTimeout != 3*time.Second {
		t.Fatalf("read header timeout = %v, want %v", snapshot.ReadHeaderTimeout, 3*time.Second)
	}
	if snapshot.WriteTimeout != 13*time.Second {
		t.Fatalf("write timeout = %v, want %v", snapshot.WriteTimeout, 13*time.Second)
	}
	if snapshot.IdleTimeout != 17*time.Second {
		t.Fatalf("idle timeout = %v, want %v", snapshot.IdleTimeout, 17*time.Second)
	}
	if snapshot.MaxHeaderBytes != 2048 {
		t.Fatalf("max header bytes = %d, want %d", snapshot.MaxHeaderBytes, 2048)
	}
	if snapshot.HTTP2Enabled {
		t.Fatal("expected http2_enabled=false")
	}
	if snapshot.DrainInterval != 250*time.Millisecond {
		t.Fatalf("drain interval = %v, want %v", snapshot.DrainInterval, 250*time.Millisecond)
	}
	if !snapshot.TLS.Enabled || snapshot.TLS.CertFile != "cert.pem" || snapshot.TLS.KeyFile != "key.pem" {
		t.Fatalf("unexpected tls snapshot: %+v", snapshot.TLS)
	}
	if !snapshot.Started {
		t.Fatal("expected started=true")
	}
	if !snapshot.ConfigFrozen {
		t.Fatal("expected config_frozen=true")
	}
	if !snapshot.ServerPrepared {
		t.Fatal("expected server_prepared=true")
	}
}

func TestRuntimeSnapshotNilApp(t *testing.T) {
	var app *App
	snapshot := app.RuntimeSnapshot()

	if snapshot.Debug {
		t.Fatal("expected zero snapshot for nil app")
	}
	if snapshot.EnvFile != "" {
		t.Fatalf("env file = %q, want empty", snapshot.EnvFile)
	}
}

func TestAttachHTTPObserverComposesObservers(t *testing.T) {
	app := newTestApp()
	first := &recordingHTTPObserver{}
	second := &recordingHTTPObserver{}

	app.AttachHTTPObserver(first)
	app.AttachHTTPObserver(nil)
	app.AttachHTTPObserver(second)

	if app.httpMetrics == nil {
		t.Fatal("expected httpMetrics to be attached")
	}

	app.httpMetrics.ObserveHTTP(context.Background(), http.MethodGet, "/metrics", http.StatusOK, 128, 5*time.Millisecond)

	if first.calls != 1 {
		t.Fatalf("first observer calls = %d, want 1", first.calls)
	}
	if second.calls != 1 {
		t.Fatalf("second observer calls = %d, want 1", second.calls)
	}
}

func TestAttachHTTPObserverNilAppAndObserver(t *testing.T) {
	var app *App
	observer := &recordingHTTPObserver{}

	app.AttachHTTPObserver(observer)

	app = newTestApp()
	app.AttachHTTPObserver(nil)

	if app.httpMetrics != nil {
		t.Fatal("expected nil observer to be ignored")
	}
}

type recordingHTTPObserver struct {
	calls int
}

func (o *recordingHTTPObserver) ObserveHTTP(context.Context, string, string, int, int, time.Duration) {
	o.calls++
}
