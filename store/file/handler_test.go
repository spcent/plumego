package file

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockStorage implements Storage interface for testing
type mockStorage struct {
	putFunc    func(ctx context.Context, opts PutOptions) (*File, error)
	getFunc    func(ctx context.Context, path string) (io.ReadCloser, error)
	deleteFunc func(ctx context.Context, path string) error
	existsFunc func(ctx context.Context, path string) (bool, error)
	statFunc   func(ctx context.Context, path string) (*FileStat, error)
	listFunc   func(ctx context.Context, prefix string, limit int) ([]*FileStat, error)
	getURLFunc func(ctx context.Context, path string, expiry time.Duration) (string, error)
	copyFunc   func(ctx context.Context, srcPath, dstPath string) error
}

func (m *mockStorage) Put(ctx context.Context, opts PutOptions) (*File, error) {
	if m.putFunc != nil {
		return m.putFunc(ctx, opts)
	}
	return &File{ID: "test-id", Path: "test/path.txt"}, nil
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

func (m *mockStorage) Exists(ctx context.Context, path string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, path)
	}
	return true, nil
}

func (m *mockStorage) Stat(ctx context.Context, path string) (*FileStat, error) {
	if m.statFunc != nil {
		return m.statFunc(ctx, path)
	}
	return &FileStat{Path: path, Size: 1024}, nil
}

func (m *mockStorage) List(ctx context.Context, prefix string, limit int) ([]*FileStat, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, prefix, limit)
	}
	return []*FileStat{}, nil
}

func (m *mockStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	if m.getURLFunc != nil {
		return m.getURLFunc(ctx, path, expiry)
	}
	return "http://example.com/" + path, nil
}

func (m *mockStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
	if m.copyFunc != nil {
		return m.copyFunc(ctx, srcPath, dstPath)
	}
	return nil
}

// mockMetadataManager implements MetadataManager interface for testing
type mockMetadataManager struct {
	saveFunc             func(ctx context.Context, file *File) error
	getFunc              func(ctx context.Context, id string) (*File, error)
	getByPathFunc        func(ctx context.Context, path string) (*File, error)
	getByHashFunc        func(ctx context.Context, hash string) (*File, error)
	listFunc             func(ctx context.Context, query Query) ([]*File, int64, error)
	deleteFunc           func(ctx context.Context, id string) error
	updateAccessTimeFunc func(ctx context.Context, id string) error
}

func (m *mockMetadataManager) Save(ctx context.Context, file *File) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, file)
	}
	return nil
}

func (m *mockMetadataManager) Get(ctx context.Context, id string) (*File, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return &File{
		ID:       id,
		Path:     "test/path.txt",
		Name:     "test.txt",
		Size:     1024,
		MimeType: "text/plain",
	}, nil
}

func (m *mockMetadataManager) GetByPath(ctx context.Context, path string) (*File, error) {
	if m.getByPathFunc != nil {
		return m.getByPathFunc(ctx, path)
	}
	return &File{Path: path}, nil
}

func (m *mockMetadataManager) GetByHash(ctx context.Context, hash string) (*File, error) {
	if m.getByHashFunc != nil {
		return m.getByHashFunc(ctx, hash)
	}
	return nil, nil
}

func (m *mockMetadataManager) List(ctx context.Context, query Query) ([]*File, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, query)
	}
	return []*File{}, 0, nil
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

func TestNewHandler(t *testing.T) {
	storage := &mockStorage{}
	metadata := &mockMetadataManager{}

	handler := NewHandler(storage, metadata)

	if handler == nil {
		t.Fatal("Handler is nil")
	}
	if handler.maxSize != 100<<20 {
		t.Errorf("Default maxSize = %d, want %d", handler.maxSize, 100<<20)
	}
}

func TestHandler_WithMaxSize(t *testing.T) {
	storage := &mockStorage{}
	metadata := &mockMetadataManager{}

	handler := NewHandler(storage, metadata).WithMaxSize(50 << 20)

	if handler.maxSize != 50<<20 {
		t.Errorf("maxSize = %d, want %d", handler.maxSize, 50<<20)
	}
}

