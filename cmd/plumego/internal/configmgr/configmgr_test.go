package configmgr

import "testing"

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
