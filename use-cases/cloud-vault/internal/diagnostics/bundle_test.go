package diagnostics

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-vault/internal/config"
	"cloud-vault/internal/database"
	"cloud-vault/internal/storage"
)

// openTestDB opens a temporary SQLite DB with migrations applied.
func openTestDB(t *testing.T) *database.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// buildService creates a Service with a real SQLite DB and local storage.
func buildService(t *testing.T) (*Service, string) {
	t.Helper()
	tmpDir := t.TempDir()
	db := openTestDB(t)
	store := storage.NewLocalStorage(filepath.Join(tmpDir, "objects"))
	outputDir := filepath.Join(tmpDir, "diagnostics")
	cfg := config.Config{
		DB: config.DBConfig{
			Path: filepath.Join(tmpDir, "app.db"),
		},
		App: config.AppConfig{
			Version: "1.2.3",
		},
	}
	svc := NewService(cfg, db, store, outputDir)
	return svc, tmpDir
}

// TestGenerateBundle_CreatesZip verifies that GenerateBundle returns a Bundle
// and that the file exists on disk.
func TestGenerateBundle_CreatesZip(t *testing.T) {
	svc, _ := buildService(t)
	ctx := context.Background()

	bundle, err := svc.GenerateBundle(ctx)
	if err != nil {
		t.Fatalf("GenerateBundle: %v", err)
	}

	if bundle == nil {
		t.Fatal("expected non-nil bundle")
	}
	if bundle.Filename == "" {
		t.Error("bundle Filename should not be empty")
	}
	if !strings.HasPrefix(bundle.Filename, "diagnostics_") {
		t.Errorf("filename %q should start with 'diagnostics_'", bundle.Filename)
	}
	if !strings.HasSuffix(bundle.Filename, ".zip") {
		t.Errorf("filename %q should end with '.zip'", bundle.Filename)
	}
	if bundle.Size <= 0 {
		t.Errorf("bundle Size = %d, want > 0", bundle.Size)
	}

	// File should exist at DownloadPath
	if _, err := os.Stat(bundle.DownloadPath); err != nil {
		t.Errorf("bundle file not found at %q: %v", bundle.DownloadPath, err)
	}
}

