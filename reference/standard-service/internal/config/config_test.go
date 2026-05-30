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

func TestValidateFailsOnNegativeMaxBodyBytes(t *testing.T) {
	cfg := Defaults()
	cfg.App.MaxBodyBytes = -1
	if err := Validate(cfg); err == nil {
		t.Fatal("Validate: want error for negative MaxBodyBytes, got nil")
	}
}

func TestValidateAllowsZeroMaxBodyBytes(t *testing.T) {
	cfg := Defaults()
	cfg.App.MaxBodyBytes = 0
	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate: want nil for MaxBodyBytes=0 (disables limit), got %v", err)
	}
}

func TestValidateTLSRequiresCertAndKey(t *testing.T) {
	t.Run("TLS enabled with no cert returns error", func(t *testing.T) {
		cfg := Defaults()
		cfg.Core.TLS.Enabled = true
		if err := Validate(cfg); err == nil {
			t.Fatal("Validate: want error when TLS enabled without cert, got nil")
		}
	})

	t.Run("TLS enabled with cert but no key returns error", func(t *testing.T) {
		cfg := Defaults()
		cfg.Core.TLS.Enabled = true
		cfg.Core.TLS.CertFile = "/path/to/cert.pem"
		if err := Validate(cfg); err == nil {
			t.Fatal("Validate: want error when TLS enabled without key, got nil")
		}
	})

	t.Run("TLS enabled with cert and key passes validation", func(t *testing.T) {
		cfg := Defaults()
		cfg.Core.TLS.Enabled = true
		cfg.Core.TLS.CertFile = "/path/to/cert.pem"
		cfg.Core.TLS.KeyFile = "/path/to/key.pem"
		if err := Validate(cfg); err != nil {
			t.Fatalf("Validate: want nil for valid TLS config, got %v", err)
		}
	})

	t.Run("TLS disabled ignores missing cert and key", func(t *testing.T) {
		cfg := Defaults()
		cfg.Core.TLS.Enabled = false
		if err := Validate(cfg); err != nil {
			t.Fatalf("Validate: want nil when TLS disabled, got %v", err)
		}
	})
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

func TestLoadEnvTLSFields(t *testing.T) {
	cfg, err := load(
		[]string{"standard-service"},
		mapLookup(map[string]string{
			"APP_TLS_CERT_FILE": "/certs/server.crt",
			"APP_TLS_KEY_FILE":  "/certs/server.key",
		}),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Core.TLS.CertFile != "/certs/server.crt" {
		t.Fatalf("TLS.CertFile = %q, want %q", cfg.Core.TLS.CertFile, "/certs/server.crt")
	}
	if cfg.Core.TLS.KeyFile != "/certs/server.key" {
		t.Fatalf("TLS.KeyFile = %q, want %q", cfg.Core.TLS.KeyFile, "/certs/server.key")
	}
}

func TestLoadEnvTLSEnabled(t *testing.T) {
	t.Run("APP_TLS_ENABLED=true sets TLS.Enabled", func(t *testing.T) {
		cfg, err := load(
			[]string{"standard-service"},
			mapLookup(map[string]string{
				"APP_TLS_ENABLED":   "true",
				"APP_TLS_CERT_FILE": "/c.pem",
				"APP_TLS_KEY_FILE":  "/k.pem",
			}),
		)
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		if !cfg.Core.TLS.Enabled {
			t.Fatal("TLS.Enabled = false, want true")
		}
	})

	t.Run("APP_TLS_ENABLED=false leaves TLS disabled", func(t *testing.T) {
		cfg, err := load(
			[]string{"standard-service"},
			mapLookup(map[string]string{"APP_TLS_ENABLED": "false"}),
		)
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		if cfg.Core.TLS.Enabled {
			t.Fatal("TLS.Enabled = true, want false")
		}
	})
}

// TestWriteKeyClearableByEmptyEnv verifies that APP_WRITE_KEY="" in the process
// environment overrides a non-empty WriteKey from the .env file.
// This matters because an empty WriteKey disables the write guard entirely; if
// the str helper's empty-skip logic were applied here, a developer could not
// disable the guard via environment once .env set it.
func TestWriteKeyClearableByEmptyEnv(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("APP_WRITE_KEY=secret-from-file\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := load(
		[]string{"standard-service", "--env-file", envFile},
		mapLookup(map[string]string{"APP_WRITE_KEY": ""}),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.App.WriteKey != "" {
		t.Fatalf("WriteKey = %q, want empty: process env APP_WRITE_KEY= must override .env value", cfg.App.WriteKey)
	}
}

func TestLoadEnvCORSAllowedOrigins(t *testing.T) {
	t.Run("comma-separated origins are split and trimmed", func(t *testing.T) {
		cfg, err := load(
			[]string{"standard-service"},
			mapLookup(map[string]string{
				"APP_CORS_ALLOWED_ORIGINS": "https://app.example.com, https://admin.example.com",
			}),
		)
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		if len(cfg.App.CORSAllowedOrigins) != 2 {
			t.Fatalf("CORSAllowedOrigins = %v, want 2 entries", cfg.App.CORSAllowedOrigins)
		}
		if cfg.App.CORSAllowedOrigins[0] != "https://app.example.com" {
			t.Fatalf("CORSAllowedOrigins[0] = %q, want %q", cfg.App.CORSAllowedOrigins[0], "https://app.example.com")
		}
		if cfg.App.CORSAllowedOrigins[1] != "https://admin.example.com" {
			t.Fatalf("CORSAllowedOrigins[1] = %q, want %q", cfg.App.CORSAllowedOrigins[1], "https://admin.example.com")
		}
	})

	t.Run("empty value leaves CORSAllowedOrigins nil (allow all default)", func(t *testing.T) {
		cfg, err := load(
			[]string{"standard-service"},
			mapLookup(map[string]string{"APP_CORS_ALLOWED_ORIGINS": ""}),
		)
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		if cfg.App.CORSAllowedOrigins != nil {
			t.Fatalf("CORSAllowedOrigins = %v, want nil when env is empty", cfg.App.CORSAllowedOrigins)
		}
	})

	t.Run("absent variable leaves CORSAllowedOrigins nil (allow all default)", func(t *testing.T) {
		cfg, err := load([]string{"standard-service"}, mapLookup(nil))
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		if cfg.App.CORSAllowedOrigins != nil {
			t.Fatalf("CORSAllowedOrigins = %v, want nil when var is absent", cfg.App.CORSAllowedOrigins)
		}
	})

	t.Run("single origin without comma", func(t *testing.T) {
		cfg, err := load(
			[]string{"standard-service"},
			mapLookup(map[string]string{"APP_CORS_ALLOWED_ORIGINS": "https://app.example.com"}),
		)
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		if len(cfg.App.CORSAllowedOrigins) != 1 || cfg.App.CORSAllowedOrigins[0] != "https://app.example.com" {
			t.Fatalf("CORSAllowedOrigins = %v, want [https://app.example.com]", cfg.App.CORSAllowedOrigins)
		}
	})
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
