package config

import (
	"os"
	"path/filepath"
	"testing"
)

const testWSSecret = "0123456789abcdef0123456789abcdef"

func TestLoadPrecedenceDefaultsEnvFileEnvironmentFlags(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	fileSecret := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	envSecret := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if err := os.WriteFile(envFile, []byte("APP_ADDR=:7000\nAPP_DEBUG=false\nWS_SECRET="+fileSecret+"\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := load(
		[]string{"with-websocket", "--env-file", envFile, "--addr", ":9000", "--debug=false"},
		mapLookup(map[string]string{
			"APP_ADDR":  ":8000",
			"APP_DEBUG": "true",
			"WS_SECRET": envSecret,
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
	if cfg.WSSecret != envSecret {
		t.Fatalf("ws secret = %q, want environment value", cfg.WSSecret)
	}
}

func TestLoadEnvironmentOverridesEnvFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_ADDR=:7000\nAPP_DEBUG=false\nWS_SECRET=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := load(
		[]string{"with-websocket", "--env-file=" + envFile},
		mapLookup(map[string]string{
			"APP_ADDR":  ":8000",
			"APP_DEBUG": "true",
			"WS_SECRET": testWSSecret,
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
	if cfg.WSSecret != testWSSecret {
		t.Fatalf("ws secret = %q, want environment value", cfg.WSSecret)
	}
}

func TestLoadIgnoresUnrelatedFlags(t *testing.T) {
	cfg, err := load(
		[]string{"with-websocket", "--output", "openapi.json", "--format=json", "--addr", ":9000"},
		mapLookup(map[string]string{
			"WS_SECRET": testWSSecret,
		}),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Core.Addr != ":9000" {
		t.Fatalf("addr = %q, want %q", cfg.Core.Addr, ":9000")
	}
}

func TestLoadRequiresWSSecret(t *testing.T) {
	_, err := load([]string{"with-websocket"}, mapLookup(nil))
	if err == nil {
		t.Fatalf("load config succeeded without WS_SECRET")
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
