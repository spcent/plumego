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
