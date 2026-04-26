package file

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func assertStringContains(t *testing.T, value string, expected string) {
	t.Helper()

	if !strings.Contains(value, expected) {
		t.Fatalf("expected %q to contain %q", value, expected)
	}
}

// --- Error type tests ---

func TestError_Error_WithPath(t *testing.T) {
	e := &Error{Op: "Put", Path: "tenant/file.txt", Err: ErrNotFound}
	msg := e.Error()
	assertStringContains(t, msg, "Put")
	assertStringContains(t, msg, "tenant/file.txt")
}

func TestError_Error_WithoutPath(t *testing.T) {
	e := &Error{Op: "Delete", Err: ErrNotFound}
	msg := e.Error()
	assertStringContains(t, msg, "Delete")
}

func TestError_Error_NilReceiver(t *testing.T) {
	var e *Error
	if got := e.Error(); got != "file: <nil>" {
		t.Fatalf("unexpected nil receiver error string: %q", got)
	}
}

func TestError_Error_WithoutCause(t *testing.T) {
	e := &Error{Op: "Stat", Path: "file.txt"}
	msg := e.Error()
	assertStringContains(t, msg, "Stat")
	assertStringContains(t, msg, "file.txt")
	if strings.Contains(msg, "<nil>") {
		t.Fatalf("error string should not expose nil cause: %q", msg)
	}
}

func TestError_Unwrap(t *testing.T) {
	e := &Error{Op: "Get", Err: ErrNotFound}
	if !errors.Is(e, ErrNotFound) {
		t.Error("Unwrap should allow errors.Is to match ErrNotFound")
	}
}

func TestError_Unwrap_NilReceiver(t *testing.T) {
	var e *Error
	if err := e.Unwrap(); err != nil {
		t.Fatalf("expected nil unwrap from nil receiver, got %v", err)
	}
}

func TestFileStat_Zero(t *testing.T) {
	var stat FileStat
	if stat.Size != 0 {
		t.Error("expected zero size")
	}
}

func TestFile_Zero(t *testing.T) {
	var f File
	if !f.CreatedAt.IsZero() {
		t.Error("expected zero CreatedAt")
	}
}

func TestQuery_Zero(t *testing.T) {
	var q Query
	_ = q.MimeType
	_ = q.PageSize
}

// --- PutOptions coverage ---

func TestPutOptions_AllFields(t *testing.T) {
	opts := PutOptions{
		FileName:    "file.txt",
		ContentType: "text/plain",
		Reader:      strings.NewReader("data"),
		Metadata:    map[string]any{"key": "val"},
	}
	if opts.FileName != "file.txt" {
		t.Error("FileName not set")
	}
}

// --- time field zero check ---

func TestFileStat_ModifiedTime(t *testing.T) {
	stat := FileStat{
		Path:         "some/path.txt",
		Size:         100,
		ModifiedTime: time.Now(),
		IsDir:        false,
	}
	if stat.ModifiedTime.IsZero() {
		t.Error("ModifiedTime should not be zero")
	}
}
