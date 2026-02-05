package http

import (
	"bytes"
	"net/http"
)

// ResponseRecorder captures response data while still writing to the underlying writer.
type ResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	header     http.Header
	body       *bytes.Buffer
	written    bool
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

	copyHeaders(r.ResponseWriter.Header(), r.header)
	r.ResponseWriter.WriteHeader(code)
}

// Write captures body bytes and writes through to the underlying writer.
func (r *ResponseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}

	r.body.Write(b)
	return r.ResponseWriter.Write(b)
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

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
