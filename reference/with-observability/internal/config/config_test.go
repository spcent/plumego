package config

import (
	"os"
	"path/filepath"
	"testing"
)

func mapLookup(m map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestLoadCORSAllowedOriginsCommaSeparated(t *testing.T) {
	cfg, err := load([]string{"with-observability"}, mapLookup(map[string]string{
		"APP_CORS_ALLOWED_ORIGINS": "https://app.example.com, https://admin.example.com",
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.App.CORSAllowedOrigins) != 2 {
		t.Fatalf("CORSAllowedOrigins = %v, want 2 entries", cfg.App.CORSAllowedOrigins)
	}
	if cfg.App.CORSAllowedOrigins[0] != "https://app.example.com" {
		t.Errorf("CORSAllowedOrigins[0] = %q, want %q", cfg.App.CORSAllowedOrigins[0], "https://app.example.com")
	}
	if cfg.App.CORSAllowedOrigins[1] != "https://admin.example.com" {
		t.Errorf("CORSAllowedOrigins[1] = %q, want %q", cfg.App.CORSAllowedOrigins[1], "https://admin.example.com")
	}
}

func TestLoadCORSAllowedOriginsEmptyVarLeavesNil(t *testing.T) {
	cfg, err := load([]string{"with-observability"}, mapLookup(map[string]string{
		"APP_CORS_ALLOWED_ORIGINS": "",
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.App.CORSAllowedOrigins != nil {
		t.Fatalf("CORSAllowedOrigins = %v, want nil for empty var", cfg.App.CORSAllowedOrigins)
	}
}

func TestLoadCORSAllowedOriginsAbsentVarLeavesNil(t *testing.T) {
	cfg, err := load([]string{"with-observability"}, mapLookup(nil))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.App.CORSAllowedOrigins != nil {
		t.Fatalf("CORSAllowedOrigins = %v, want nil when var absent", cfg.App.CORSAllowedOrigins)
	}
}

func TestLoadCORSAllowedOriginsSingleOrigin(t *testing.T) {
	cfg, err := load([]string{"with-observability"}, mapLookup(map[string]string{
		"APP_CORS_ALLOWED_ORIGINS": "https://app.example.com",
	}))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.App.CORSAllowedOrigins) != 1 || cfg.App.CORSAllowedOrigins[0] != "https://app.example.com" {
		t.Fatalf("CORSAllowedOrigins = %v, want [https://app.example.com]", cfg.App.CORSAllowedOrigins)
	}
}

func TestLoadEnvFileIsReadBeforeEnv(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_ADDR=:7001\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	cfg, err := load(
		[]string{"with-observability", "--env-file", envFile},
		mapLookup(map[string]string{"APP_ADDR": ":8001"}),
	)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// Process env overrides .env file.
	if cfg.Core.Addr != ":8001" {
		t.Fatalf("addr = %q, want :8001", cfg.Core.Addr)
	}
}
