package file

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newS3Server starts a minimal httptest server that responds to basic S3 ops.
// It stores objects in memory.
func newS3Server(t *testing.T) (*httptest.Server, map[string][]byte) {
	t.Helper()
	store := make(map[string][]byte)

	mux := http.NewServeMux()

	// PUT /{bucket}/{key} — upload
	// HEAD /{bucket}/{key} — exists/stat
	// GET /{bucket}/{key} — download
	// DELETE /{bucket}/{key} — delete
	// GET /{bucket}?list-type=2 — list

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Extract key (strip leading "/bucket/")
		path := r.URL.Path
		parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
		key := ""
		if len(parts) == 2 {
			key = parts[1]
		}

		// List
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

// newTestS3Storage creates an S3Storage pointing to the test server.
func newTestS3Storage(t *testing.T, srv *httptest.Server) *S3Storage {
	t.Helper()
	// Extract host from server URL (strip "http://").
	host := strings.TrimPrefix(srv.URL, "http://")
	s, err := NewS3Storage(StorageConfig{
		S3Endpoint:  host,
		S3Bucket:    "testbucket",
		S3UseSSL:    false,
		S3PathStyle: true,
	}, nil)
	if err != nil {
		t.Fatalf("NewS3Storage: %v", err)
	}
	// Override the HTTP client to not follow redirects and to not timeout quickly.
	s.client = &http.Client{}
	return s
}

func TestS3Storage_Put_Get(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := context.Background()

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

func TestS3Storage_Get_NotFound(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := context.Background()

	_, err := s.Get(ctx, "nonexistent/key.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestS3Storage_Delete(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := context.Background()

	// Manually insert into mock store.
	store["mykey.txt"] = []byte("data")

	if err := s.Delete(ctx, "mykey.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := store["mykey.txt"]; ok {
		t.Error("key should be deleted from mock store")
	}
}

func TestS3Storage_Delete_NotFound(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := context.Background()

	err := s.Delete(ctx, "ghost.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestS3Storage_Exists(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := context.Background()

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

func TestS3Storage_Stat(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := context.Background()

	_, err := s.Stat(ctx, "missing.txt")
	if !errors.Is(err, ErrNotFound) {
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
	ctx := context.Background()

	url, err := s.GetURL(ctx, "some/file.txt", time.Minute)
	if err != nil {
		t.Fatalf("GetURL: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty URL")
	}
}

func TestS3Storage_GetURL_DefaultExpiry(t *testing.T) {
	srv, _ := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := context.Background()

	// Zero expiry should use default.
	url, err := s.GetURL(ctx, "some/file.txt", 0)
	if err != nil {
		t.Fatalf("GetURL with zero expiry: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty URL")
	}
}

func TestS3Storage_List(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := context.Background()

	store["t1/file1.txt"] = []byte("a")
	store["t1/file2.txt"] = []byte("b")
	store["t2/file3.txt"] = []byte("c")

	files, err := s.List(ctx, "t1/", 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("List count = %d, want 2", len(files))
	}
}

func TestS3Storage_Copy(t *testing.T) {
	srv, store := newS3Server(t)
	s := newTestS3Storage(t, srv)
	ctx := context.Background()

	store["src/file.txt"] = []byte("copy me")

	if err := s.Copy(ctx, "src/file.txt", "dst/file.txt"); err != nil {
		t.Fatalf("Copy: %v", err)
	}
}

func TestS3Storage_Put_Deduplication(t *testing.T) {
	srv, _ := newS3Server(t)
	host := strings.TrimPrefix(srv.URL, "http://")
	s, _ := NewS3Storage(StorageConfig{
		S3Endpoint:  host,
		S3Bucket:    "testbucket",
		S3PathStyle: true,
	}, &mockMetadata{})
	s.client = &http.Client{}

	ctx := context.Background()
	content := []byte("deduplicated s3 content")

	first, err := s.Put(ctx, PutOptions{
		TenantID: "t1", Reader: bytes.NewReader(content), FileName: "dup.bin",
	})
	if err != nil {
		t.Fatalf("first Put: %v", err)
	}

	second, err := s.Put(ctx, PutOptions{
		TenantID: "t1", Reader: bytes.NewReader(content), FileName: "dup.bin",
	})
	if err != nil {
		t.Fatalf("second Put: %v", err)
	}
	if second.Hash != first.Hash {
		t.Errorf("expected deduplication, hash mismatch: %q vs %q", first.Hash, second.Hash)
	}
}
