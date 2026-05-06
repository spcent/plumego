package transport

import (
	"net"
	"net/http"
	"strings"

	httputil "github.com/spcent/plumego/internal/httputil"
)

const (
	HeaderForwardedFor       = "X-Forwarded-For"
	HeaderRealIP             = "X-Real-IP"
	HeaderContentTypeNoSniff = httputil.HeaderContentTypeNoSniff
	ContentTypeNoSniffValue  = httputil.ContentTypeNoSniffValue
)

func EnsureNoSniff(header http.Header) {
	httputil.EnsureNoSniff(header)
}

func SafeWrite(w http.ResponseWriter, body []byte) (int, error) {
	return httputil.SafeWrite(w, body)
}

// AddVary appends Vary header tokens without duplicating existing values.
func AddVary(header http.Header, values ...string) {
	if header == nil {
		return
	}

	existing := map[string]struct{}{}
	for _, value := range header.Values("Vary") {
		for _, token := range strings.Split(value, ",") {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}
			existing[strings.ToLower(token)] = struct{}{}
		}
	}

	for _, value := range values {
		for _, token := range strings.Split(value, ",") {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}
			key := strings.ToLower(token)
			if _, ok := existing[key]; ok {
				continue
			}
			header.Add("Vary", token)
			existing[key] = struct{}{}
		}
	}
}

// CopyHeaders overlays cloned source values onto destination headers.
// Destination keys absent from src are preserved.
func CopyHeaders(dst, src http.Header) {
	if dst == nil || src == nil {
		return
	}
	for key, values := range src {
		cloned := make([]string, len(values))
		copy(cloned, values)
		dst[key] = cloned
	}
}

// ReplaceHeaders replaces the complete destination header map with cloned
// source values. Destination keys absent from src are removed.
func ReplaceHeaders(dst, src http.Header) {
	if dst == nil {
		return
	}
	for key := range dst {
		delete(dst, key)
	}
	CopyHeaders(dst, src)
}

func ClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	if ip := strings.TrimSpace(strings.Split(r.Header.Get(HeaderForwardedFor), ",")[0]); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(r.Header.Get(HeaderRealIP)); ip != "" {
		return ip
	}

	return DirectClientIP(r)
}

// DirectClientIP extracts the peer IP from RemoteAddr only.
//
// Use this for security-sensitive defaults such as rate limiting when the
// application has not explicitly configured trusted proxy handling.
func DirectClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}

type ResponseRecorder = httputil.ResponseRecorder

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return httputil.NewResponseRecorder(w)
}
