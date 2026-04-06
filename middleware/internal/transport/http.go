package transport

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"strings"
)

func EnsureNoSniff(header http.Header) {
	if header == nil {
		return
	}
	if header.Get("X-Content-Type-Options") == "" {
		header.Set("X-Content-Type-Options", "nosniff")
	}
}

func SafeWrite(w http.ResponseWriter, body []byte) (int, error) {
	if w == nil {
		return 0, nil
	}
	EnsureNoSniff(w.Header())
	return w.Write(body)
}

func ClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	if ip := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}

type ResponseRecorder struct {
	http.ResponseWriter
	statusCode   int
	header       http.Header
	body         *bytes.Buffer
	bytesWritten int
	written      bool
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		header:         make(http.Header),
		body:           &bytes.Buffer{},
	}
}

func (r *ResponseRecorder) Header() http.Header {
	return r.header
}

func (r *ResponseRecorder) WriteHeader(code int) {
	if r.written {
		return
	}
	r.statusCode = code
	r.written = true

	copyHeaders(r.ResponseWriter.Header(), r.header)
	EnsureNoSniff(r.ResponseWriter.Header())
	r.ResponseWriter.WriteHeader(code)
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}

	r.body.Write(b)
	n, err := SafeWrite(r.ResponseWriter, b)
	r.bytesWritten += n
	return n, err
}

func (r *ResponseRecorder) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

func (r *ResponseRecorder) Body() []byte {
	return r.body.Bytes()
}

func (r *ResponseRecorder) BytesWritten() int {
	return r.bytesWritten
}

func (r *ResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hj.Hijack()
}

func (r *ResponseRecorder) Flush() {
	if fl, ok := r.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
