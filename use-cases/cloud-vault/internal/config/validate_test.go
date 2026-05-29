package config

import (
	"path/filepath"
	"testing"
)

func TestValidateConfig_Server(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{"valid port", ":8080", false},
		{"valid host:port", "localhost:8080", false},
		{"empty addr", "", true},
		{"invalid port", ":99999", true},
		{"non-numeric port", ":abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Defaults()
			cfg.Core.Addr = tt.addr
			cfg.Storage.Provider = "local"
			cfg.Local.Root = t.TempDir()
			err := ValidateConfig(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_Database(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", filepath.Join(tmpDir, "test.db"), false},
		{"empty path", "", true},
		{"creates parent dir", filepath.Join(tmpDir, "subdir", "test.db"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Defaults()
			cfg.DB.Path = tt.path
			cfg.Storage.Provider = "local"
			cfg.Local.Root = tmpDir
			err := ValidateConfig(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_Storage(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("local valid", func(t *testing.T) {
		cfg := Defaults()
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err != nil {
			t.Errorf("ValidateConfig() error = %v", err)
		}
	})

	t.Run("local empty root", func(t *testing.T) {
		cfg := Defaults()
		cfg.Storage.Provider = "local"
		cfg.Local.Root = ""
		if err := ValidateConfig(cfg); err == nil {
			t.Error("ValidateConfig() expected error for empty local.root")
		}
	})

	t.Run("qiniu valid", func(t *testing.T) {
		cfg := Defaults()
		cfg.Storage.Provider = "qiniu"
		cfg.Qiniu.AccessKey = "ak"
		cfg.Qiniu.SecretKey = "sk"
		cfg.Qiniu.Bucket = "bucket"
		cfg.Qiniu.Domain = "https://example.com"
		cfg.Qiniu.Region = "z0"
		// Skip local.root validation for qiniu
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err != nil {
			t.Errorf("ValidateConfig() error = %v", err)
		}
	})

	t.Run("qiniu missing fields", func(t *testing.T) {
		cfg := Defaults()
		cfg.Storage.Provider = "qiniu"
		// All fields empty
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err == nil {
			t.Error("ValidateConfig() expected error for missing qiniu fields")
		}
	})

	t.Run("qiniu invalid region", func(t *testing.T) {
		cfg := Defaults()
		cfg.Storage.Provider = "qiniu"
		cfg.Qiniu.AccessKey = "ak"
		cfg.Qiniu.SecretKey = "sk"
		cfg.Qiniu.Bucket = "bucket"
		cfg.Qiniu.Domain = "https://example.com"
		cfg.Qiniu.Region = "invalid"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err == nil {
			t.Error("ValidateConfig() expected error for invalid qiniu.region")
		}
	})

	t.Run("invalid provider", func(t *testing.T) {
		cfg := Defaults()
		cfg.Storage.Provider = "s3"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err == nil {
			t.Error("ValidateConfig() expected error for invalid provider")
		}
	})
}

func TestValidateConfig_Auth(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("auth disabled", func(t *testing.T) {
		cfg := Defaults()
		cfg.Auth.Enabled = false
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err != nil {
			t.Errorf("ValidateConfig() error = %v", err)
		}
	})

	t.Run("auth enabled valid", func(t *testing.T) {
		cfg := Defaults()
		cfg.Auth.Enabled = true
		cfg.Auth.CookieName = "session"
		cfg.Auth.SessionTTLHours = 24
		cfg.Auth.PasswordMinLength = 10
		cfg.Auth.MaxLoginFailures = 5
		cfg.Auth.LockoutMinutes = 30
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err != nil {
			t.Errorf("ValidateConfig() error = %v", err)
		}
	})

	t.Run("auth enabled missing cookie_name", func(t *testing.T) {
		cfg := Defaults()
		cfg.Auth.Enabled = true
		cfg.Auth.CookieName = ""
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err == nil {
			t.Error("ValidateConfig() expected error for missing cookie_name")
		}
	})

	t.Run("auth enabled invalid session_ttl", func(t *testing.T) {
		cfg := Defaults()
		cfg.Auth.Enabled = true
		cfg.Auth.CookieName = "session"
		cfg.Auth.SessionTTLHours = 0
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err == nil {
			t.Error("ValidateConfig() expected error for invalid session_ttl_hours")
		}
	})

	t.Run("auth enabled weak password_min_length", func(t *testing.T) {
		cfg := Defaults()
		cfg.Auth.Enabled = true
		cfg.Auth.CookieName = "session"
		cfg.Auth.PasswordMinLength = 6
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err == nil {
			t.Error("ValidateConfig() expected error for weak password_min_length")
		}
	})
}

func TestValidateConfig_AI(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("ai disabled", func(t *testing.T) {
		cfg := Defaults()
		cfg.AI.Enabled = false
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err != nil {
			t.Errorf("ValidateConfig() error = %v", err)
		}
	})

	t.Run("ai local_mock valid", func(t *testing.T) {
		cfg := Defaults()
		cfg.AI.Enabled = true
		cfg.AI.Provider = "local_mock"
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err != nil {
			t.Errorf("ValidateConfig() error = %v", err)
		}
	})

	t.Run("ai openai_compatible valid", func(t *testing.T) {
		cfg := Defaults()
		cfg.AI.Enabled = true
		cfg.AI.Provider = "openai_compatible"
		cfg.AI.BaseURL = "https://api.openai.com/v1"
		cfg.AI.APIKey = "sk-test"
		cfg.AI.Model = "gpt-4"
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err != nil {
			t.Errorf("ValidateConfig() error = %v", err)
		}
	})

	t.Run("ai openai_compatible missing api_key", func(t *testing.T) {
		cfg := Defaults()
		cfg.AI.Enabled = true
		cfg.AI.Provider = "openai_compatible"
		cfg.AI.BaseURL = "https://api.openai.com/v1"
		cfg.AI.APIKey = ""
		cfg.AI.Model = "gpt-4"
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err == nil {
			t.Error("ValidateConfig() expected error for missing ai.api_key")
		}
	})

	t.Run("ai empty provider", func(t *testing.T) {
		cfg := Defaults()
		cfg.AI.Enabled = true
		cfg.AI.Provider = ""
		cfg.Storage.Provider = "local"
		cfg.Local.Root = tmpDir
		if err := ValidateConfig(cfg); err == nil {
			t.Error("ValidateConfig() expected error for empty ai.provider")
		}
	})
}

func TestWarnings(t *testing.T) {
	cfg := Defaults()
	cfg.Auth.Enabled = true
	cfg.Auth.SecureCookie = false

	w := Warnings(cfg)
	if len(w) == 0 {
		t.Error("Warnings() expected warning for secure_cookie=false")
	}

	cfg.Auth.SecureCookie = true
	w = Warnings(cfg)
	if len(w) != 0 {
		t.Errorf("Warnings() expected no warnings, got %v", w)
	}
}