// TestGenerateBundle_ZipContents verifies mandatory entries are present in the zip.
func TestGenerateBundle_ZipContents(t *testing.T) {
	svc, _ := buildService(t)
	ctx := context.Background()

	bundle, err := svc.GenerateBundle(ctx)
	if err != nil {
		t.Fatalf("GenerateBundle: %v", err)
	}

	r, err := zip.OpenReader(bundle.DownloadPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()

	wantFiles := []string{"config.json", "system_info.json", "build_info.json", "manifest.json"}
	found := make(map[string]bool)
	for _, f := range r.File {
		found[f.Name] = true
	}
	for _, name := range wantFiles {
		if !found[name] {
			t.Errorf("zip missing %q; entries: %v", name, r.File)
		}
	}
}

// TestGenerateBundle_ConfigIsRedacted verifies that the config.json in the zip
// does NOT contain raw secret values.
func TestGenerateBundle_ConfigIsRedacted(t *testing.T) {
	tmpDir := t.TempDir()
	db := openTestDB(t)
	store := storage.NewLocalStorage(filepath.Join(tmpDir, "objects"))
	outputDir := filepath.Join(tmpDir, "diag")

	cfg := config.Config{
		DB: config.DBConfig{Path: filepath.Join(tmpDir, "app.db")},
		Qiniu: config.QiniuConfig{
			AccessKey: "super-secret-access",
			SecretKey: "super-secret-key",
			Bucket:    "my-bucket",
		},
		AI: config.AIConfig{
			APIKey: "sk-ant-api-key-12345",
		},
		Auth: config.AuthConfig{
			BootstrapAdminPassword: "admin-password-plain",
		},
	}
	svc := NewService(cfg, db, store, outputDir)
	ctx := context.Background()

	bundle, err := svc.GenerateBundle(ctx)
	if err != nil {
		t.Fatalf("GenerateBundle: %v", err)
	}

	data, err := os.ReadFile(bundle.DownloadPath)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	content := string(data)

	// Raw secrets must NOT appear in any form inside the zip bytes
	for _, secret := range []string{
		"super-secret-access",
		"super-secret-key",
		"sk-ant-api-key-12345",
		"admin-password-plain",
	} {
		if strings.Contains(content, secret) {
			t.Errorf("bundle contains raw secret %q", secret)
		}
	}
}

// TestListBundles_EmptyDir returns an empty slice when no bundles exist.
func TestListBundles_EmptyDir(t *testing.T) {
	svc, _ := buildService(t)

	bundles, err := svc.ListBundles()
	if err != nil {
		t.Fatalf("ListBundles: %v", err)
	}
	if len(bundles) != 0 {
		t.Errorf("expected 0 bundles, got %d", len(bundles))
	}
}

// TestListBundles_AfterGenerate verifies that ListBundles returns the generated bundle.
func TestListBundles_AfterGenerate(t *testing.T) {
	svc, _ := buildService(t)
	ctx := context.Background()

	_, err := svc.GenerateBundle(ctx)
	if err != nil {
		t.Fatalf("GenerateBundle: %v", err)
	}

	bundles, err := svc.ListBundles()
	if err != nil {
		t.Fatalf("ListBundles: %v", err)
	}
	if len(bundles) != 1 {
		t.Errorf("expected 1 bundle, got %d", len(bundles))
	}
}

// TestListBundles_IgnoresNonDiagnosticsFiles verifies that non-diagnostic files
// in the output directory are ignored.
func TestListBundles_IgnoresNonDiagnosticsFiles(t *testing.T) {
	svc, tmpDir := buildService(t)
	ctx := context.Background()

	// Create a diagnostics dir with a non-diagnostic file
	outputDir := filepath.Join(tmpDir, "diagnostics")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "other.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("writefile: %v", err)
	}

	// Generate one real bundle
	_, err := svc.GenerateBundle(ctx)
	if err != nil {
		t.Fatalf("GenerateBundle: %v", err)
	}

	bundles, err := svc.ListBundles()
	if err != nil {
		t.Fatalf("ListBundles: %v", err)
	}
	if len(bundles) != 1 {
		t.Errorf("expected 1 bundle (non-zip file should be ignored), got %d", len(bundles))
	}
}

// TestGetBundle_Found returns bundle metadata for an existing file.
func TestGetBundle_Found(t *testing.T) {
	svc, _ := buildService(t)
	ctx := context.Background()

	generated, err := svc.GenerateBundle(ctx)
	if err != nil {
		t.Fatalf("GenerateBundle: %v", err)
	}

	got, err := svc.GetBundle(generated.Filename)
	if err != nil {
		t.Fatalf("GetBundle: %v", err)
	}

	if got.Filename != generated.Filename {
		t.Errorf("Filename = %q, want %q", got.Filename, generated.Filename)
	}
	if got.Size <= 0 {
		t.Errorf("Size = %d, want > 0", got.Size)
	}
}

// TestGetBundle_NotFound returns an error for a missing bundle.
func TestGetBundle_NotFound(t *testing.T) {
	svc, _ := buildService(t)

	_, err := svc.GetBundle("nonexistent_bundle.zip")
	if err == nil {
		t.Error("expected error for missing bundle, got nil")
	}
}

// TestGetBundle_PathTraversal rejects filenames that contain directory separators.
func TestGetBundle_PathTraversal(t *testing.T) {
	svc, _ := buildService(t)

	_, err := svc.GetBundle("../some/other/file.zip")
	if err == nil {
		t.Error("expected error for path traversal attempt, got nil")
	}
}

// ---- redactor edge-case tests ----

