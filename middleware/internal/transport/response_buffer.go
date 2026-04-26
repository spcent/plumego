package transport

import (
	"net/http"
)

// BufferedResponse captures headers, status, and body in memory.
// It does not write to an underlying ResponseWriter.
type BufferedResponse struct {
	header       http.Header
	statusCode   int
	body         []byte
	bytesWritten int
	wroteHeader  bool
	maxBytes     int
	overflow     bool
}

// NewBufferedResponse creates a buffered response with an optional max size.
// A maxBytes of 0 disables the size limit.
func NewBufferedResponse(maxBytes int) *BufferedResponse {
	return &BufferedResponse{
		header:   make(http.Header),
		maxBytes: maxBytes,
	}
}

func (b *BufferedResponse) Header() http.Header {
	return b.header
}

func (b *BufferedResponse) WriteHeader(statusCode int) {
	if b.wroteHeader {
		return
	}
	b.statusCode = statusCode
	b.wroteHeader = true
}

func (b *BufferedResponse) Write(p []byte) (int, error) {
	if !b.wroteHeader {
		b.WriteHeader(http.StatusOK)
	}
	if b.maxBytes > 0 && len(b.body)+len(p) > b.maxBytes {
		b.overflow = true
		return 0, http.ErrBodyNotAllowed
	}
	b.body = append(b.body, p...)
	b.bytesWritten += len(p)
	return len(p), nil
}

func (b *BufferedResponse) StatusCode() int {
	if b.statusCode == 0 {
		return http.StatusOK
	}
	return b.statusCode
}

func (b *BufferedResponse) Body() []byte {
	return b.body
}

func (b *BufferedResponse) BytesWritten() int {
	return b.bytesWritten
}

func (b *BufferedResponse) Overflowed() bool {
	return b.overflow
}

func (b *BufferedResponse) Len() int {
	return len(b.body)
}

func (b *BufferedResponse) ClearBody() {
	b.body = nil
	b.bytesWritten = 0
}

// WriteTo writes the buffered response to the destination writer.
func (b *BufferedResponse) WriteTo(dst http.ResponseWriter) (int, error) {
	if dst == nil {
		return 0, nil
	}
	copyBufferedHeaders(dst.Header(), b.header)
	EnsureNoSniff(dst.Header())
	dst.WriteHeader(b.StatusCode())
	return SafeWrite(dst, b.body)
}

func copyBufferedHeaders(dst, src http.Header) {
	for key, values := range src {
		cloned := make([]string, len(values))
		copy(cloned, values)
		dst[key] = cloned
	}
}
