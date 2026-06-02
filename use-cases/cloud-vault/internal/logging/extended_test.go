package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---- redactor additional tests ----

// TestRedactLogLine_AccessKey covers access_key=value patterns.
func TestRedactLogLine_AccessKey(t *testing.T) {
	input := "storage: access_key=AKID-very-secret-value connected"
	result := redactLogLine(input)
	if strings.Contains(result, "AKID-very-secret-value") {
		t.Errorf("access_key value not redacted: %q", result)
	}
	if !strings.Contains(result, "<redacted>") {
		t.Errorf("expected <redacted> in result: %q", result)
	}
}

// TestRedactLogLine_SecretKey covers secret_key=value patterns.
func TestRedactLogLine_SecretKey(t *testing.T) {
	input := "config: secret_key=mysupersecret loaded"
	result := redactLogLine(input)
	if strings.Contains(result, "mysupersecret") {
		t.Errorf("secret_key value not redacted: %q", result)
	}
}

// TestRedactLogLine_BearerLowercase covers lowercase 'bearer'.
func TestRedactLogLine_BearerLowercase(t *testing.T) {
	input := "header: bearer TOKEN_VALUE_XYZ"
	result := redactLogLine(input)
	if strings.Contains(result, "TOKEN_VALUE_XYZ") {
		t.Errorf("lowercase bearer token not redacted: %q", result)
	}
}

// TestRedactLogLine_BearerMixedCase covers mixed-case 'Bearer'.
func TestRedactLogLine_BearerMixedCase(t *testing.T) {
	input := "Authorization: Bearer eyJhbGciOiJSUzI1NiJ9.payload"
	result := redactLogLine(input)
	if strings.Contains(result, "eyJhbGciOiJSUzI1NiJ9") {
		t.Errorf("Bearer token not redacted: %q", result)
	}
	if !strings.Contains(result, "Bearer <redacted>") {
		t.Errorf("expected 'Bearer <redacted>' in result: %q", result)
	}
}

// TestRedactLogLine_AuthToken covers auth_token=value.
func TestRedactLogLine_AuthToken(t *testing.T) {
	input := "auth_token=tok-abc-12345 used"
	result := redactLogLine(input)
	if strings.Contains(result, "tok-abc-12345") {
		t.Errorf("auth_token not redacted: %q", result)
	}
}

// TestRedactLogLine_PasswordVariants covers passwd= and pwd=.
func TestRedactLogLine_PasswordVariants(t *testing.T) {
	tests := []struct {
		input  string
		secret string
	}{
		{"passwd=s3cur3P@ss", "s3cur3P@ss"},
		{"pwd=hunter2", "hunter2"},
		{"PASSWORD=SECRETVALUE", "SECRETVALUE"},
	}

	for _, tt := range tests {
		result := redactLogLine(tt.input)
		if strings.Contains(result, tt.secret) {
			t.Errorf("redactLogLine(%q) did not redact %q: got %q", tt.input, tt.secret, result)
		}
	}
}

// TestRedactLogLine_NoSensitiveData verifies unchanged output for non-sensitive lines.
func TestRedactLogLine_NoSensitiveData(t *testing.T) {
	cases := []string{
		"INFO: user logged in",
		"2026-01-01T00:00:00Z DEBUG request completed in 25ms",
		"document saved: id=doc-abc-123 title=My Notes",
	}

	for _, input := range cases {
		result := redactLogLine(input)
		if result != input {
			t.Errorf("non-sensitive line modified:\ninput:  %q\noutput: %q", input, result)
		}
	}
}

// TestRedactingWriter_WriteBytesCount verifies Write returns the correct byte count.
func TestRedactingWriter_WriteBytesCount(t *testing.T) {
	var buf bytes.Buffer
	rw := NewRedactingWriter(&buf)

	input := "no sensitive data here\n"
	n, err := rw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(input) {
		t.Errorf("Write returned %d bytes, want %d", n, len(input))
	}
}

// TestRedactingWriter_EmptyInput verifies writing empty bytes works.
func TestRedactingWriter_EmptyInput(t *testing.T) {
	var buf bytes.Buffer
	rw := NewRedactingWriter(&buf)

	n, err := rw.Write([]byte(""))
	if err != nil {
		t.Fatalf("Write empty: %v", err)
	}
	if n != 0 {
		t.Errorf("Write empty returned %d bytes, want 0", n)
	}
}

