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

	for _, part := range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		if ip := validIP(part); ip != "" {
			return ip
		}
	}
	if ip := validIP(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return strings.TrimSpace(host)
	}

	return strings.TrimSpace(r.RemoteAddr)
}

func validIP(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if ip := net.ParseIP(value); ip != nil {
		return value
	}
	return ""
}
