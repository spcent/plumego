package config

import (
	"testing"
)

func TestDefaultsAddr(t *testing.T) {
	cfg := Defaults()
	if cfg.Core.Addr != ":8084" {
		t.Fatalf("Defaults().Core.Addr = %q, want %q", cfg.Core.Addr, ":8084")
	}
}

func TestDefaultsCORSAllowedOriginsNil(t *testing.T) {
	cfg := Defaults()
	if cfg.App.CORSAllowedOrigins != nil {
		t.Fatalf("Defaults().App.CORSAllowedOrigins = %v, want nil", cfg.App.CORSAllowedOrigins)
	}
}

func TestLoadCORSAllowedOriginsCommaSeparated(t *testing.T) {
	t.Setenv("APP_CORS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.App.CORSAllowedOrigins) != 2 {
		t.Fatalf("CORSAllowedOrigins = %v, want 2 entries", cfg.App.CORSAllowedOrigins)
	}
	if cfg.App.CORSAllowedOrigins[0] != "https://app.example.com" {
		t.Errorf("CORSAllowedOrigins[0] = %q, want https://app.example.com", cfg.App.CORSAllowedOrigins[0])
	}
}

func TestLoadCORSAllowedOriginsEmptyLeavesNil(t *testing.T) {
	t.Setenv("APP_CORS_ALLOWED_ORIGINS", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.App.CORSAllowedOrigins != nil {
		t.Fatalf("CORSAllowedOrigins = %v, want nil for empty var", cfg.App.CORSAllowedOrigins)
	}
}
