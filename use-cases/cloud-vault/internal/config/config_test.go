package config

import (
	"os"
	"testing"

	"github.com/spcent/plumego/core"
)

func TestTOMLConfig(t *testing.T) {
	// Create a temporary TOML file
	content := `
[server]
addr = ":9090"

[app]
max_upload_size_mb = 25
version_policy = "limited"

[database]
path = "./data/test.db"

[storage]
provider = "local"

[storage.local]
root = "./data/test-objects"

[auth]
enabled = true
session_ttl_hours = 48
password_min_length = 14

[auth.bootstrap_admin]
enabled = true
username = "testadmin"
email = "test@example.com"
`
	tmpfile, err := os.CreateTemp("", "test-config-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Test loading
	cfg, err := load([]string{"cmd", "--config", tmpfile.Name()}, nil)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Verify values
	if cfg.Core.Addr != ":9090" {
		t.Errorf("expected addr :9090, got %s", cfg.Core.Addr)
	}
	if cfg.App.MaxUploadSizeMB != 25 {
		t.Errorf("expected max_upload_size_mb 25, got %d", cfg.App.MaxUploadSizeMB)
	}
	if cfg.App.VersionPolicy != "limited" {
		t.Errorf("expected version_policy limited, got %s", cfg.App.VersionPolicy)
	}
	if cfg.DB.Path != "./data/test.db" {
		t.Errorf("expected db path ./data/test.db, got %s", cfg.DB.Path)
	}
	if cfg.Local.Root != "./data/test-objects" {
		t.Errorf("expected local root ./data/test-objects, got %s", cfg.Local.Root)
	}
	if !cfg.Auth.Enabled {
		t.Error("expected auth.enabled = true")
	}
	if cfg.Auth.SessionTTLHours != 48 {
		t.Errorf("expected auth.session_ttl_hours 48, got %d", cfg.Auth.SessionTTLHours)
	}
	if cfg.Auth.PasswordMinLength != 14 {
		t.Errorf("expected auth.password_min_length 14, got %d", cfg.Auth.PasswordMinLength)
	}
	if !cfg.Auth.BootstrapAdminEnabled {
		t.Error("expected auth.bootstrap_admin.enabled = true")
	}
	if cfg.Auth.BootstrapAdminUsername != "testadmin" {
		t.Errorf("expected bootstrap_admin.username testadmin, got %s", cfg.Auth.BootstrapAdminUsername)
	}
	if cfg.Auth.BootstrapAdminEmail != "test@example.com" {
		t.Errorf("expected bootstrap_admin.email test@example.com, got %s", cfg.Auth.BootstrapAdminEmail)
	}
}

func TestEnvOverride(t *testing.T) {
	// Create a temporary TOML file
	content := `
[server]
addr = ":9090"

[auth]
enabled = false
session_ttl_hours = 24
`
	tmpfile, err := os.CreateTemp("", "test-config-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Mock environment variables
	env := map[string]string{
		"APP_ADDR":              ":7070",
		"AUTH_ENABLED":          "true",
		"AUTH_SESSION_TTL_HOURS": "72",
	}

	mockLookup := func(key string) (string, bool) {
		val, ok := env[key]
		return val, ok
	}

	// Test loading with env overrides
	cfg, err := load([]string{"cmd", "--config", tmpfile.Name()}, mockLookup)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Verify env overrides TOML
	if cfg.Core.Addr != ":7070" {
		t.Errorf("expected addr :7070 from env, got %s", cfg.Core.Addr)
	}
	if !cfg.Auth.Enabled {
		t.Error("expected auth.enabled = true from env")
	}
	if cfg.Auth.SessionTTLHours != 72 {
		t.Errorf("expected auth.session_ttl_hours 72 from env, got %d", cfg.Auth.SessionTTLHours)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				Core:    coreConfig(":8080"),
				Storage: storageConfig("local"),
				Auth:    authConfig(false, 24, 12),
			},
			wantErr: false,
		},
		{
			name: "missing addr",
			cfg: Config{
				Core:    coreConfig(""),
				Storage: storageConfig("local"),
			},
			wantErr: true,
		},
		{
			name: "invalid storage provider",
			cfg: Config{
				Core:    coreConfig(":8080"),
				Storage: storageConfig("s3"),
			},
			wantErr: true,
		},
		{
			name: "auth enabled with invalid session ttl",
			cfg: Config{
				Core:    coreConfig(":8080"),
				Storage: storageConfig("local"),
				Auth:    authConfig(true, 0, 12),
			},
			wantErr: true,
		},
		{
			name: "auth enabled with invalid password length",
			cfg: Config{
				Core:    coreConfig(":8080"),
				Storage: storageConfig("local"),
				Auth:    authConfig(true, 24, 6),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions for test cases
func coreConfig(addr string) core.AppConfig {
	cfg := core.DefaultConfig()
	cfg.Addr = addr
	return cfg
}

func storageConfig(provider string) StorageConfig {
	return StorageConfig{Provider: provider}
}

func authConfig(enabled bool, sessionTTL, pwdMinLen int) AuthConfig {
	return AuthConfig{
		Enabled:           enabled,
		SessionTTLHours:   sessionTTL,
		PasswordMinLength: pwdMinLen,
	}
}
