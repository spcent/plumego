package httputil

import (
	"bytes"
	"net/http"
)

// BufferedResponseRecorder captures response data without writing through.
//
// Use this when middleware must inspect or replace the response before any
// bytes are committed to the original ResponseWriter. Use ResponseRecorder when
// callers need write-through recording instead.
type BufferedResponseRecorder struct {
	responseWriter http.ResponseWriter
	statusCode     int
	header         http.Header
	body           *bytes.Buffer
	written        bool
}

// NewBufferedResponseRecorder creates a non-forwarding response recorder.
func NewBufferedResponseRecorder(w http.ResponseWriter) *BufferedResponseRecorder {
	return &BufferedResponseRecorder{
		responseWriter: w,
		statusCode:     http.StatusOK,
		header:         make(http.Header),
		body:           &bytes.Buffer{},
	}
}

// Wrapped returns the original writer, if one was supplied.
func (r *BufferedResponseRecorder) Wrapped() http.ResponseWriter {
	if r == nil {
		return nil
	}
	return r.responseWriter
}

// Header returns the recorded header map.
func (r *BufferedResponseRecorder) Header() http.Header {
	return r.header
}

// WriteHeader records the status code without committing it to the wrapped writer.
func (r *BufferedResponseRecorder) WriteHeader(code int) {
	if r.written {
		return
	}
	r.statusCode = code
	r.written = true
}

// Write records body bytes without writing through to the wrapped writer.
func (r *BufferedResponseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	return r.body.Write(b)
}

// StatusCode returns the recorded status code.
func (r *BufferedResponseRecorder) StatusCode() int {
	if r == nil || r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

// Body returns the recorded response body bytes.
func (r *BufferedResponseRecorder) Body() []byte {
	if r == nil || r.body == nil {
		return nil
	}
	return r.body.Bytes()
}
