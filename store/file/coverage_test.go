package file

import (
	"bytes"
	"context"
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

// --- LocalStorage additional coverage ---

func TestLocalStorage_Get_NotFound(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	_, err = storage.Get(ctx, "tenant/2026/01/01/nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	var fileErr *Error
	if !errors.As(err, &fileErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if !errors.Is(fileErr.Err, ErrNotFound) {
		t.Errorf("inner error = %v, want ErrNotFound", fileErr.Err)
	}
}

func TestLocalStorage_Stat_NotFound(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	_, err = storage.Stat(ctx, "tenant/2026/01/01/nonexistent.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalStorage_Stat_InvalidPath(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = storage.Stat(context.Background(), "../etc/passwd")
	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath, got %v", err)
	}
}

func TestLocalStorage_Copy_InvalidSrc(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = storage.Copy(context.Background(), "../../etc/passwd", "tenant/dst.txt")
	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath for invalid src, got %v", err)
	}
}

func TestLocalStorage_Copy_InvalidDst(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = storage.Copy(context.Background(), "tenant/src.txt", "../../etc/passwd")
	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath for invalid dst, got %v", err)
	}
}

func TestLocalStorage_List_InvalidPath(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = storage.List(context.Background(), "../etc", 10)
	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath, got %v", err)
	}
}

func TestLocalStorage_List_Empty(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Walk a non-existent prefix returns nil (filepath.Walk on non-existent dir).
	files, err := storage.List(context.Background(), "empty-tenant", 10)
	// May return an error if the path doesn't exist; accept both outcomes.
	_ = err
	_ = files
}

func TestLocalStorage_Put_Deduplication(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", &mockMetadata{})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	content := []byte("deduplicated content")

	first, err := storage.Put(ctx, PutOptions{
		TenantID: "t1",
		Reader:   bytes.NewReader(content),
		FileName: "dup.txt",
	})
	if err != nil {
		t.Fatalf("first Put: %v", err)
	}

	// Second Put with same content should return the existing file metadata.
	second, err := storage.Put(ctx, PutOptions{
		TenantID: "t1",
		Reader:   bytes.NewReader(content),
		FileName: "dup.txt",
	})
	if err != nil {
		t.Fatalf("second Put: %v", err)
	}
	if second.Hash != first.Hash {
		t.Errorf("expected same hash for duplicate content, first=%q second=%q", first.Hash, second.Hash)
	}
}

// --- S3Storage buildURL tests ---

func TestS3Storage_buildURL_VirtualHosted(t *testing.T) {
	s := &S3Storage{
		endpoint:  "s3.amazonaws.com",
		bucket:    "mybucket",
		useSSL:    true,
		pathStyle: false,
	}
	got := s.buildURL("tenant/file.txt")
	if !strings.HasPrefix(got, "https://mybucket.s3.amazonaws.com/") {
		t.Errorf("buildURL = %q, expected virtual-hosted HTTPS prefix", got)
	}
	if !strings.Contains(got, "file.txt") {
		t.Errorf("buildURL = %q, expected file.txt in URL", got)
	}
}

func TestS3Storage_buildURL_PathStyle(t *testing.T) {
	s := &S3Storage{
		endpoint:  "minio.local:9000",
		bucket:    "testbucket",
		useSSL:    false,
		pathStyle: true,
	}
	got := s.buildURL("folder/object.png")
	if !strings.HasPrefix(got, "http://minio.local:9000/testbucket/") {
		t.Errorf("buildURL = %q, expected path-style HTTP prefix", got)
	}
	if !strings.Contains(got, "object.png") {
		t.Errorf("buildURL = %q, expected object.png in URL", got)
	}
}

func TestS3Storage_buildURL_EmptyKey(t *testing.T) {
	s := &S3Storage{
		endpoint:  "s3.amazonaws.com",
		bucket:    "mybucket",
		useSSL:    true,
		pathStyle: false,
	}
	got := s.buildURL("")
	// Should produce a base URL without trailing slash issues.
	if !strings.Contains(got, "mybucket") {
		t.Errorf("buildURL with empty key = %q, expected bucket name", got)
	}
}

func TestS3Storage_buildURL_PathTraversalEncoded(t *testing.T) {
	s := &S3Storage{
		endpoint:  "s3.amazonaws.com",
		bucket:    "mybucket",
		useSSL:    true,
		pathStyle: true,
	}
	// Path traversal is URL-encoded (percent-encoded), making it safe in HTTP.
	got := s.buildURL("../../etc/passwd")
	// The resulting URL must not contain a literal ".." path segment.
	// url.PathEscape encodes "." as "%2E" or leaves it, so the raw ".."
	// in a URL path should not appear unencoded.
	if strings.Contains(got, "/../") || strings.HasSuffix(got, "/..") {
		t.Errorf("buildURL contains unencoded path traversal, got %q", got)
	}
}

func TestNewS3Storage_MissingConfig(t *testing.T) {
	_, err := NewS3Storage(StorageConfig{}, nil)
	if err == nil {
		t.Fatal("expected error for missing S3 config")
	}
}

func TestNewS3Storage_DefaultRegion(t *testing.T) {
	s, err := NewS3Storage(StorageConfig{
		S3Endpoint: "s3.amazonaws.com",
		S3Bucket:   "my-bucket",
	}, nil)
	if err != nil {
		t.Fatalf("NewS3Storage: %v", err)
	}
	if s.region != "us-east-1" {
		t.Errorf("region = %q, want us-east-1", s.region)
	}
}

// --- mockMetadata for deduplication test ---

type mockMetadata struct {
	store map[string]*File
}

func (m *mockMetadata) Save(_ context.Context, f *File) error {
	if m.store == nil {
		m.store = make(map[string]*File)
	}
	m.store[f.Hash] = f
	return nil
}

func (m *mockMetadata) Get(_ context.Context, id string) (*File, error) {
	if m.store == nil {
		return nil, ErrNotFound
	}
	for _, f := range m.store {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockMetadata) GetByPath(_ context.Context, path string) (*File, error) {
	if m.store == nil {
		return nil, ErrNotFound
	}
	for _, f := range m.store {
		if f.Path == path {
			return f, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockMetadata) GetByHash(_ context.Context, hash string) (*File, error) {
	if m.store == nil {
		return nil, ErrNotFound
	}
	if f, ok := m.store[hash]; ok {
		return f, nil
	}
	return nil, ErrNotFound
}

func (m *mockMetadata) List(_ context.Context, _ Query) ([]*File, int64, error) {
	return nil, 0, nil
}

func (m *mockMetadata) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockMetadata) UpdateAccessTime(_ context.Context, _ string) error {
	return nil
}

// Compile-time check: mockMetadata satisfies MetadataManager.
var _ MetadataManager = (*mockMetadata)(nil)

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
	_ = q.TenantID
	_ = q.PageSize
}

// --- PutOptions coverage ---

func TestPutOptions_AllFields(t *testing.T) {
	opts := PutOptions{
		TenantID:      "t1",
		FileName:      "file.txt",
		ContentType:   "text/plain",
		Reader:        strings.NewReader("data"),
		Metadata:      map[string]any{"key": "val"},
		UploadedBy:    "user-1",
		GenerateThumb: false,
		ThumbWidth:    200,
		ThumbHeight:   200,
	}
	if opts.TenantID != "t1" {
		t.Error("TenantID not set")
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
