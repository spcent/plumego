package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrecedenceDefaultsEnvFileEnvironmentFlags(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_ADDR=:7000\nAPP_ENV_FILE=ignored-from-file\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := load(
		[]string{"standard-service", "--env-file", envFile, "--addr", ":9000"},
		mapLookup(map[string]string{
			"APP_ADDR": ":8000",
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
}

func TestLoadEnvironmentOverridesEnvFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_ADDR=:7000\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := load(
		[]string{"standard-service", "--env-file=" + envFile},
		mapLookup(map[string]string{
			"APP_ADDR": ":8000",
		}),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Core.Addr != ":8000" {
		t.Fatalf("addr = %q, want %q", cfg.Core.Addr, ":8000")
	}
}

func TestLoadIgnoresUnrelatedFlags(t *testing.T) {
	cfg, err := load(
		[]string{"standard-service", "--output", "openapi.json", "--format=json", "--addr", ":9000"},
		mapLookup(nil),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Core.Addr != ":9000" {
		t.Fatalf("addr = %q, want %q", cfg.Core.Addr, ":9000")
	}
}

func TestValidateFailsOnEmptyAddr(t *testing.T) {
	cfg := Defaults()
	cfg.Core.Addr = ""
	if err := Validate(cfg); err == nil {
		t.Fatal("Validate: want error for empty addr, got nil")
	}
}

func TestLoadEnvWriteKey(t *testing.T) {
	cfg, err := load(
		[]string{"standard-service"},
		mapLookup(map[string]string{"APP_WRITE_KEY": "mysecret"}),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.App.WriteKey != "mysecret" {
		t.Fatalf("WriteKey = %q, want %q", cfg.App.WriteKey, "mysecret")
	}
}

func TestLoadEnvServiceName(t *testing.T) {
	cfg, err := load(
		[]string{"standard-service"},
		mapLookup(map[string]string{"APP_SERVICE_NAME": "my-service"}),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.App.ServiceName != "my-service" {
		t.Fatalf("ServiceName = %q, want %q", cfg.App.ServiceName, "my-service")
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
