package core

import (
	"testing"
	"time"
)

func TestDefaultConfigAllFields(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", cfg.Addr)
	}
	if cfg.TLS.Enabled {
		t.Errorf("TLS.Enabled = true, want false")
	}
	if cfg.TLS.CertFile != "" {
		t.Errorf("TLS.CertFile = %q, want empty", cfg.TLS.CertFile)
	}
	if cfg.TLS.KeyFile != "" {
		t.Errorf("TLS.KeyFile = %q, want empty", cfg.TLS.KeyFile)
	}
	if cfg.Router.MethodNotAllowed {
		t.Errorf("Router.MethodNotAllowed = true, want false")
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("ReadTimeout = %v, want 30s", cfg.ReadTimeout)
	}
	if cfg.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 5s", cfg.ReadHeaderTimeout)
	}
	if cfg.WriteTimeout != 30*time.Second {
		t.Errorf("WriteTimeout = %v, want 30s", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 60*time.Second {
		t.Errorf("IdleTimeout = %v, want 60s", cfg.IdleTimeout)
	}
	if cfg.MaxHeaderBytes != 1<<20 {
		t.Errorf("MaxHeaderBytes = %d, want %d (1 MiB)", cfg.MaxHeaderBytes, 1<<20)
	}
	if !cfg.HTTP2Enabled {
		t.Errorf("HTTP2Enabled = false, want true")
	}
	if cfg.DrainInterval != defaultDrainInterval {
		t.Errorf("DrainInterval = %v, want %v", cfg.DrainInterval, defaultDrainInterval)
	}
}

// TestDefaultConfigIsIndependent verifies that DefaultConfig returns a fresh
// value each time, so callers cannot accidentally share mutable state.
func TestDefaultConfigIsIndependent(t *testing.T) {
	a := DefaultConfig()
	b := DefaultConfig()

	a.Addr = ":9999"
	a.ReadTimeout = 1 * time.Hour

	if b.Addr == a.Addr {
		t.Errorf("DefaultConfig returns shared Addr; want independent values")
	}
	if b.ReadTimeout == a.ReadTimeout {
		t.Errorf("DefaultConfig returns shared ReadTimeout; want independent values")
	}
}

func TestValidateServerConfigRejectsEmpty(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*AppConfig)
		wantMsg string
	}{
		{
			name:    "blank address",
			mutate:  func(c *AppConfig) { c.Addr = "   " },
			wantMsg: "server address cannot be empty",
		},
		{
			name:    "negative read timeout",
			mutate:  func(c *AppConfig) { c.ReadTimeout = -time.Millisecond },
			wantMsg: "read timeout cannot be negative",
		},
		{
			name:    "negative read header timeout",
			mutate:  func(c *AppConfig) { c.ReadHeaderTimeout = -time.Millisecond },
			wantMsg: "read header timeout cannot be negative",
		},
		{
			name:    "negative write timeout",
			mutate:  func(c *AppConfig) { c.WriteTimeout = -time.Millisecond },
			wantMsg: "write timeout cannot be negative",
		},
		{
			name:    "negative idle timeout",
			mutate:  func(c *AppConfig) { c.IdleTimeout = -time.Millisecond },
			wantMsg: "idle timeout cannot be negative",
		},
		{
			name:    "negative max header bytes",
			mutate:  func(c *AppConfig) { c.MaxHeaderBytes = -1 },
			wantMsg: "max header bytes cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.mutate(&cfg)
			err := validateServerConfig(cfg)
			if err == nil {
				t.Fatalf("validateServerConfig accepted invalid config")
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestValidateServerConfigAcceptsZeroTimeouts(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ReadTimeout = 0
	cfg.ReadHeaderTimeout = 0
	cfg.WriteTimeout = 0
	cfg.IdleTimeout = 0
	cfg.MaxHeaderBytes = 0

	if err := validateServerConfig(cfg); err != nil {
		t.Errorf("validateServerConfig rejected zero timeouts: %v", err)
	}
}
