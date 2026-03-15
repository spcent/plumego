package fileapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	storefile "github.com/spcent/plumego/store/file"
	datafile "github.com/spcent/plumego/x/data/file"
)

// --- mock Storage ---

type mockStorage struct {
	putFunc    func(ctx context.Context, opts datafile.PutOptions) (*datafile.File, error)
	getFunc    func(ctx context.Context, path string) (io.ReadCloser, error)
	deleteFunc func(ctx context.Context, path string) error
	getURLFunc func(ctx context.Context, path string, expiry time.Duration) (string, error)
}

func (m *mockStorage) Put(ctx context.Context, opts datafile.PutOptions) (*datafile.File, error) {
	if m.putFunc != nil {
		return m.putFunc(ctx, opts)
	}
	return &datafile.File{ID: "test-id", TenantID: opts.TenantID, Path: "test/path.txt"}, nil
}

func (m *mockStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, path)
	}
	return io.NopCloser(strings.NewReader("test content")), nil
}

func (m *mockStorage) Delete(ctx context.Context, path string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, path)
	}
	return nil
}

func (m *mockStorage) Exists(ctx context.Context, path string) (bool, error) { return true, nil }

func (m *mockStorage) Stat(ctx context.Context, path string) (*storefile.FileStat, error) {
	return &storefile.FileStat{Path: path, Size: 1024}, nil
}

func (m *mockStorage) List(ctx context.Context, prefix string, limit int) ([]*storefile.FileStat, error) {
	return []*storefile.FileStat{}, nil
}

func (m *mockStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	if m.getURLFunc != nil {
		return m.getURLFunc(ctx, path, expiry)
	}
	return "http://example.com/" + path, nil
}

func (m *mockStorage) Copy(ctx context.Context, src, dst string) error { return nil }

// --- mock MetadataManager ---

type mockMetadataManager struct {
	getFunc              func(ctx context.Context, id string) (*datafile.File, error)
	listFunc             func(ctx context.Context, q datafile.Query) ([]*datafile.File, int64, error)
	deleteFunc           func(ctx context.Context, id string) error
	updateAccessTimeFunc func(ctx context.Context, id string) error
}

func (m *mockMetadataManager) Save(ctx context.Context, file *datafile.File) error { return nil }

func (m *mockMetadataManager) Get(ctx context.Context, id string) (*datafile.File, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return &datafile.File{ID: id, Path: "test/path.txt", Name: "test.txt", Size: 1024, MimeType: "text/plain"}, nil
}

func (m *mockMetadataManager) GetByPath(ctx context.Context, path string) (*datafile.File, error) {
	return &datafile.File{Path: path}, nil
}

func (m *mockMetadataManager) GetByHash(ctx context.Context, hash string) (*datafile.File, error) {
	return nil, nil
}

func (m *mockMetadataManager) List(ctx context.Context, q datafile.Query) ([]*datafile.File, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, q)
	}
	return []*datafile.File{}, 0, nil
}

func (m *mockMetadataManager) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockMetadataManager) UpdateAccessTime(ctx context.Context, id string) error {
	if m.updateAccessTimeFunc != nil {
		return m.updateAccessTimeFunc(ctx, id)
	}
	return nil
}

// Compile-time checks
var _ datafile.Storage = (*mockStorage)(nil)
var _ datafile.MetadataManager = (*mockMetadataManager)(nil)

// --- Tests ---

func TestNewHandler(t *testing.T) {
	h := NewHandler(&mockStorage{}, &mockMetadataManager{})
	if h == nil {
		t.Fatal("Handler is nil")
	}
	if h.maxSize != 100<<20 {
		t.Errorf("Default maxSize = %d, want %d", h.maxSize, 100<<20)
	}
}

func TestHandler_WithMaxSize(t *testing.T) {
	h := NewHandler(&mockStorage{}, &mockMetadataManager{}).WithMaxSize(50 << 20)
	if h.maxSize != 50<<20 {
		t.Errorf("maxSize = %d, want %d", h.maxSize, 50<<20)
	}
}

