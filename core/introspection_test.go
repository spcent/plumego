package core

import (
	"net/http"
	"testing"
	"time"
)

func TestRuntimeSnapshotIncludesStableFields(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = ":9090"
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
	app := New(cfg, AppDependencies{})
	app.preparationState = PreparationStateServerPrepared
	app.httpServer = &http.Server{}

	snapshot := app.RuntimeSnapshot()

	if snapshot.Addr != ":9090" {
		t.Fatalf("addr = %q, want %q", snapshot.Addr, ":9090")
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
	if snapshot.PreparationState != PreparationStateServerPrepared {
		t.Fatalf("preparation_state = %q, want %q", snapshot.PreparationState, PreparationStateServerPrepared)
	}
}

func TestRuntimeSnapshotNilApp(t *testing.T) {
	var app *App
	snapshot := app.RuntimeSnapshot()

	if snapshot.Addr != "" {
		t.Fatalf("addr = %q, want empty", snapshot.Addr)
	}
	if snapshot.PreparationState != "" {
		t.Fatalf("preparation_state = %q, want empty", snapshot.PreparationState)
	}
}
