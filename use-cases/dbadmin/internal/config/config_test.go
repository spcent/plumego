package config

import (
	"strings"
	"testing"
)

func TestValidateRequiresExplicitPassword(t *testing.T) {
	cfg := Defaults()
	err := Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "DBADMIN_PASSWORD is required") {
		t.Fatalf("Validate error = %v, want password required", err)
	}
}

func TestValidateRejectsDefaultCredentialsUnlessAllowed(t *testing.T) {
	cfg := Defaults()
	cfg.App.AdminPassword = "admin"
	err := Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "refusing default admin/admin") {
		t.Fatalf("Validate error = %v, want default credential refusal", err)
	}

	cfg.App.AllowDefaultPassword = true
	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate with explicit demo override error = %v", err)
	}
}

func TestValidateRejectsDefaultPasswordOnPublicAddress(t *testing.T) {
	cfg := Defaults()
	cfg.Core.Addr = ":8080"
	cfg.App.AdminUser = "root"
	cfg.App.AdminPassword = "admin"

	err := Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "non-loopback") {
		t.Fatalf("Validate error = %v, want public address default password refusal", err)
	}
}

func TestLoadReadsAllowDefaultPassword(t *testing.T) {
	cfg, err := load([]string{"dbadmin", "--env-file", t.TempDir() + "/missing.env"}, func(key string) (string, bool) {
		values := map[string]string{
			"DBADMIN_PASSWORD":               "admin",
			"DBADMIN_ALLOW_DEFAULT_PASSWORD": "true",
			"DBADMIN_AUDIT_RETENTION_DAYS":   "30",
			"DBADMIN_AUDIT_MAX_EVENTS":       "250",
		}
		v, ok := values[key]
		return v, ok
	})
	if err != nil {
		t.Fatalf("load error = %v", err)
	}
	if !cfg.App.AllowDefaultPassword {
		t.Fatal("AllowDefaultPassword = false, want true")
	}
	if cfg.App.AuditRetentionDays != 30 {
		t.Fatalf("AuditRetentionDays = %d, want 30", cfg.App.AuditRetentionDays)
	}
	if cfg.App.AuditMaxEvents != 250 {
		t.Fatalf("AuditMaxEvents = %d, want 250", cfg.App.AuditMaxEvents)
	}
}

func TestValidateRejectsInvalidAuditRetention(t *testing.T) {
	cfg := Defaults()
	cfg.App.AdminPassword = "secret"
	cfg.App.AuditRetentionDays = 0
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "DBADMIN_AUDIT_RETENTION_DAYS") {
		t.Fatalf("Validate error = %v, want audit retention error", err)
	}
	cfg.App.AuditRetentionDays = 90
	cfg.App.AuditMaxEvents = 0
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "DBADMIN_AUDIT_MAX_EVENTS") {
		t.Fatalf("Validate error = %v, want audit max events error", err)
	}
}