// TestRedactKeyValueQuoted tests redaction of key="value" and key='value' patterns.
func TestRedactKeyValueQuoted(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
		absent   string
	}{
		{
			name:     "double-quoted password",
			input:    `password="s3cr3t!"`,
			contains: `"<redacted>"`,
			absent:   "s3cr3t!",
		},
		{
			name:     "single-quoted token",
			input:    `token='abc-token-123'`,
			contains: `"<redacted>"`,
			absent:   "abc-token-123",
		},
		{
			name:     "api_key double-quoted",
			input:    `api_key="my-api-key-value"`,
			contains: `"<redacted>"`,
			absent:   "my-api-key-value",
		},
		{
			name:     "mixed case SECRET_KEY",
			input:    `SECRET_KEY="topsecret"`,
			contains: `"<redacted>"`,
			absent:   "topsecret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactLogLine(tt.input)
			if strings.Contains(result, tt.absent) {
				t.Errorf("RedactLogLine(%q) = %q; still contains %q", tt.input, result, tt.absent)
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("RedactLogLine(%q) = %q; expected to contain %q", tt.input, result, tt.contains)
			}
		})
	}
}

// TestRedactLogLine_BearerCaseInsensitive covers lowercase 'bearer'.
func TestRedactLogLine_BearerCaseInsensitive(t *testing.T) {
	input := "auth: bearer eyJhbGciOiJIUzI1NiJ9.payload.sig"
	result := RedactLogLine(input)
	if strings.Contains(result, "eyJhbGciOiJIUzI1NiJ9") {
		t.Errorf("lowercase bearer token not redacted: %q", result)
	}
}

// TestRedactLogLine_JSONStylePassword covers {"password": "secret"}.
func TestRedactLogLine_JSONStylePassword(t *testing.T) {
	input := `{"user":"admin","password":"p@ssw0rd","status":"ok"}`
	result := RedactLogLine(input)
	if strings.Contains(result, "p@ssw0rd") {
		t.Errorf("JSON password not redacted: %q", result)
	}
	if !strings.Contains(result, "<redacted>") {
		t.Errorf("expected <redacted> in result: %q", result)
	}
	// non-sensitive fields should survive
	if !strings.Contains(result, "admin") {
		t.Errorf("non-sensitive 'admin' field was removed from: %q", result)
	}
}

// TestRedactLogLine_SessionKey covers 'session=<value>' patterns.
func TestRedactLogLine_SessionKey(t *testing.T) {
	input := "request session=sess-abc-123-xyz"
	result := RedactLogLine(input)
	if strings.Contains(result, "sess-abc-123-xyz") {
		t.Errorf("session value not redacted: %q", result)
	}
}

// TestRedactLogLine_MultipleKeysOnOneLine covers lines with multiple sensitive fields.
func TestRedactLogLine_MultipleKeysOnOneLine(t *testing.T) {
	input := `api_key=key123 password=pass456 username=alice`
	result := RedactLogLine(input)
	if strings.Contains(result, "key123") {
		t.Errorf("api_key not redacted: %q", result)
	}
	if strings.Contains(result, "pass456") {
		t.Errorf("password not redacted: %q", result)
	}
	// non-sensitive username should be preserved
	if !strings.Contains(result, "alice") {
		t.Errorf("non-sensitive 'alice' unexpectedly removed: %q", result)
	}
}

// TestRedactConfig_EmptySecrets verifies empty secret fields stay empty (not "<redacted>").
func TestRedactConfig_EmptySecrets(t *testing.T) {
	cfg := config.Config{
		Qiniu: config.QiniuConfig{
			AccessKey: "",
			SecretKey: "",
		},
		AI: config.AIConfig{
			APIKey: "",
		},
	}

	redacted := RedactConfig(cfg)

	if redacted.Storage["qiniu_access_key"] != "" {
		t.Errorf("empty access key should stay empty, got: %v", redacted.Storage["qiniu_access_key"])
	}
	if redacted.Storage["qiniu_secret_key"] != "" {
		t.Errorf("empty secret key should stay empty, got: %v", redacted.Storage["qiniu_secret_key"])
	}
	if redacted.AI["api_key"] != "" {
		t.Errorf("empty api key should stay empty, got: %v", redacted.AI["api_key"])
	}
}

