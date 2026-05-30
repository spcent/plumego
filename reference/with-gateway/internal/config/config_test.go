package config

import (
	"testing"
)

func mapLookup(m map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestDefaultsGatewayBackend(t *testing.T) {
	cfg := Defaults()
	if cfg.GatewayBackend == "" {
		t.Fatal("Defaults().GatewayBackend is empty; must have a safe non-empty default")
	}
}

func TestLoadGatewayBackendFromEnv(t *testing.T) {
	cfg, err := load([]string{"with-gateway"}, mapLookup(map[string]string{
		"GATEWAY_BACKEND": "http://backend.internal:8080",
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.GatewayBackend != "http://backend.internal:8080" {
		t.Errorf("GatewayBackend = %q, want %q", cfg.GatewayBackend, "http://backend.internal:8080")
	}
}

func TestValidateRejectsEmptyGatewayBackend(t *testing.T) {
	cfg := Defaults()
	cfg.GatewayBackend = ""
	if err := Validate(cfg); err == nil {
		t.Fatal("Validate: want error for empty GatewayBackend, got nil")
	}
}

func TestLoadDebugFlagFromEnv(t *testing.T) {
	cfg, err := load([]string{"with-gateway"}, mapLookup(map[string]string{
		"APP_DEBUG": "true",
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !cfg.App.Debug {
		t.Error("App.Debug = false, want true when APP_DEBUG=true")
	}
}
