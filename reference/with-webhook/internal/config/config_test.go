package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrecedenceDefaultsEnvFileEnvironmentFlags(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_ADDR=:7000\nAPP_DEBUG=false\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := load(
		[]string{"with-webhook", "--env-file", envFile, "--addr", ":9000", "--debug=false"},
		mapLookup(map[string]string{
			"APP_ADDR":  ":8000",
			"APP_DEBUG": "true",
		}),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Core.Addr != ":9000" {
		t.Fatalf("addr = %q, want %q", cfg.Core.Addr, ":9000")
	}
	if cfg.App.EnvFile != envFile {
		t.Fatalf("env file = %q, want %q", cfg.App.EnvFile, envFile)
	}
	if cfg.App.Debug {
		t.Fatalf("debug = true, want false")
	}
}

func TestLoadEnvironmentOverridesEnvFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_ADDR=:7000\nAPP_DEBUG=false\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := load(
		[]string{"with-webhook", "--env-file=" + envFile},
		mapLookup(map[string]string{
			"APP_ADDR":  ":8000",
			"APP_DEBUG": "true",
		}),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Core.Addr != ":8000" {
		t.Fatalf("addr = %q, want %q", cfg.Core.Addr, ":8000")
	}
	if !cfg.App.Debug {
		t.Fatalf("debug = false, want true")
	}
}

func TestLoadIgnoresUnrelatedFlags(t *testing.T) {
	cfg, err := load(
		[]string{"with-webhook", "--output", "openapi.json", "--format=json", "--addr", ":9000"},
		mapLookup(nil),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Core.Addr != ":9000" {
		t.Fatalf("addr = %q, want %q", cfg.Core.Addr, ":9000")
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