func TestHandler_Upload(t *testing.T) {
	storage := &mockStorage{
		putFunc: func(ctx context.Context, opts datafile.PutOptions) (*datafile.File, error) {
			return &datafile.File{
				ID:       "test-id",
				TenantID: opts.TenantID,
				Name:     opts.FileName,
				Path:     "test/path.txt",
				Size:     100,
			}, nil
		},
	}
	h := NewHandler(storage, &mockMetadataManager{})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Upload(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var result datafile.File
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if result.ID != "test-id" {
		t.Errorf("ID = %q, want %q", result.ID, "test-id")
	}
	if result.TenantID != "tenant-123" {
		t.Errorf("TenantID = %q, want %q", result.TenantID, "tenant-123")
	}
}

func TestHandler_Upload_MissingTenantID(t *testing.T) {
	h := NewHandler(&mockStorage{}, &mockMetadataManager{})

	req := httptest.NewRequest(http.MethodPost, "/files", nil)
	w := httptest.NewRecorder()
	h.Upload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Download(t *testing.T) {
	storage := &mockStorage{
		getFunc: func(ctx context.Context, path string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("file content")), nil
		},
	}
	metadata := &mockMetadataManager{
		getFunc: func(ctx context.Context, id string) (*datafile.File, error) {
			return &datafile.File{ID: id, Path: "test/path.txt", Name: "test.txt", Size: 12, MimeType: "text/plain"}, nil
		},
	}
	h := NewHandler(storage, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()
	h.Download(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Content-Type = %q", w.Header().Get("Content-Type"))
	}
	if w.Body.String() != "file content" {
		t.Errorf("Body = %q, want %q", w.Body.String(), "file content")
	}
}

func TestHandler_Download_NotFound(t *testing.T) {
	metadata := &mockMetadataManager{
		getFunc: func(ctx context.Context, id string) (*datafile.File, error) {
			return nil, storefile.ErrNotFound
		},
	}
	h := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()
	h.Download(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_GetInfo(t *testing.T) {
	metadata := &mockMetadataManager{
		getFunc: func(ctx context.Context, id string) (*datafile.File, error) {
			return &datafile.File{ID: id, Name: "test.txt", Size: 1024, MimeType: "text/plain"}, nil
		},
	}
	h := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files/test-id/info", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()
	h.GetInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var result datafile.File
	json.NewDecoder(w.Body).Decode(&result)
	if result.ID != "test-id" {
		t.Errorf("ID = %q, want %q", result.ID, "test-id")
	}
}

func TestHandler_Delete(t *testing.T) {
	h := NewHandler(&mockStorage{}, &mockMetadataManager{})

	req := httptest.NewRequest(http.MethodDelete, "/files/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["message"] != "file deleted" {
		t.Errorf("Message = %q, want %q", result["message"], "file deleted")
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	metadata := &mockMetadataManager{
		deleteFunc: func(ctx context.Context, id string) error {
			return storefile.ErrNotFound
		},
	}
	h := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodDelete, "/files/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_List(t *testing.T) {
	metadata := &mockMetadataManager{
		listFunc: func(ctx context.Context, q datafile.Query) ([]*datafile.File, int64, error) {
			return []*datafile.File{
				{ID: "file1", Name: "test1.txt"},
				{ID: "file2", Name: "test2.txt"},
			}, 2, nil
		},
	}
	h := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if int(result["total"].(float64)) != 2 {
		t.Errorf("Total = %v, want 2", result["total"])
	}
}

func TestHandler_List_WithFilters(t *testing.T) {
	var capturedQuery datafile.Query
	metadata := &mockMetadataManager{
		listFunc: func(ctx context.Context, q datafile.Query) ([]*datafile.File, int64, error) {
			capturedQuery = q
			return []*datafile.File{}, 0, nil
		},
	}
	h := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files?tenant_id=tenant-123&mime_type=image/jpeg&page=2&page_size=10", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if capturedQuery.TenantID != "tenant-123" {
		t.Errorf("TenantID = %q, want tenant-123", capturedQuery.TenantID)
	}
	if capturedQuery.MimeType != "image/jpeg" {
		t.Errorf("MimeType = %q, want image/jpeg", capturedQuery.MimeType)
	}
	if capturedQuery.Page != 2 {
		t.Errorf("Page = %d, want 2", capturedQuery.Page)
	}
	if capturedQuery.PageSize != 10 {
		t.Errorf("PageSize = %d, want 10", capturedQuery.PageSize)
	}
}

func TestHandler_GetURL(t *testing.T) {
	storage := &mockStorage{
		getURLFunc: func(ctx context.Context, path string, expiry time.Duration) (string, error) {
			return "https://example.com/presigned-url", nil
		},
	}
	metadata := &mockMetadataManager{
		getFunc: func(ctx context.Context, id string) (*datafile.File, error) {
			return &datafile.File{ID: id, Path: "test/path.txt"}, nil
		},
	}
	h := NewHandler(storage, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files/test-id/url?expiry=3600", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()
	h.GetURL(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["url"] != "https://example.com/presigned-url" {
		t.Errorf("URL = %q", result["url"])
	}
	if result["expires_in"] != "3600" {
		t.Errorf("ExpiresIn = %q, want 3600", result["expires_in"])
	}
}

func TestHandler_Download_MetadataError(t *testing.T) {
	metadata := &mockMetadataManager{
		getFunc: func(ctx context.Context, id string) (*datafile.File, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()
	h.Download(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// Benchmarks

func BenchmarkHandler_Upload(b *testing.B) {
	h := NewHandler(&mockStorage{}, &mockMetadataManager{})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test content"))
	writer.Close()
	bodyBytes := body.Bytes()
	contentType := writer.FormDataContentType()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/files", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", contentType)
		ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		h.Upload(w, req)
	}
}

func BenchmarkHandler_Download(b *testing.B) {
	h := NewHandler(&mockStorage{}, &mockMetadataManager{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/files/test-id", nil)
		req.SetPathValue("id", "test-id")
		w := httptest.NewRecorder()
		h.Download(w, req)
	}
}
