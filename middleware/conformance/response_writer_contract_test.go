package conformance_test

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/coalesce"
	"github.com/spcent/plumego/middleware/compression"
	"github.com/spcent/plumego/middleware/timeout"
)

func TestResponseWriterConformancePanicPropagation(t *testing.T) {
	tests := []struct {
		name string
		mw   middleware.Middleware
	}{
		{name: "bodylimit", mw: bodylimit.BodyLimit(1024, nil)},
		{name: "coalesce", mw: coalesce.Middleware(coalesce.Config{Timeout: time.Second})},
		{name: "compression", mw: compression.Gzip(compression.GzipConfig{})},
		{name: "timeout", mw: timeout.Timeout(timeout.TimeoutConfig{Timeout: time.Second})},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("conformance panic")
			}))
			req := httptest.NewRequest(http.MethodGet, "/panic", nil)
			req.Header.Set("Accept-Encoding", "gzip")

			defer func() {
				if rec := recover(); rec != "conformance panic" {
					t.Fatalf("panic = %v, want conformance panic", rec)
				}
			}()
			handler.ServeHTTP(httptest.NewRecorder(), req)
		})
	}
}

func TestResponseWriterConformanceFlushForwarding(t *testing.T) {
	tests := []struct {
		name string
		mw   middleware.Middleware
	}{
		{name: "bodylimit", mw: bodylimit.BodyLimit(1024, nil)},
		{name: "coalesce", mw: coalesce.Middleware(coalesce.Config{Timeout: time.Second})},
		{name: "compression", mw: compression.Gzip(compression.GzipConfig{})},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writer := &conformanceFlushWriter{ResponseRecorder: httptest.NewRecorder()}
			handler := tc.mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatalf("%s wrapper does not expose http.Flusher", tc.name)
				}
				flusher.Flush()
			}))
			req := httptest.NewRequest(http.MethodGet, "/flush", nil)
			req.Header.Set("Accept-Encoding", "gzip")

			handler.ServeHTTP(writer, req)

			if writer.flushed == 0 {
				t.Fatalf("%s wrapper did not forward Flush", tc.name)
			}
		})
	}
}

func TestResponseWriterConformanceGzipHijackBeforeCompression(t *testing.T) {
	writer := &conformanceHijackWriter{}
	handler := compression.Gzip(compression.GzipConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("gzip wrapper does not expose http.Hijacker before compression")
		}
		if _, _, err := hijacker.Hijack(); err != nil {
			t.Fatalf("Hijack returned error: %v", err)
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "/hijack", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	handler.ServeHTTP(writer, req)

	if !writer.hijacked {
		t.Fatal("underlying writer was not hijacked")
	}
}

func TestResponseWriterConformancePostTimeoutWriteReturnsDeadline(t *testing.T) {
	type writeResult struct {
		n   int
		err error
	}
	result := make(chan writeResult, 1)

	handler := timeout.Timeout(timeout.TimeoutConfig{Timeout: 10 * time.Millisecond})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(40 * time.Millisecond)
		n, err := w.Write([]byte("late"))
		result <- writeResult{n: n, err: err}
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/timeout", nil))

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusGatewayTimeout)
	}
	select {
	case got := <-result:
		if got.n != 0 {
			t.Fatalf("post-timeout write bytes = %d, want 0", got.n)
		}
		if !errors.Is(got.err, context.DeadlineExceeded) {
			t.Fatalf("post-timeout write error = %v, want context deadline exceeded", got.err)
		}
	case <-time.After(time.Second):
		t.Fatal("handler did not attempt post-timeout write")
	}
}

func TestResponseWriterConformanceGzipPartialPanicFinalizesStream(t *testing.T) {
	handler := compression.Gzip(compression.GzipConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("partial"))
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/partial", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		handler.ServeHTTP(rec, req)
	}()

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", rec.Header().Get("Content-Encoding"))
	}
	if got := conformanceGunzip(t, rec.Body.Bytes()); got != "partial" {
		t.Fatalf("decompressed body = %q, want partial", got)
	}
}

func TestResponseWriterConformanceGzipFlushBeforeWritePassesThrough(t *testing.T) {
	handler := compression.Gzip(compression.GzipConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("gzip wrapper does not expose http.Flusher")
		}
		flusher.Flush()
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("after flush"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/flush-write", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !rec.Flushed {
		t.Fatal("underlying recorder was not flushed")
	}
	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("Content-Encoding = %q, want empty", rec.Header().Get("Content-Encoding"))
	}
	if rec.Body.String() != "after flush" {
		t.Fatalf("body = %q, want after flush", rec.Body.String())
	}
}

type conformanceFlushWriter struct {
	*httptest.ResponseRecorder
	flushed int
}

func (w *conformanceFlushWriter) Flush() {
	w.flushed++
	w.ResponseRecorder.Flush()
}

type conformanceHijackWriter struct {
	header   http.Header
	hijacked bool
}

func (w *conformanceHijackWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *conformanceHijackWriter) WriteHeader(int) {}

func (w *conformanceHijackWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (w *conformanceHijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	var rw bytes.Buffer
	return nil, bufio.NewReadWriter(bufio.NewReader(&rw), bufio.NewWriter(&rw)), nil
}

func conformanceGunzip(t *testing.T, body []byte) string {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("open gzip reader: %v", err)
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read gzip body: %v", err)
	}
	return string(decoded)
}
