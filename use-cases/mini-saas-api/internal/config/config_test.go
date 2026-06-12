package config

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultsPassValidation(t *testing.T) {
	if err := Validate(Defaults()); err != nil {
		t.Fatalf("Defaults() failed validation: %v", err)
	}
}

func TestValidateRejectsEmptyAddr(t *testing.T) {
	cfg := Defaults()
	cfg.Core.Addr = ""
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for empty addr, got nil")
	}
}

func TestValidateRejectsShortJWTSecret(t *testing.T) {
	cfg := Defaults()
	cfg.App.JWTSecret = "tooshort"
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for short JWT secret, got nil")
	}
	if !strings.Contains(err.Error(), "APP_JWT_SECRET") {
		t.Fatalf("error should mention APP_JWT_SECRET, got: %v", err)
	}
}

func TestValidateRejectsEmptyJWTSecret(t *testing.T) {
	cfg := Defaults()
	cfg.App.JWTSecret = ""
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for empty JWT secret, got nil")
	}
}

func TestValidateRejectsNegativeBodyLimit(t *testing.T) {
	cfg := Defaults()
	cfg.App.MaxBodyBytes = -1
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for negative MaxBodyBytes, got nil")
	}
}

func TestValidateAcceptsZeroBodyLimit(t *testing.T) {
	cfg := Defaults()
	cfg.App.MaxBodyBytes = 0 // disables limit
	if err := Validate(cfg); err != nil {
		t.Fatalf("zero body limit should be valid: %v", err)
	}
}

func TestValidateRejectsZeroAccessTTL(t *testing.T) {
	cfg := Defaults()
	cfg.App.JWTAccessTTL = 0
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for zero JWTAccessTTL, got nil")
	}
}

func TestValidateRejectsZeroRefreshTTL(t *testing.T) {
	cfg := Defaults()
	cfg.App.JWTRefreshTTL = 0
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for zero JWTRefreshTTL, got nil")
	}
}

func TestValidateTLSRequiresCertAndKey(t *testing.T) {
	cfg := Defaults()
	cfg.Core.TLS.Enabled = true
	cfg.Core.TLS.CertFile = ""
	cfg.Core.TLS.KeyFile = ""
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error when TLS enabled without cert/key, got nil")
	}
}

func TestLoadFromEnvMap(t *testing.T) {
	secret := strings.Repeat("x", 32)
	cfg, err := load([]string{"app"}, func(key string) (string, bool) {
		switch key {
		case "APP_JWT_SECRET":
			return secret, true
		case "APP_JWT_ACCESS_TTL":
			return "5m", true
		case "APP_SERVICE_NAME":
			return "test-svc", true
		}
		return "", false
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.App.JWTSecret != secret {
		t.Fatalf("JWTSecret = %q, want %q", cfg.App.JWTSecret, secret)
	}
	if cfg.App.JWTAccessTTL != 5*time.Minute {
		t.Fatalf("JWTAccessTTL = %v, want 5m", cfg.App.JWTAccessTTL)
	}
	if cfg.App.ServiceName != "test-svc" {
		t.Fatalf("ServiceName = %q, want test-svc", cfg.App.ServiceName)
	}
}

func TestCORSOriginsFromEnv(t *testing.T) {
	cfg, err := load([]string{"app"}, func(key string) (string, bool) {
		if key == "APP_CORS_ALLOWED_ORIGINS" {
			return "https://a.com, https://b.com", true
		}
		return "", false
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.App.CORSAllowedOrigins) != 2 {
		t.Fatalf("CORSAllowedOrigins = %v, want 2 entries", cfg.App.CORSAllowedOrigins)
	}
}
