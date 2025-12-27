package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware_CompressesResponse(t *testing.T) {
	const body = "hello world"

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}

	wrapped := ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", rr.Header().Get("Content-Encoding"))
	}

	if rr.Header().Get("Vary") == "" {
		t.Fatalf("expected Vary header to be set")
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(rr.Body.Bytes()))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("failed to read compressed body: %v", err)
	}

	if string(decompressed) != body {
		t.Fatalf("unexpected body: %q", decompressed)
	}
}

func TestGzipMiddleware_PassthroughWhenNotAccepted(t *testing.T) {
	const body = "plain body"

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(body))
	}

	wrapped := ApplyFunc(handler, Gzip())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected no content encoding, got %q", rr.Header().Get("Content-Encoding"))
	}

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rr.Code)
	}

	if rr.Body.String() != body {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}
