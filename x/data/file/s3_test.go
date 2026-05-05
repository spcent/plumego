package file

import (
	"bytes"

	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	storefile "github.com/spcent/plumego/store/file"
)

func newS3Server(t *testing.T) (*httptest.Server, map[string][]byte) {
	t.Helper()
	store := make(map[string][]byte)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pathStr := r.URL.Path
		parts := strings.SplitN(strings.TrimPrefix(pathStr, "/"), "/", 2)
		key := ""
		if len(parts) == 2 {
			key = parts[1]
		}

		if r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2" {
			type content struct {
				XMLName      struct{}  `xml:"Contents"`
				Key          string    `xml:"Key"`
				Size         int64     `xml:"Size"`
				LastModified time.Time `xml:"LastModified"`
			}
			type listResult struct {
				XMLName  xml.Name  `xml:"ListBucketResult"`
				Contents []content `xml:"Contents"`
			}
			var result listResult
			prefix := r.URL.Query().Get("prefix")
			for k, v := range store {
				if prefix == "" || strings.HasPrefix(k, prefix) {
					result.Contents = append(result.Contents, content{
						Key:          k,
						Size:         int64(len(v)),
						LastModified: time.Now(),
					})
				}
			}
			w.Header().Set("Content-Type", "application/xml")
			xml.NewEncoder(w).Encode(result)
			return
		}

		switch r.Method {
		case http.MethodPut:
			if source := r.Header.Get("x-amz-copy-source"); source != "" {
				sourceKey := strings.TrimPrefix(source, "/testbucket/")
				data, ok := store[sourceKey]
				if !ok {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				store[key] = append([]byte(nil), data...)
				w.WriteHeader(http.StatusOK)
				return
			}
			body, _ := io.ReadAll(r.Body)
			store[key] = body
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			data, ok := store[key]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case http.MethodHead:
			if key == "forbidden.txt" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			data, ok := store[key]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			if _, ok := store[key]; !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			delete(store, key)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, store
}

func newTestS3Storage(t *testing.T, srv *httptest.Server) *S3Storage {
	t.Helper()
	host := strings.TrimPrefix(srv.URL, "http://")
	s, err := NewS3Storage(S3Config{
		Endpoint:  host,
		Bucket:    "testbucket",
		UseSSL:    false,
		PathStyle: true,
	}, nil)
	if err != nil {
		t.Fatalf("NewS3Storage: %v", err)
	}
	s.client = &http.Client{}
	return s
}

func TestS3Storage_Put_Get(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := t.Context()

	content := []byte("hello s3")
	result, err := s.Put(ctx, PutOptions{
		TenantID:    "t1",
		Reader:      bytes.NewReader(content),
		FileName:    "test.txt",
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if result.StorageType != "s3" {
		t.Errorf("StorageType = %q, want s3", result.StorageType)
	}
	if result.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", result.Size, len(content))
	}

	reader, err := s.Get(ctx, result.Path)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	if !bytes.Equal(got, content) {
		t.Errorf("Get content = %q, want %q", got, content)
	}
}

func TestS3Storage_PutRejectsUnsafeTenantID(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)

	for _, tenantID := range []string{"../tenant", "tenant/child", `tenant\child`, ""} {
		t.Run(tenantID, func(t *testing.T) {
			_, err := s.Put(t.Context(), PutOptions{
				TenantID: tenantID,
				Reader:   strings.NewReader("data"),
				FileName: "unsafe.txt",
			})
			if !errors.Is(err, storefile.ErrInvalidPath) {
				t.Fatalf("Put error = %v, want ErrInvalidPath", err)
			}
		})
	}
	if len(store) != 0 {
		t.Fatalf("unsafe tenant uploads should not reach server, stored keys=%d", len(store))
	}
}

func TestS3Storage_Put_RejectsKnownOversizedUpload(t *testing.T) {
	srv, store := newS3Server(t)
	host := strings.TrimPrefix(srv.URL, "http://")
	s, err := NewS3Storage(S3Config{
		Endpoint:      host,
		Bucket:        "testbucket",
		UseSSL:        false,
		PathStyle:     true,
		MaxUploadSize: 4,
	}, nil)
	if err != nil {
		t.Fatalf("NewS3Storage: %v", err)
	}
	s.client = &http.Client{}

	_, err = s.Put(t.Context(), PutOptions{
		TenantID: "t1",
		Reader:   strings.NewReader("hello"),
		FileName: "too-large.txt",
		Size:     5,
	})
	if !errors.Is(err, storefile.ErrInvalidSize) {
		t.Fatalf("Put error = %v, want ErrInvalidSize", err)
	}
	if len(store) != 0 {
		t.Fatalf("oversized upload should not reach server, stored keys=%d", len(store))
	}
}

func TestS3Storage_Put_RejectsUnknownOversizedUpload(t *testing.T) {
	srv, store := newS3Server(t)
	host := strings.TrimPrefix(srv.URL, "http://")
	s, err := NewS3Storage(S3Config{
		Endpoint:      host,
		Bucket:        "testbucket",
		UseSSL:        false,
		PathStyle:     true,
		MaxUploadSize: 4,
	}, nil)
	if err != nil {
		t.Fatalf("NewS3Storage: %v", err)
	}
	s.client = &http.Client{}

	_, err = s.Put(t.Context(), PutOptions{
		TenantID: "t1",
		Reader:   strings.NewReader("hello"),
		FileName: "too-large.txt",
		Size:     -1,
	})
	if !errors.Is(err, storefile.ErrInvalidSize) {
		t.Fatalf("Put error = %v, want ErrInvalidSize", err)
	}
	if len(store) != 0 {
		t.Fatalf("oversized upload should not reach server, stored keys=%d", len(store))
	}
}

func TestS3Storage_Get_NotFound(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)

	_, err := s.Get(t.Context(), "nonexistent/key.txt")
	if !errors.Is(err, storefile.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestS3Storage_RejectsUnsafePaths(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)

	if _, err := s.Get(t.Context(), "../secret.txt"); !errors.Is(err, storefile.ErrInvalidPath) {
		t.Fatalf("Get error = %v, want ErrInvalidPath", err)
	}
	if err := s.Delete(t.Context(), "/absolute.txt"); !errors.Is(err, storefile.ErrInvalidPath) {
		t.Fatalf("Delete error = %v, want ErrInvalidPath", err)
	}
	if _, err := s.Exists(t.Context(), `tenant\file.txt`); !errors.Is(err, storefile.ErrInvalidPath) {
		t.Fatalf("Exists error = %v, want ErrInvalidPath", err)
	}
	if _, err := s.Stat(t.Context(), "tenant/../file.txt"); !errors.Is(err, storefile.ErrInvalidPath) {
		t.Fatalf("Stat error = %v, want ErrInvalidPath", err)
	}
	if _, err := s.List(t.Context(), "../tenant/", 0); !errors.Is(err, storefile.ErrInvalidPath) {
		t.Fatalf("List error = %v, want ErrInvalidPath", err)
	}
	if _, err := s.GetURL(t.Context(), "../tenant/file.txt", time.Minute); !errors.Is(err, storefile.ErrInvalidPath) {
		t.Fatalf("GetURL error = %v, want ErrInvalidPath", err)
	}
	if err := s.Copy(t.Context(), "../src.txt", "dst.txt"); !errors.Is(err, storefile.ErrInvalidPath) {
		t.Fatalf("Copy source error = %v, want ErrInvalidPath", err)
	}
	if err := s.Copy(t.Context(), "src.txt", "../dst.txt"); !errors.Is(err, storefile.ErrInvalidPath) {
		t.Fatalf("Copy destination error = %v, want ErrInvalidPath", err)
	}
}

func TestS3Storage_Delete(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)

	store["mykey.txt"] = []byte("data")

	if err := s.Delete(t.Context(), "mykey.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := store["mykey.txt"]; ok {
		t.Error("key should be deleted from mock store")
	}
}

