package httpx

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP returns the best-effort client IP for a request.
// It checks X-Forwarded-For, X-Real-IP, then RemoteAddr.
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
