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

func TestDefaultsAuditMaxEventsIs10000(t *testing.T) {
	cfg := Defaults()
	if cfg.App.AuditMaxEvents != 10000 {
		t.Fatalf("AuditMaxEvents default = %d, want 10000", cfg.App.AuditMaxEvents)
	}
}

func TestLoadParsesUsersEnvVar(t *testing.T) {
	cfg, err := load([]string{"dbadmin", "--env-file", t.TempDir() + "/missing.env"}, func(key string) (string, bool) {
		values := map[string]string{
			"DBADMIN_PASSWORD": "secret",
			"DBADMIN_USERS":    `[{"Username":"alice","Password":"alicepw","Role":"admin"},{"Username":"bob","Password":"bobpw","Role":"readonly"}]`,
		}
		v, ok := values[key]
		return v, ok
	})
	if err != nil {
		t.Fatalf("load error = %v", err)
	}
	if len(cfg.App.Users) != 2 {
		t.Fatalf("Users length = %d, want 2", len(cfg.App.Users))
	}
	if cfg.App.Users[0].Username != "alice" || cfg.App.Users[0].Role != "admin" {
		t.Fatalf("Users[0] = %+v, want alice/admin", cfg.App.Users[0])
	}
	if cfg.App.Users[1].Username != "bob" || cfg.App.Users[1].Role != "readonly" {
		t.Fatalf("Users[1] = %+v, want bob/readonly", cfg.App.Users[1])
	}
}

func TestLoadParsesAllowedIPsEnvVar(t *testing.T) {
	cfg, err := load([]string{"dbadmin", "--env-file", t.TempDir() + "/missing.env"}, func(key string) (string, bool) {
		values := map[string]string{
			"DBADMIN_PASSWORD":    "secret",
			"DBADMIN_ALLOWED_IPS": "192.168.1.0/24, 10.0.0.1 ,2001:db8::1",
		}
		v, ok := values[key]
		return v, ok
	})
	if err != nil {
		t.Fatalf("load error = %v", err)
	}
	want := []string{"192.168.1.0/24", "10.0.0.1", "2001:db8::1"}
	if len(cfg.App.AllowedIPs) != len(want) {
		t.Fatalf("AllowedIPs = %v, want %v", cfg.App.AllowedIPs, want)
	}
	for i, ip := range want {
		if cfg.App.AllowedIPs[i] != ip {
			t.Fatalf("AllowedIPs[%d] = %q, want %q", i, cfg.App.AllowedIPs[i], ip)
		}
	}
}

func TestResolveUsersFallsBackToSingleUser(t *testing.T) {
	cfg := Defaults()
	cfg.App.AdminUser = "admin"
	cfg.App.AdminPassword = "secret"
	cfg.App.AdminRole = "admin"

	users := cfg.App.ResolveUsers()
	if len(users) != 1 {
		t.Fatalf("ResolveUsers length = %d, want 1", len(users))
	}
	if users[0].Username != "admin" || users[0].Password != "secret" || users[0].Role != "admin" {
		t.Fatalf("ResolveUsers()[0] = %+v, want admin/secret/admin", users[0])
	}
}

func TestResolveUsersReturnsConfiguredUsers(t *testing.T) {
	cfg := Defaults()
	cfg.App.Users = []UserConfig{
		{Username: "alice", Password: "alicepw", Role: "admin"},
		{Username: "bob", Password: "bobpw", Role: "readonly"},
	}

	users := cfg.App.ResolveUsers()
	if len(users) != 2 {
		t.Fatalf("ResolveUsers length = %d, want 2", len(users))
	}
	if users[0].Username != "alice" || users[1].Username != "bob" {
		t.Fatalf("ResolveUsers() = %+v, want alice then bob", users)
	}
}

func TestValidateRejectsInvalidRoleInUsers(t *testing.T) {
	cfg := Defaults()
	cfg.App.Users = []UserConfig{
		{Username: "alice", Password: "alicepw", Role: "superuser"},
	}
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "role must be admin or readonly") {
		t.Fatalf("Validate error = %v, want invalid role error", err)
	}
}

func TestValidateRejectsEmptyUsernameOrPasswordInUsers(t *testing.T) {
	cfg := Defaults()
	cfg.App.Users = []UserConfig{{Username: "", Password: "pw", Role: "admin"}}
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "username is required") {
		t.Fatalf("Validate error = %v, want username required error", err)
	}

	cfg.App.Users = []UserConfig{{Username: "alice", Password: "", Role: "admin"}}
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "password is required") {
		t.Fatalf("Validate error = %v, want password required error", err)
	}
}

func TestValidateRejectsInvalidAllowedIPs(t *testing.T) {
	cfg := Defaults()
	cfg.App.AdminPassword = "secret"
	cfg.App.AllowedIPs = []string{"not-an-ip"}
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "DBADMIN_ALLOWED_IPS") {
		t.Fatalf("Validate error = %v, want allowed IPs error", err)
	}
}

func TestValidateAcceptsValidAllowedIPs(t *testing.T) {
	cfg := Defaults()
	cfg.App.AdminPassword = "secret"
	cfg.App.AllowedIPs = []string{"192.168.1.0/24", "10.0.0.1", "::1"}
	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate error = %v, want no error", err)
	}
}