func TestS3Storage_Delete_NotFound(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)

	err := s.Delete(t.Context(), "ghost.txt")
	if !errors.Is(err, storefile.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestS3Storage_Exists(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := t.Context()

	exists, err := s.Exists(ctx, "nope.txt")
	if err != nil || exists {
		t.Errorf("expected not found, got exists=%v err=%v", exists, err)
	}

	store["exists.txt"] = []byte("yes")
	exists, err = s.Exists(ctx, "exists.txt")
	if err != nil || !exists {
		t.Errorf("expected found, got exists=%v err=%v", exists, err)
	}
}

func TestS3Storage_Exists_NonOKStatusReturnsError(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)

	exists, err := s.Exists(t.Context(), "forbidden.txt")
	if err == nil || exists {
		t.Fatalf("Exists = %v, %v; want false with error", exists, err)
	}
	var fileErr *storefile.Error
	if !errors.As(err, &fileErr) {
		t.Fatalf("Exists error = %T, want *storefile.Error", err)
	}
	if fileErr.Op != "Exists" || fileErr.Path != "forbidden.txt" {
		t.Fatalf("Exists file error = %+v, want op Exists and path forbidden.txt", fileErr)
	}
}

func TestS3Storage_Stat(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := t.Context()

	_, err := s.Stat(ctx, "missing.txt")
	if !errors.Is(err, storefile.ErrNotFound) {
		t.Errorf("expected ErrNotFound for missing file, got %v", err)
	}

	store["stat.txt"] = []byte("content")
	stat, err := s.Stat(ctx, "stat.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if stat.Path != "stat.txt" {
		t.Errorf("Path = %q, want stat.txt", stat.Path)
	}
}

