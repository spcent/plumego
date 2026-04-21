package main

import (
	"testing"
	"time"
)

func TestLoadServerConfigDefaults(t *testing.T) {
	cfg, err := loadServerConfig(func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}
	if cfg.Core.Addr != ":8080" {
		t.Fatalf("addr = %q, want :8080", cfg.Core.Addr)
	}
	if cfg.ShutdownTimeout != defaultShutdownTimeout {
		t.Fatalf("shutdown timeout = %s, want %s", cfg.ShutdownTimeout, defaultShutdownTimeout)
	}
}

func TestLoadServerConfigFromEnv(t *testing.T) {
	values := map[string]string{
		"WORKERFLEET_HTTP_ADDR":        ":9090",
		"WORKERFLEET_SHUTDOWN_TIMEOUT": "15s",
	}
	cfg, err := loadServerConfig(func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	})
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}
	if cfg.Core.Addr != ":9090" {
		t.Fatalf("addr = %q, want :9090", cfg.Core.Addr)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Fatalf("shutdown timeout = %s, want 15s", cfg.ShutdownTimeout)
	}
}

func TestLoadServerConfigRejectsInvalidTimeout(t *testing.T) {
	_, err := loadServerConfig(func(key string) (string, bool) {
		if key == "WORKERFLEET_SHUTDOWN_TIMEOUT" {
			return "0s", true
		}
		return "", false
	})
	if err == nil {
		t.Fatalf("expected invalid shutdown timeout error")
	}
}
