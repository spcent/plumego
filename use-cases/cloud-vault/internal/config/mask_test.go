package config

import (
	"testing"
)

func TestMasked(t *testing.T) {
	cfg := Defaults()
	cfg.Qiniu.SecretKey = "my-secret-key"
	cfg.AI.APIKey = "sk-123456"
	cfg.Auth.BootstrapAdminPassword = "admin-pass"

	masked := cfg.Masked()

	if masked.Qiniu.SecretKey != "[REDACTED]" {
		t.Errorf("Masked().Qiniu.SecretKey = %q, want [REDACTED]", masked.Qiniu.SecretKey)
	}
	if masked.AI.APIKey != "[REDACTED]" {
		t.Errorf("Masked().AI.APIKey = %q, want [REDACTED]", masked.AI.APIKey)
	}
	if masked.Auth.BootstrapAdminPassword != "[REDACTED]" {
		t.Errorf("Masked().Auth.BootstrapAdminPassword = %q, want [REDACTED]", masked.Auth.BootstrapAdminPassword)
	}

	// Original unchanged
	if cfg.Qiniu.SecretKey != "my-secret-key" {
		t.Errorf("Original Qiniu.SecretKey modified")
	}
}

func TestMasked_EmptyFields(t *testing.T) {
	cfg := Defaults()
	// All sensitive fields empty
	masked := cfg.Masked()

	if masked.Qiniu.SecretKey != "" {
		t.Errorf("Masked() should leave empty fields empty, got %q", masked.Qiniu.SecretKey)
	}
	if masked.AI.APIKey != "" {
		t.Errorf("Masked() should leave empty fields empty, got %q", masked.AI.APIKey)
	}
}

func TestHasSecrets(t *testing.T) {
	cfg := Defaults()
	if cfg.HasSecrets() {
		t.Error("HasSecrets() should be false for defaults")
	}

	cfg.Qiniu.SecretKey = "sk"
	if !cfg.HasSecrets() {
		t.Error("HasSecrets() should be true when Qiniu.SecretKey set")
	}

	cfg = Defaults()
	cfg.AI.APIKey = "key"
	if !cfg.HasSecrets() {
		t.Error("HasSecrets() should be true when AI.APIKey set")
	}

	cfg = Defaults()
	cfg.Auth.BootstrapAdminPassword = "pass"
	if !cfg.HasSecrets() {
		t.Error("HasSecrets() should be true when Auth.BootstrapAdminPassword set")
	}
}
