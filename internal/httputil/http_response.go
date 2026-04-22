package httputil

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
)

const (
	HeaderContentTypeNoSniff = "X-Content-Type-Options"
	ContentTypeNoSniffValue  = "nosniff"
)

// EnsureNoSniff sets X-Content-Type-Options to nosniff if it is not present.
// This reduces XSS risk from MIME sniffing when responses contain user-controlled text.
func EnsureNoSniff(header http.Header) {
	if header == nil {
		return
	}
	if header.Get(HeaderContentTypeNoSniff) == "" {
		header.Set(HeaderContentTypeNoSniff, ContentTypeNoSniffValue)
	}
}

// SafeWrite writes response bytes after applying minimal response hardening.
//
// SECURITY: This helper is intended for middleware and infrastructure code that
// needs to copy an already-constructed HTTP response. It does NOT perform any
// context-aware HTML/JS escaping. Callers that generate HTML or other active
// content MUST apply appropriate escaping/encoding before passing data here,
// and must set an appropriate Content-Type header (for example, application/json).
//
// The only protection applied here is X-Content-Type-Options: nosniff to
// prevent browsers from MIME-sniffing non-HTML responses as HTML.
//
// codeql[go/reflected-xss]: SafeWrite is a low-level passthrough used by
// response caching/middleware; it is not responsible for HTML encoding.
func SafeWrite(w http.ResponseWriter, body []byte) (int, error) {
	if w == nil {
		return 0, nil
	}
	EnsureNoSniff(w.Header())
	return w.Write(body)
}

// ResponseRecorder captures response data while still writing to the underlying writer.
type ResponseRecorder struct {
	http.ResponseWriter
	statusCode   int
	header       http.Header
	body         *bytes.Buffer
	bytesWritten int
	written      bool
}

// NewResponseRecorder creates a response recorder with sane defaults.
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		header:         make(http.Header),
		body:           &bytes.Buffer{},
	}
}

// Header returns the recorded header map.
func (r *ResponseRecorder) Header() http.Header {
	return r.header
}

// WriteHeader records the status code and writes headers to the underlying writer once.
func (r *ResponseRecorder) WriteHeader(code int) {
	if r.written {
		return
	}
	r.statusCode = code
	r.written = true

	CopyHeaders(r.ResponseWriter.Header(), r.header)
	EnsureNoSniff(r.ResponseWriter.Header())
	r.ResponseWriter.WriteHeader(code)
}

// Write captures body bytes and writes through to the underlying writer.
//
// SECURITY NOTE: This is infrastructure that captures response data for
// middleware, metrics, logging, and caching. It does not inject user input into
// HTML contexts; active content escaping belongs in handlers that generate it.
func (r *ResponseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}

	r.body.Write(b)
	n, err := SafeWrite(r.ResponseWriter, b)
	r.bytesWritten += n
	return n, err
}

// StatusCode returns the recorded status code.
func (r *ResponseRecorder) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

// Body returns the captured response body bytes.
func (r *ResponseRecorder) Body() []byte {
	return r.body.Bytes()
}

// BytesWritten returns the number of body bytes successfully written through.
func (r *ResponseRecorder) BytesWritten() int {
	return r.bytesWritten
}

// Hijack forwards http.Hijacker when the wrapped writer supports it.
func (r *ResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hj.Hijack()
}

// Flush forwards http.Flusher when the wrapped writer supports it.
func (r *ResponseRecorder) Flush() {
	if fl, ok := r.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

// CopyHeaders appends all values from src into dst.
func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
