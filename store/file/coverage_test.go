package file

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// --- Error type tests ---

func TestError_Error_WithPath(t *testing.T) {
	e := &Error{Op: "Put", Path: "tenant/file.txt", Err: ErrNotFound}
	msg := e.Error()
	if !strings.Contains(msg, "Put") {
		t.Errorf("Error() should contain op, got %q", msg)
	}
	if !strings.Contains(msg, "tenant/file.txt") {
		t.Errorf("Error() should contain path, got %q", msg)
	}
}

func TestError_Error_WithoutPath(t *testing.T) {
	e := &Error{Op: "Delete", Err: ErrNotFound}
	msg := e.Error()
	if !strings.Contains(msg, "Delete") {
		t.Errorf("Error() should contain op, got %q", msg)
	}
}

func TestError_Unwrap(t *testing.T) {
	e := &Error{Op: "Get", Err: ErrNotFound}
	if !errors.Is(e, ErrNotFound) {
		t.Error("Unwrap should allow errors.Is to match ErrNotFound")
	}
}

// --- StorageConfig and types smoke tests ---

func TestStorageConfig_Zero(t *testing.T) {
	var cfg StorageConfig
	// Zero value should not panic when accessed.
	_ = cfg.LocalBasePath
	_ = cfg.S3Endpoint
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
	_ = q.UploadedBy
	_ = q.PageSize
}

// --- PutOptions coverage ---

func TestPutOptions_AllFields(t *testing.T) {
	opts := PutOptions{
		FileName:      "file.txt",
		ContentType:   "text/plain",
		Reader:        strings.NewReader("data"),
		Metadata:      map[string]any{"key": "val"},
		UploadedBy:    "user-1",
		GenerateThumb: false,
		ThumbWidth:    200,
		ThumbHeight:   200,
	}
	if opts.FileName != "file.txt" {
		t.Error("FileName not set")
	}
}

// --- ImageInfo tests ---

func TestImageInfo_Zero(t *testing.T) {
	var info ImageInfo
	if info.Width != 0 || info.Height != 0 {
		t.Error("expected zero dimensions")
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
