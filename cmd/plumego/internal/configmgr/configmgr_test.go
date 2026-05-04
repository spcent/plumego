package configmgr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvFileSkipsCommentsAndBlankLines(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".env")
	content := "\n# comment\nAPP_ADDR=:8080\n WS_SECRET = secret \n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	vars, err := ParseEnvFile(path)
	if err != nil {
		t.Fatalf("parse env file: %v", err)
	}
	if vars["APP_ADDR"] != ":8080" {
		t.Fatalf("APP_ADDR = %q, want :8080", vars["APP_ADDR"])
	}
	if vars["WS_SECRET"] != "secret" {
		t.Fatalf("WS_SECRET = %q, want secret", vars["WS_SECRET"])
	}
}

func TestParseEnvFileSupportsCommonDotenvForms(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".env")
	content := `
export APP_ADDR=":8080" # local port
JWT_SECRET='quoted secret'
WS_SECRET=abc#not-comment
APP_DEBUG=true # inline comment
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	vars, err := ParseEnvFile(path)
	if err != nil {
		t.Fatalf("parse env file: %v", err)
	}
	if vars["APP_ADDR"] != ":8080" {
		t.Fatalf("APP_ADDR = %q, want :8080", vars["APP_ADDR"])
	}
	if vars["JWT_SECRET"] != "quoted secret" {
		t.Fatalf("JWT_SECRET = %q, want quoted secret", vars["JWT_SECRET"])
	}
	if vars["WS_SECRET"] != "abc#not-comment" {
		t.Fatalf("WS_SECRET = %q, want abc#not-comment", vars["WS_SECRET"])
	}
	if vars["APP_DEBUG"] != "true" {
		t.Fatalf("APP_DEBUG = %q, want true", vars["APP_DEBUG"])
	}
}

func TestParseEnvFileRejectsInvalidLine(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".env")
	if err := os.WriteFile(path, []byte("INVALID\n"), 0644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	if _, err := ParseEnvFile(path); err == nil {
		t.Fatal("expected invalid env line error")
	}
}

func TestValidateConfigChecksSharedRequiredSecrets(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/app\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("WS_SECRET=secret\n"), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Setenv("JWT_SECRET", "")

	result := ValidateConfig(tmp, ".env")
	if !result.Valid {
		t.Fatalf("expected config to remain valid with missing optional runtime secret warning: %#v", result)
	}
	found := false
	for _, warning := range result.Warnings {
		if warning.Field == "JWT_SECRET" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected JWT_SECRET warning, got %#v", result.Warnings)
	}
}

func TestValidateConfigReportsEnvParseError(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/app\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("INVALID\n"), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	result := ValidateConfig(tmp, ".env")
	if result.Valid {
		t.Fatal("expected invalid config for env parse error")
	}
	if len(result.Errors) != 1 || result.Errors[0].Type != "invalid_env_file" {
		t.Fatalf("expected invalid_env_file error, got %#v", result.Errors)
	}
}

func TestRedactSensitiveRedactsNestedSecrets(t *testing.T) {
	cfg := &Config{
		Config: map[string]any{
			"security": map[string]any{
				"ws_secret":  "secret",
				"jwt_key":    "key",
				"jwt_expiry": "15m",
			},
			"database": map[string]any{
				"db_url": "postgres://user:pass@example/db",
			},
		},
	}

	redacted := RedactSensitive(cfg)
	security := redacted.Config["security"].(map[string]any)
	if security["ws_secret"] != "***REDACTED***" {
		t.Fatalf("ws_secret was not redacted: %#v", security["ws_secret"])
	}
	if security["jwt_key"] != "***REDACTED***" {
		t.Fatalf("jwt_key was not redacted: %#v", security["jwt_key"])
	}
	if security["jwt_expiry"] != "15m" {
		t.Fatalf("jwt_expiry should remain visible: %#v", security["jwt_expiry"])
	}

	database := redacted.Config["database"].(map[string]any)
	if database["db_url"] != "***REDACTED***" {
		t.Fatalf("db_url was not redacted: %#v", database["db_url"])
	}
}
