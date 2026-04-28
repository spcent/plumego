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

// CopyHeaders replaces destination header values with cloned source values.
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
