package config

import (
	"testing"
	"time"
)

func TestValidateRequiresAddr(t *testing.T) {
	cfg := Defaults()
	cfg.Core.Addr = ""
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for empty addr")
	}
}

func TestValidateRequiresServiceName(t *testing.T) {
	cfg := Defaults()
	cfg.App.ServiceName = ""
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestValidateRequiresEnvironment(t *testing.T) {
	cfg := Defaults()
	cfg.App.Environment = ""
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for empty environment")
	}
}

func TestValidateBodyLimitMustBePositive(t *testing.T) {
	cfg := Defaults()
	cfg.App.BodyLimitBytes = 0
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for non-positive body limit")
	}
	cfg.App.BodyLimitBytes = -1
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for negative body limit")
	}
}

func TestValidateRequestTimeoutMustBePositive(t *testing.T) {
	cfg := Defaults()
	cfg.App.RequestTimeout = 0
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for zero request timeout")
	}
}

func TestValidateRateLimitMustBePositive(t *testing.T) {
	cfg := Defaults()
	cfg.App.RateLimit = 0
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for zero rate limit")
	}
}

func TestValidateRateBurstMustBePositive(t *testing.T) {
	cfg := Defaults()
	cfg.App.RateBurst = 0
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for zero rate burst")
	}
}

func TestValidateTLSRequiresCertAndKey(t *testing.T) {
	t.Run("enabled with no cert returns error", func(t *testing.T) {
		cfg := Defaults()
		cfg.Core.TLS.Enabled = true
		cfg.Core.TLS.CertFile = ""
		cfg.Core.TLS.KeyFile = "key.pem"
		if err := Validate(cfg); err == nil {
			t.Fatal("expected error: TLS enabled but cert missing")
		}
	})

	t.Run("enabled with no key returns error", func(t *testing.T) {
		cfg := Defaults()
		cfg.Core.TLS.Enabled = true
		cfg.Core.TLS.CertFile = "cert.pem"
		cfg.Core.TLS.KeyFile = ""
		if err := Validate(cfg); err == nil {
			t.Fatal("expected error: TLS enabled but key missing")
		}
	})

	t.Run("enabled with both cert and key is valid", func(t *testing.T) {
		cfg := Defaults()
		cfg.Core.TLS.Enabled = true
		cfg.Core.TLS.CertFile = "cert.pem"
		cfg.Core.TLS.KeyFile = "key.pem"
		if err := Validate(cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("disabled ignores missing cert and key", func(t *testing.T) {
		cfg := Defaults()
		cfg.Core.TLS.Enabled = false
		if err := Validate(cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDefaultsAreValid(t *testing.T) {
	if err := Validate(Defaults()); err != nil {
		t.Fatalf("Defaults() does not pass Validate: %v", err)
	}
}

func TestLoadEnvOverridesDefaults(t *testing.T) {
	env := map[string]string{
		"APP_ADDR":             ":9090",
		"APP_ENV":              "staging",
		"APP_SERVICE_NAME":     "my-service",
		"APP_API_TOKEN":        "tok-abc",
		"OPS_TOKEN":            "ops-xyz",
		"APP_RATE_LIMIT":       "50",
		"APP_RATE_BURST":       "100",
		"APP_REQUEST_TIMEOUT":  "10s",
		"APP_BODY_LIMIT_BYTES": "2097152",
	}
	cfg, err := load(nil, func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Core.Addr != ":9090" {
		t.Errorf("Addr = %q, want :9090", cfg.Core.Addr)
	}
	if cfg.App.Environment != "staging" {
		t.Errorf("Environment = %q, want staging", cfg.App.Environment)
	}
	if cfg.App.ServiceName != "my-service" {
		t.Errorf("ServiceName = %q, want my-service", cfg.App.ServiceName)
	}
	if cfg.App.APIToken != "tok-abc" {
		t.Errorf("APIToken = %q, want tok-abc", cfg.App.APIToken)
	}
	if cfg.App.OpsToken != "ops-xyz" {
		t.Errorf("OpsToken = %q, want ops-xyz", cfg.App.OpsToken)
	}
	if cfg.App.RateLimit != 50 {
		t.Errorf("RateLimit = %v, want 50", cfg.App.RateLimit)
	}
	if cfg.App.RateBurst != 100 {
		t.Errorf("RateBurst = %d, want 100", cfg.App.RateBurst)
	}
	if cfg.App.RequestTimeout != 10*time.Second {
		t.Errorf("RequestTimeout = %v, want 10s", cfg.App.RequestTimeout)
	}
	if cfg.App.BodyLimitBytes != 2097152 {
		t.Errorf("BodyLimitBytes = %d, want 2097152", cfg.App.BodyLimitBytes)
	}
}

func TestLoadEnvTLSFields(t *testing.T) {
	env := map[string]string{
		"APP_TLS_ENABLED":   "true",
		"APP_TLS_CERT_FILE": "/etc/tls/cert.pem",
		"APP_TLS_KEY_FILE":  "/etc/tls/key.pem",
	}
	cfg, err := load(nil, func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !cfg.Core.TLS.Enabled {
		t.Error("TLS.Enabled = false, want true")
	}
	if cfg.Core.TLS.CertFile != "/etc/tls/cert.pem" {
		t.Errorf("TLS.CertFile = %q, want /etc/tls/cert.pem", cfg.Core.TLS.CertFile)
	}
	if cfg.Core.TLS.KeyFile != "/etc/tls/key.pem" {
		t.Errorf("TLS.KeyFile = %q, want /etc/tls/key.pem", cfg.Core.TLS.KeyFile)
	}
}

func TestLoadEnvIgnoresMalformedNumerics(t *testing.T) {
	// Malformed values must not corrupt the config; default falls through.
	env := map[string]string{
		"APP_RATE_LIMIT":       "not-a-number",
		"APP_RATE_BURST":       "not-a-number",
		"APP_BODY_LIMIT_BYTES": "not-a-number",
	}
	defaults := Defaults()
	cfg, err := load(nil, func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.App.RateLimit != defaults.App.RateLimit {
		t.Errorf("RateLimit = %v, want default %v", cfg.App.RateLimit, defaults.App.RateLimit)
	}
	if cfg.App.RateBurst != defaults.App.RateBurst {
		t.Errorf("RateBurst = %d, want default %d", cfg.App.RateBurst, defaults.App.RateBurst)
	}
	if cfg.App.BodyLimitBytes != defaults.App.BodyLimitBytes {
		t.Errorf("BodyLimitBytes = %d, want default %d", cfg.App.BodyLimitBytes, defaults.App.BodyLimitBytes)
	}
}

func TestLoadEnvEmptyKeySkipped(t *testing.T) {
	// An env var that is present but blank should not overwrite the default string.
	env := map[string]string{
		"APP_ENV":          "  ",
		"APP_SERVICE_NAME": "",
	}
	defaults := Defaults()
	cfg, err := load(nil, func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.App.Environment != defaults.App.Environment {
		t.Errorf("Environment = %q, want default %q", cfg.App.Environment, defaults.App.Environment)
	}
	if cfg.App.ServiceName != defaults.App.ServiceName {
		t.Errorf("ServiceName = %q, want default %q", cfg.App.ServiceName, defaults.App.ServiceName)
	}
}
