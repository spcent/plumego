package configmgr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvFileSkipsCommentsAndBlankLines(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".env")
	content := "\n# comment\nAPP_ADDR=:8080\n WS_SECRET = secret \nINVALID\n"
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
	if _, ok := vars["INVALID"]; ok {
		t.Fatalf("invalid line should be skipped: %#v", vars)
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