func TestS3Storage_GetURL(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)

	url, err := s.GetURL(t.Context(), "some/file.txt", time.Minute)
	if err != nil {
		t.Fatalf("GetURL: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty URL")
	}
}

func TestS3Storage_List(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)

	store["t1/file2.txt"] = []byte("b")
	store["t1/file1.txt"] = []byte("a")
	store["t2/file3.txt"] = []byte("c")

	files, err := s.List(t.Context(), "t1/", 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("List count = %d, want 2", len(files))
	}
	if files[0].Path != "t1/file1.txt" || files[1].Path != "t1/file2.txt" {
		t.Fatalf("List paths should be sorted, got %q then %q", files[0].Path, files[1].Path)
	}
}

func TestS3Storage_List_NegativeLimit(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)

	_, err := s.List(t.Context(), "t1/", -1)
	if !errors.Is(err, storefile.ErrInvalidSize) {
		t.Fatalf("List negative limit error = %v, want ErrInvalidSize", err)
	}
}

func TestS3Storage_Copy(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)

	store["src/file.txt"] = []byte("copy me")

	if err := s.Copy(t.Context(), "src/file.txt", "dst/file.txt"); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if got := string(store["dst/file.txt"]); got != "copy me" {
		t.Fatalf("copied content = %q, want copy me", got)
	}

	store["src/file.txt"] = []byte("replacement")
	if err := s.Copy(t.Context(), "src/file.txt", "dst/file.txt"); err != nil {
		t.Fatalf("Copy overwrite: %v", err)
	}
	if got := string(store["dst/file.txt"]); got != "replacement" {
		t.Fatalf("overwritten content = %q, want replacement", got)
	}
}

func TestS3Storage_Copy_MissingSourceWrapsNotFound(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)

	err := s.Copy(t.Context(), "missing/file.txt", "dst/file.txt")
	if !errors.Is(err, storefile.ErrNotFound) {
		t.Fatalf("Copy error = %v, want ErrNotFound", err)
	}
	var fileErr *storefile.Error
	if !errors.As(err, &fileErr) {
		t.Fatalf("Copy error = %T, want *storefile.Error", err)
	}
	if fileErr.Op != "Copy" || fileErr.Path != "missing/file.txt" {
		t.Fatalf("Copy file error = %+v, want op Copy and source path", fileErr)
	}
}

func TestS3Storage_Put_Deduplication(t *testing.T) {
	srv, _ := newS3Server(t)
	host := strings.TrimPrefix(srv.URL, "http://")
	s, _ := NewS3Storage(S3Config{
		Endpoint:  host,
		Bucket:    "testbucket",
		PathStyle: true,
	}, &mockMetadata{})
	s.client = &http.Client{}

	ctx := t.Context()
	content := []byte("deduplicated s3 content")

	first, err := s.Put(ctx, PutOptions{TenantID: "t1", Reader: bytes.NewReader(content), FileName: "dup.bin"})
	if err != nil {
		t.Fatalf("first Put: %v", err)
	}

	second, err := s.Put(ctx, PutOptions{TenantID: "t1", Reader: bytes.NewReader(content), FileName: "dup.bin"})
	if err != nil {
		t.Fatalf("second Put: %v", err)
	}
	if second.Hash != first.Hash {
		t.Errorf("expected deduplication, hash mismatch: %q vs %q", first.Hash, second.Hash)
	}
}