// TestRedactConfig_AuthFields verifies auth-specific fields are handled correctly.
func TestRedactConfig_AuthFields(t *testing.T) {
	cfg := config.Config{
		Auth: config.AuthConfig{
			Enabled:                true,
			SessionTTLHours:        24,
			CookieName:             "session_id",
			SecureCookie:           true,
			BootstrapAdminPassword: "secret-password",
		},
	}

	redacted := RedactConfig(cfg)

	// Non-secret auth fields should be visible
	if redacted.Auth["session_ttl_hours"] != 24 {
		t.Errorf("session_ttl_hours = %v, want 24", redacted.Auth["session_ttl_hours"])
	}
	if redacted.Auth["enabled"] != true {
		t.Errorf("enabled = %v, want true", redacted.Auth["enabled"])
	}
	// Password must be redacted
	if redacted.Auth["bootstrap_admin_password"] != "<redacted>" {
		t.Errorf("bootstrap_admin_password = %v, want '<redacted>'", redacted.Auth["bootstrap_admin_password"])
	}
}

// TestSplitLines handles CRLF, LF, and no trailing newline.
func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "unix newlines",
			input: "line1\nline2\nline3",
			want:  []string{"line1", "line2", "line3"},
		},
		{
			name:  "windows CRLF",
			input: "line1\r\nline2\r\nline3",
			want:  []string{"line1", "line2", "line3"},
		},
		{
			name:  "trailing newline",
			input: "line1\nline2\n",
			want:  []string{"line1", "line2"},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single line no newline",
			input: "hello world",
			want:  []string{"hello world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitLines(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.want, len(tt.want))
				return
			}
			for i, line := range got {
				if line != tt.want[i] {
					t.Errorf("splitLines(%q)[%d] = %q, want %q", tt.input, i, line, tt.want[i])
				}
			}
		})
	}
}

// TestCollectSystemInfo verifies collectSystemInfo returns a populated struct.
func TestCollectSystemInfo(t *testing.T) {
	info := collectSystemInfo()
	if info.OS == "" {
		t.Error("OS should not be empty")
	}
	if info.Arch == "" {
		t.Error("Arch should not be empty")
	}
	if info.NumCPU <= 0 {
		t.Errorf("NumCPU = %d, want > 0", info.NumCPU)
	}
	if info.NumGoroutine <= 0 {
		t.Errorf("NumGoroutine = %d, want > 0", info.NumGoroutine)
	}
}

// TestCollectRecentLogs_WithLogFile verifies log collection when a log file exists.
func TestCollectRecentLogs_WithLogFile(t *testing.T) {
	logDir := t.TempDir()

	// Create a log file with sensitive content
	logContent := "2026-01-01 INFO starting\n2026-01-01 DEBUG api_key=sk-secret123 connected\n"
	if err := os.WriteFile(filepath.Join(logDir, "app.log"), []byte(logContent), 0644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	logs, err := collectRecentLogs(logDir, 1000)
	if err != nil {
		t.Fatalf("collectRecentLogs: %v", err)
	}

	if logs == "" {
		t.Error("expected non-empty logs")
	}
	// Sensitive value should have been redacted
	if strings.Contains(logs, "sk-secret123") {
		t.Error("log collection did not redact api_key value")
	}
	// Non-sensitive content should survive
	if !strings.Contains(logs, "starting") {
		t.Error("non-sensitive log content was lost")
	}
}

// TestCollectRecentLogs_EmptyDir returns empty string for empty directory.
func TestCollectRecentLogs_EmptyDir(t *testing.T) {
	logDir := t.TempDir()

	logs, err := collectRecentLogs(logDir, 100)
	if err != nil {
		t.Fatalf("collectRecentLogs empty dir: %v", err)
	}
	if logs != "" {
		t.Errorf("expected empty logs for empty dir, got %q", logs)
	}
}

// TestCollectRecentLogs_Truncation verifies that only maxLines are kept.
func TestCollectRecentLogs_Truncation(t *testing.T) {
	logDir := t.TempDir()

	var sb strings.Builder
	for i := 0; i < 50; i++ {
		sb.WriteString("line\n")
	}
	if err := os.WriteFile(filepath.Join(logDir, "app.log"), []byte(sb.String()), 0644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	// Request only 10 lines
	logs, err := collectRecentLogs(logDir, 10)
	if err != nil {
		t.Fatalf("collectRecentLogs: %v", err)
	}

	lines := splitLines(strings.TrimRight(logs, "\n"))
	if len(lines) > 10 {
		t.Errorf("expected at most 10 lines, got %d", len(lines))
	}
}