func TestHandler_Upload(t *testing.T) {
	storage := &mockStorage{
		putFunc: func(ctx context.Context, opts PutOptions) (*File, error) {
			return &File{
				ID:       "test-id",
				TenantID: opts.TenantID,
				Name:     opts.FileName,
				Path:     "test/path.txt",
				Size:     100,
			}, nil
		},
	}
	metadata := &mockMetadataManager{}
	handler := NewHandler(storage, metadata)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test content"))
	writer.Close()

	// Create request with tenant context
	req := httptest.NewRequest(http.MethodPost, "/files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	req = req.WithContext(ctx)

	// Create response recorder
	w := httptest.NewRecorder()

	// Handle request
	handler.Upload(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	// Parse response
	var result File
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
	storage := &mockStorage{}
	metadata := &mockMetadataManager{}
	handler := NewHandler(storage, metadata)

	req := httptest.NewRequest(http.MethodPost, "/files", nil)
	w := httptest.NewRecorder()

	handler.Upload(w, req)

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
		getFunc: func(ctx context.Context, id string) (*File, error) {
			return &File{
				ID:       id,
				Path:     "test/path.txt",
				Name:     "test.txt",
				Size:     12,
				MimeType: "text/plain",
			}, nil
		},
	}
	handler := NewHandler(storage, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handler.Download(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify headers
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Content-Type = %q, want %q", w.Header().Get("Content-Type"), "text/plain")
	}
	if w.Header().Get("Content-Length") != "12" {
		t.Errorf("Content-Length = %q, want %q", w.Header().Get("Content-Length"), "12")
	}

	// Verify content
	body := w.Body.String()
	if body != "file content" {
		t.Errorf("Body = %q, want %q", body, "file content")
	}
}

func TestHandler_Download_NotFound(t *testing.T) {
	storage := &mockStorage{}
	metadata := &mockMetadataManager{
		getFunc: func(ctx context.Context, id string) (*File, error) {
			return nil, ErrNotFound
		},
	}
	handler := NewHandler(storage, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	handler.Download(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_GetInfo(t *testing.T) {
	metadata := &mockMetadataManager{
		getFunc: func(ctx context.Context, id string) (*File, error) {
			return &File{
				ID:       id,
				Name:     "test.txt",
				Size:     1024,
				MimeType: "text/plain",
			}, nil
		},
	}
	handler := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files/test-id/info", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handler.GetInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var result File
	json.NewDecoder(w.Body).Decode(&result)

	if result.ID != "test-id" {
		t.Errorf("ID = %q, want %q", result.ID, "test-id")
	}
	if result.Name != "test.txt" {
		t.Errorf("Name = %q, want %q", result.Name, "test.txt")
	}
}

func TestHandler_Delete(t *testing.T) {
	metadata := &mockMetadataManager{
		deleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}
	handler := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodDelete, "/files/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handler.Delete(w, req)

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
			return ErrNotFound
		},
	}
	handler := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodDelete, "/files/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_List(t *testing.T) {
	metadata := &mockMetadataManager{
		listFunc: func(ctx context.Context, query Query) ([]*File, int64, error) {
			files := []*File{
				{ID: "file1", Name: "test1.txt"},
				{ID: "file2", Name: "test2.txt"},
			}
			return files, 2, nil
		},
	}
	handler := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files?page=1&page_size=20", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)

	if int(result["total"].(float64)) != 2 {
		t.Errorf("Total = %v, want 2", result["total"])
	}
	if int(result["page"].(float64)) != 1 {
		t.Errorf("Page = %v, want 1", result["page"])
	}
	if int(result["page_size"].(float64)) != 20 {
		t.Errorf("PageSize = %v, want 20", result["page_size"])
	}
}

func TestHandler_List_WithFilters(t *testing.T) {
	var capturedQuery Query

	metadata := &mockMetadataManager{
		listFunc: func(ctx context.Context, query Query) ([]*File, int64, error) {
			capturedQuery = query
			return []*File{}, 0, nil
		},
	}
	handler := NewHandler(&mockStorage{}, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files?tenant_id=tenant-123&mime_type=image/jpeg&page=2&page_size=10", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if capturedQuery.TenantID != "tenant-123" {
		t.Errorf("TenantID = %q, want %q", capturedQuery.TenantID, "tenant-123")
	}
	if capturedQuery.MimeType != "image/jpeg" {
		t.Errorf("MimeType = %q, want %q", capturedQuery.MimeType, "image/jpeg")
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
		getFunc: func(ctx context.Context, id string) (*File, error) {
			return &File{ID: id, Path: "test/path.txt"}, nil
		},
	}
	handler := NewHandler(storage, metadata)

	req := httptest.NewRequest(http.MethodGet, "/files/test-id/url?expiry=3600", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handler.GetURL(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)

	if result["url"] != "https://example.com/presigned-url" {
		t.Errorf("URL = %q, want %q", result["url"], "https://example.com/presigned-url")
	}
	if result["expires_in"] != "3600" {
		t.Errorf("ExpiresIn = %q, want %q", result["expires_in"], "3600")
	}
}

// Benchmark handler operations
func BenchmarkHandler_Upload(b *testing.B) {
	storage := &mockStorage{}
	metadata := &mockMetadataManager{}
	handler := NewHandler(storage, metadata)

	// Prepare request body
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

		handler.Upload(w, req)
	}
}

func BenchmarkHandler_Download(b *testing.B) {
	storage := &mockStorage{}
	metadata := &mockMetadataManager{}
	handler := NewHandler(storage, metadata)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/files/test-id", nil)
		req.SetPathValue("id", "test-id")
		w := httptest.NewRecorder()

		handler.Download(w, req)
	}
}