func TestS3Storage_Put_DeduplicationIsTenantScoped(t *testing.T) {
	srv, store := newS3Server(t)
	host := strings.TrimPrefix(srv.URL, "http://")
	s, _ := NewS3Storage(S3Config{
		Endpoint:  host,
		Bucket:    "testbucket",
		PathStyle: true,
	}, &mockMetadata{})
	s.client = &http.Client{}

	ctx := t.Context()
	content := []byte("tenant scoped s3 duplicate content")

	first, err := s.Put(ctx, PutOptions{TenantID: "t1", Reader: bytes.NewReader(content), FileName: "dup.bin"})
	if err != nil {
		t.Fatalf("first Put: %v", err)
	}

	second, err := s.Put(ctx, PutOptions{TenantID: "t2", Reader: bytes.NewReader(content), FileName: "dup.bin"})
	if err != nil {
		t.Fatalf("second Put: %v", err)
	}
	if second.TenantID != "t2" {
		t.Fatalf("second Put returned tenant %q, want t2", second.TenantID)
	}
	if second.Path == first.Path {
		t.Fatalf("cross-tenant duplicate reused path %q", second.Path)
	}
	if len(store) != 2 {
		t.Fatalf("stored object count = %d, want 2", len(store))
	}
}

func TestS3Storage_buildURL_VirtualHosted(t *testing.T) {
	s := &S3Storage{endpoint: "s3.amazonaws.com", bucket: "mybucket", useSSL: true, pathStyle: false}
	got := s.buildURL("tenant/file.txt")
	if !strings.HasPrefix(got, "https://mybucket.s3.amazonaws.com/") {
		t.Errorf("buildURL = %q, expected virtual-hosted HTTPS prefix", got)
	}
	if !strings.Contains(got, "file.txt") {
		t.Errorf("buildURL = %q, expected file.txt in URL", got)
	}
	if strings.Contains(got, "%2F") {
		t.Errorf("buildURL = %q, object key slash should remain a path separator", got)
	}
}

func TestS3Storage_buildURL_PathStyle(t *testing.T) {
	s := &S3Storage{endpoint: "minio.local:9000", bucket: "testbucket", useSSL: false, pathStyle: true}
	got := s.buildURL("folder/object.png")
	if !strings.HasPrefix(got, "http://minio.local:9000/testbucket/") {
		t.Errorf("buildURL = %q, expected path-style HTTP prefix", got)
	}
	if !strings.Contains(got, "object.png") {
		t.Errorf("buildURL = %q, expected object.png in URL", got)
	}
	if strings.Contains(got, "%2F") {
		t.Errorf("buildURL = %q, object key slash should remain a path separator", got)
	}
}

func TestS3Storage_buildURL_EscapesSegments(t *testing.T) {
	s := &S3Storage{endpoint: "s3.amazonaws.com", bucket: "mybucket", useSSL: true, pathStyle: true}
	got := s.buildURL("tenant/my file.txt")
	if !strings.Contains(got, "tenant/my%20file.txt") {
		t.Errorf("buildURL = %q, expected segment escaping", got)
	}
	if strings.Contains(got, "tenant%2F") {
		t.Errorf("buildURL = %q, slash should not be escaped", got)
	}
}

func TestNewS3Storage_MissingConfig(t *testing.T) {
	_, err := NewS3Storage(S3Config{}, nil)
	if err == nil {
		t.Fatal("expected error for missing S3 config")
	}
}

func TestNewS3Storage_DefaultRegion(t *testing.T) {
	s, err := NewS3Storage(S3Config{
		Endpoint: "s3.amazonaws.com",
		Bucket:   "my-bucket",
	}, nil)
	if err != nil {
		t.Fatalf("NewS3Storage: %v", err)
	}
	if s.region != "us-east-1" {
		t.Errorf("region = %q, want us-east-1", s.region)
	}
	if s.maxUploadSize != DefaultS3MaxUploadSize {
		t.Errorf("maxUploadSize = %d, want %d", s.maxUploadSize, DefaultS3MaxUploadSize)
	}
}
