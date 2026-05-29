package config_test

import (
	"errors"
	"testing"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/config"
	"guardus/internal/storage"
)

func testLogger() plumelog.StructuredLogger {
	return plumelog.NewLogger(plumelog.LoggerConfig{Format: plumelog.LoggerFormatDiscard})
}

// resetEnv clears every GUARDUS_* variable for the duration of the test
// so individual cases can build a clean environment regardless of test order.
func resetEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"GUARDUS_METRICS",
		"GUARDUS_CONCURRENCY",
		"GUARDUS_WEB_ADDRESS",
		"GUARDUS_WEB_PORT",
		"GUARDUS_WEB_READ_BUFFER_SIZE",
		"GUARDUS_WEB_TLS_CERT_FILE",
		"GUARDUS_WEB_TLS_KEY_FILE",
		"GUARDUS_STORAGE_TYPE",
		"GUARDUS_STORAGE_PATH",
		"GUARDUS_STORAGE_CACHING",
		"GUARDUS_STORAGE_MAX_RESULTS",
		"GUARDUS_STORAGE_MAX_EVENTS",
		"GUARDUS_BOOTSTRAP_FILE",
		"GUARDUS_SECURITY_BASIC_USERNAME",
		"GUARDUS_SECURITY_BASIC_PASSWORD_BCRYPT_BASE64",
		"GUARDUS_SECURITY_BASIC_PASSWORD_PBKDF2",
		"GUARDUS_UI_TITLE",
		"GUARDUS_UI_DARK_MODE",
		"GUARDUS_UI_DEFAULT_SORT_BY",
		"GUARDUS_UI_DEFAULT_FILTER_BY",
		"GUARDUS_ALERTING_SLACK",
		"GUARDUS_ALERTING_DISCORD",
		"GUARDUS_ALERTING_TELEGRAM",
		"GUARDUS_ALERTING_EMAIL",
		"GUARDUS_ALERTING_CUSTOM",
		"GUARDUS_CONNECTIVITY_TARGET",
		"GUARDUS_CONNECTIVITY_INTERVAL",
		"GUARDUS_MAINTENANCE_ENABLED",
		"GUARDUS_MAINTENANCE_START",
		"GUARDUS_MAINTENANCE_DURATION",
		"GUARDUS_MAINTENANCE_TIMEZONE",
		"GUARDUS_MAINTENANCE_EVERY",
	} {
		t.Setenv(k, "")
	}
}

func TestLoadFromEnv_Defaults(t *testing.T) {
	resetEnv(t)
	cfg, err := config.LoadFromEnv(testLogger())
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if cfg.Web == nil || cfg.Web.Port == 0 {
		t.Errorf("web default not applied: %+v", cfg.Web)
	}
	if cfg.Storage == nil || cfg.Storage.Type != storage.TypeMemory {
		t.Errorf("storage default not memory: %+v", cfg.Storage)
	}
	if cfg.Concurrency != config.DefaultConcurrency {
		t.Errorf("concurrency: got %d want %d", cfg.Concurrency, config.DefaultConcurrency)
	}
	if cfg.Security != nil {
		t.Errorf("security must default to nil: %+v", cfg.Security)
	}
	if cfg.Alerting != nil {
		t.Errorf("alerting must default to nil: %+v", cfg.Alerting)
	}
}

func TestLoadFromEnv_WebAndStorage(t *testing.T) {
	resetEnv(t)
	t.Setenv("GUARDUS_WEB_ADDRESS", "127.0.0.1")
	t.Setenv("GUARDUS_WEB_PORT", "9090")
	t.Setenv("GUARDUS_STORAGE_TYPE", "sqlite")
	t.Setenv("GUARDUS_STORAGE_PATH", "./tmp.db")
	cfg, err := config.LoadFromEnv(testLogger())
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if cfg.Web.Address != "127.0.0.1" || cfg.Web.Port != 9090 {
		t.Errorf("web override not applied: %+v", cfg.Web)
	}
	if cfg.Storage.Type != storage.TypeSQLite || cfg.Storage.Path != "./tmp.db" {
		t.Errorf("storage override not applied: %+v", cfg.Storage)
	}
}

func TestLoadFromEnv_SQLiteRequiresPath(t *testing.T) {
	resetEnv(t)
	t.Setenv("GUARDUS_STORAGE_TYPE", "sqlite")
	_, err := config.LoadFromEnv(testLogger())
	if !errors.Is(err, storage.ErrSQLStorageRequiresPath) {
		t.Fatalf("expected ErrSQLStorageRequiresPath, got %v", err)
	}
}

func TestLoadFromEnv_SecurityIncomplete(t *testing.T) {
	resetEnv(t)
	t.Setenv("GUARDUS_SECURITY_BASIC_USERNAME", "admin")
	_, err := config.LoadFromEnv(testLogger())
	if !errors.Is(err, config.ErrInvalidSecurityConfig) {
		t.Fatalf("expected ErrInvalidSecurityConfig, got %v", err)
	}
}

func TestLoadFromEnv_AlertingSlackJSON(t *testing.T) {
	resetEnv(t)
	t.Setenv("GUARDUS_ALERTING_SLACK", `{"default-config":{"webhook-url":"https://hooks.slack.com/services/x/y/z"}}`)
	cfg, err := config.LoadFromEnv(testLogger())
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}
	if cfg.Alerting == nil || cfg.Alerting.Slack == nil {
		t.Fatalf("slack provider not parsed: %+v", cfg.Alerting)
	}
	if cfg.Alerting.Slack.DefaultConfig.WebhookURL == "" {
		t.Errorf("webhook-url not parsed: %+v", cfg.Alerting.Slack)
	}
}

func TestLoadFromEnv_AlertingInvalidJSON(t *testing.T) {
	resetEnv(t)
	t.Setenv("GUARDUS_ALERTING_SLACK", "not-json")
	_, err := config.LoadFromEnv(testLogger())
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadFromEnv_ConnectivityRequires53(t *testing.T) {
	resetEnv(t)
	t.Setenv("GUARDUS_CONNECTIVITY_TARGET", "1.1.1.1")
	_, err := config.LoadFromEnv(testLogger())
	if err == nil {
		t.Fatal("expected error for connectivity target without :53 suffix")
	}
}
