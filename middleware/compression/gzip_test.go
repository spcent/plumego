package compression

import (
	"bytes"
	"compress/gzip"
	"github.com/spcent/plumego/middleware"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzip_SmallResponse(t *testing.T) {
	// Small response should be compressed
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}

	wrapped := middleware.ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip encoding, got %q", rr.Header().Get("Content-Encoding"))
	}

	// Verify content is actually compressed
	reader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if string(decompressed) != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", string(decompressed))
	}
}

func TestGzip_LargeResponse(t *testing.T) {
	// Large response should be compressed
	largeData := strings.Repeat("A", 2048) // 2KB > 1KB threshold

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeData))
	}

	wrapped := middleware.ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip encoding, got %q", rr.Header().Get("Content-Encoding"))
	}
}

func TestGzip_SkipWebSocket(t *testing.T) {
	// WebSocket upgrade should not be compressed
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusSwitchingProtocols)
		w.Write([]byte("upgrade"))
	}

	wrapped := middleware.ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusSwitchingProtocols {
		t.Fatalf("expected status %d, got %d", http.StatusSwitchingProtocols, rr.Code)
	}

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("websocket should not be compressed")
	}

	if rr.Body.String() != "upgrade" {
		t.Fatalf("expected 'upgrade', got %q", rr.Body.String())
	}
}

func TestGzip_SkipSSE(t *testing.T) {
	// SSE should not be compressed
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: hello\n\n"))
	}

	wrapped := middleware.ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Accept", "text/event-stream")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("SSE should not be compressed")
	}

	if rr.Body.String() != "data: hello\n\n" {
		t.Fatalf("expected SSE data, got %q", rr.Body.String())
	}
}

func TestGzip_SkipBinaryContent(t *testing.T) {
	// Binary content should not be compressed
	testCases := []struct {
		contentType string
		data        []byte
	}{
		{"image/png", []byte{0x89, 0x50, 0x4E, 0x47}}, // PNG header
		{"video/mp4", []byte{0x00, 0x00, 0x00, 0x20}}, // MP4 header
		{"application/zip", []byte{0x50, 0x4B, 0x03, 0x04}},
		{"application/pdf", []byte("%PDF-1.4")},
	}

	for _, tc := range testCases {
		t.Run(tc.contentType, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tc.contentType)
				w.WriteHeader(http.StatusOK)
				w.Write(tc.data)
			}

			wrapped := middleware.ApplyFunc(handler, Gzip())

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rr := httptest.NewRecorder()

			wrapped(rr, req)

			if rr.Header().Get("Content-Encoding") == "gzip" {
				t.Fatalf("%s should not be compressed", tc.contentType)
			}

			if !bytes.Equal(rr.Body.Bytes(), tc.data) {
				t.Fatalf("data mismatch for %s", tc.contentType)
			}
		})
	}
}

func TestGzip_SkipAlreadyCompressed(t *testing.T) {
	// Already compressed content should not be compressed again
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("already compressed"))
	}

	wrapped := middleware.ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("should preserve existing gzip encoding")
	}

	if rr.Body.String() != "already compressed" {
		t.Fatalf("expected 'already compressed', got %q", rr.Body.String())
	}
}

func TestGzip_SkipNoAcceptEncoding(t *testing.T) {
	// Should not compress if client doesn't support gzip
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}

	wrapped := middleware.ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Accept-Encoding header
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("should not compress without Accept-Encoding")
	}

	if rr.Body.String() != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", rr.Body.String())
	}
}

func TestGzip_SkipErrorResponses(t *testing.T) {
	// Error responses should not be compressed
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	}

	wrapped := middleware.ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("error responses should not be compressed")
	}

	if rr.Body.String() != "Error" {
		t.Fatalf("expected 'Error', got %q", rr.Body.String())
	}
}

func TestGzip_CustomMaxBuffer(t *testing.T) {
	// Test custom max buffer configuration
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World")) // 11 bytes
	}

	cfg := GzipConfig{MaxBufferBytes: 5} // Only compress if <= 5 bytes
	wrapped := middleware.ApplyFunc(handler, GzipWithConfig(cfg))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Should not compress because 11 bytes > 5 bytes max buffer
	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("should not compress large response with custom max buffer")
	}

	if rr.Body.String() != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", rr.Body.String())
	}
}

func TestGzip_VaryHeader(t *testing.T) {
	// Vary header should be set when compression is used
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}

	wrapped := middleware.ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Header().Get("Vary") != "Accept-Encoding" {
		t.Fatalf("expected Vary: Accept-Encoding, got %q", rr.Header().Get("Vary"))
	}
}

func TestGzip_EmptyResponse(t *testing.T) {
	// Empty response should not cause issues
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}

	wrapped := middleware.ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestGzip_ChunkedWrite(t *testing.T) {
	// Test that large responses that exceed max buffer are written in chunks
	// Create a response larger than the default 10MB buffer
	largeData := strings.Repeat("X", 11<<20) // 11MB

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeData))
	}

	// Use a custom config with a small max buffer to trigger chunked writing
	cfg := GzipConfig{MaxBufferBytes: 1 << 20} // 1MB max buffer
	wrapped := middleware.ApplyFunc(handler, GzipWithConfig(cfg))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Should not compress because response exceeds max buffer
	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("should not compress large response that exceeds max buffer")
	}

	// Verify the full data was written
	if rr.Body.Len() != len(largeData) {
		t.Fatalf("expected %d bytes, got %d", len(largeData), rr.Body.Len())
	}

	// Verify content integrity
	if !bytes.Equal(rr.Body.Bytes(), []byte(largeData)) {
		t.Fatalf("data integrity check failed")
	}
}

func TestGzip_ErrorHandling(t *testing.T) {
	// Test error handling when writing buffered data fails
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		// Write data that will exceed max buffer
		w.Write([]byte(strings.Repeat("A", 11<<20))) // 11MB
	}

	// Use a custom config with a small max buffer
	cfg := GzipConfig{MaxBufferBytes: 1 << 20} // 1MB max buffer
	wrapped := middleware.ApplyFunc(handler, GzipWithConfig(cfg))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// Use a custom ResponseWriter that simulates write errors
	rr := httptest.NewRecorder()

	// Wrap the recorder to simulate write errors after a certain amount of data
	errorWriter := &errorSimulatingWriter{
		ResponseWriter: rr,
		failAfter:      1 << 20, // Fail after 1MB
	}

	wrapped(errorWriter, req)

	// The handler should handle the error gracefully
	// We're mainly testing that the error doesn't cause a panic
}

// errorSimulatingWriter simulates write errors for testing
type errorSimulatingWriter struct {
	http.ResponseWriter
	failAfter int
	written   int
}

func (w *errorSimulatingWriter) Write(p []byte) (int, error) {
	if w.written+len(p) > w.failAfter {
		return 0, io.ErrUnexpectedEOF
	}
	w.written += len(p)
	return w.ResponseWriter.Write(p)
}
