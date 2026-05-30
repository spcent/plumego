package config

import (
	"testing"
)

func TestDefaultsAdminToken(t *testing.T) {
	cfg := Defaults()
	if cfg.AdminToken == "" {
		t.Fatal("Defaults().AdminToken is empty; must have a safe non-empty default for local dev")
	}
}

func TestValidateRejectsEmptyAdminToken(t *testing.T) {
	cfg := Defaults()
	cfg.AdminToken = ""
	if err := Validate(cfg); err == nil {
		t.Fatal("Validate: want error for empty AdminToken, got nil")
	}
}

func TestLoadAdminTokenFromEnv(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "prod-secret-token")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AdminToken != "prod-secret-token" {
		t.Errorf("AdminToken = %q, want %q", cfg.AdminToken, "prod-secret-token")
	}
}

func TestLoadEmptyAdminTokenFromEnvPreservesDefault(t *testing.T) {
	// An explicitly empty env var should not overwrite the default (strings.TrimSpace("") == "")
	t.Setenv("ADMIN_TOKEN", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AdminToken == "" {
		t.Error("AdminToken is empty; empty env var should leave the default intact")
	}
}
