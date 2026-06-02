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
}