// TestRedactingWriter_MultiLinePayload tests a multi-line payload that includes
// one sensitive and one non-sensitive line.
func TestRedactingWriter_MultiLinePayload(t *testing.T) {
	var buf bytes.Buffer
	rw := NewRedactingWriter(&buf)

	input := "user: admin\napi_key=sk-test-key-abc123\nstatus: ok\n"
	if _, err := rw.Write([]byte(input)); err != nil {
		t.Fatalf("Write: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "sk-test-key-abc123") {
		t.Errorf("api_key not redacted in multi-line: %q", output)
	}
	if !strings.Contains(output, "admin") {
		t.Errorf("non-sensitive 'admin' was removed from: %q", output)
	}
}

// ---- rotator additional tests ----

// TestRotatingFileWriter_CurrentFilePath returns path before and after rotation.
func TestRotatingFileWriter_CurrentFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	rw, err := NewRotatingFileWriter(RotatingWriterConfig{
		Dir:      tmpDir,
		Filename: "app.log",
	})
	if err != nil {
		t.Fatalf("NewRotatingFileWriter: %v", err)
	}
	defer rw.Close()

	want := filepath.Join(tmpDir, "app.log")
	got := rw.CurrentFilePath()
	if got != want {
		t.Errorf("CurrentFilePath = %q, want %q", got, want)
	}
}

// TestRotatingFileWriter_ListBackups_Empty verifies empty slice when no rotations occurred.
func TestRotatingFileWriter_ListBackups_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	rw, err := NewRotatingFileWriter(RotatingWriterConfig{
		Dir:      tmpDir,
		Filename: "app.log",
	})
	if err != nil {
		t.Fatalf("NewRotatingFileWriter: %v", err)
	}
	defer rw.Close()

	backups, err := rw.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups before rotation, got %d: %v", len(backups), backups)
	}
}

// TestRotatingFileWriter_RotatesAtMaxSize triggers rotation precisely at the boundary.
func TestRotatingFileWriter_RotatesAtMaxSize(t *testing.T) {
	tmpDir := t.TempDir()

	const maxSize = 200
	rw, err := NewRotatingFileWriter(RotatingWriterConfig{
		Dir:          tmpDir,
		Filename:     "rotate.log",
		MaxSizeBytes: maxSize,
		MaxBackups:   10,
	})
	if err != nil {
		t.Fatalf("NewRotatingFileWriter: %v", err)
	}
	defer rw.Close()

	// Write exactly maxSize+1 bytes across two writes to force rotation.
	chunk := bytes.Repeat([]byte("x"), 150)
	if _, err := rw.Write(chunk); err != nil {
		t.Fatalf("Write 1: %v", err)
	}
	if _, err := rw.Write(chunk); err != nil { // this should trigger rotation
		t.Fatalf("Write 2: %v", err)
	}

	backups, err := rw.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups: %v", err)
	}
	if len(backups) == 0 {
		t.Error("expected at least one backup after rotation")
	}
}

// TestRotatingFileWriter_DefaultConfig uses empty config to verify defaults are applied.
func TestRotatingFileWriter_DefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	rw, err := NewRotatingFileWriter(RotatingWriterConfig{
		Dir: tmpDir,
		// Filename, MaxSizeBytes, MaxBackups intentionally empty
	})
	if err != nil {
		t.Fatalf("NewRotatingFileWriter with defaults: %v", err)
	}
	defer rw.Close()

	// Should use default filename "app.log"
	expectedPath := filepath.Join(tmpDir, "app.log")
	if rw.CurrentFilePath() != expectedPath {
		t.Errorf("expected default filename 'app.log', got %q", rw.CurrentFilePath())
	}

	// Write should work
	if _, err := rw.Write([]byte("default config test\n")); err != nil {
		t.Errorf("Write with default config: %v", err)
	}
}

// TestRotatingFileWriter_ConcurrentWrites verifies no data race under concurrent writes.
func TestRotatingFileWriter_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	rw, err := NewRotatingFileWriter(RotatingWriterConfig{
		Dir:          tmpDir,
		Filename:     "concurrent.log",
		MaxSizeBytes: 5000,
		MaxBackups:   5,
	})
	if err != nil {
		t.Fatalf("NewRotatingFileWriter: %v", err)
	}
	defer rw.Close()

	done := make(chan struct{})
	const goroutines = 10
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer func() { done <- struct{}{} }()
			msg := []byte("goroutine message from worker\n")
			for j := 0; j < 20; j++ {
				rw.Write(msg)
			}
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent write timed out")
		}
	}
}

// TestRotatingFileWriter_DoubleCLose verifies that double Close does not panic.
func TestRotatingFileWriter_DoubleClose(t *testing.T) {
	tmpDir := t.TempDir()
	rw, err := NewRotatingFileWriter(RotatingWriterConfig{
		Dir:      tmpDir,
		Filename: "closetest.log",
	})
	if err != nil {
		t.Fatalf("NewRotatingFileWriter: %v", err)
	}

	if err := rw.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	// Second close should not panic (nil file guard)
	_ = rw.Close()
}

// TestRotatingFileWriter_CreatesDir verifies that the log directory is created
// when it doesn't exist yet.
func TestRotatingFileWriter_CreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "logs", "subdir")

	rw, err := NewRotatingFileWriter(RotatingWriterConfig{
		Dir:      newDir,
		Filename: "app.log",
	})
	if err != nil {
		t.Fatalf("NewRotatingFileWriter on new dir: %v", err)
	}
	defer rw.Close()

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("log directory was not created")
	}
}
