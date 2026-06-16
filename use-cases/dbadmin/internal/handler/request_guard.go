package handler

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// SameOriginMiddleware rejects unsafe browser requests that carry a foreign
// Origin or Referer. Requests without those headers are allowed so local CLI
// automation and health checks keep working.
func SameOriginMiddleware(logger plumelog.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) || sameOrigin(r) {
				next.ServeHTTP(w, r)
				return
			}
			logger.Warn("same-origin check failed", plumelog.Fields{"method": r.Method, "path": r.URL.Path})
			logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeForbidden).
				Message("cross-origin request rejected").
				Build()))
		})
	}
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func sameOrigin(r *http.Request) bool {
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		return originHostMatches(origin, r.Host)
	}
	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		return originHostMatches(referer, r.Host)
	}
	return true
}

func originHostMatches(rawURL, host string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false
	}
	return strings.EqualFold(u.Host, host)
}

// IPAllowlistMiddleware rejects requests from client IPs outside allowedIPs.
// An empty allowlist disables the check entirely.
func IPAllowlistMiddleware(allowedIPs []string, logger plumelog.StructuredLogger) func(http.Handler) http.Handler {
	if len(allowedIPs) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	nets := make([]*net.IPNet, 0, len(allowedIPs))
	for _, cidr := range allowedIPs {
		c := cidr
		if !strings.Contains(c, "/") {
			if strings.Contains(c, ":") {
				c = c + "/128"
			} else {
				c = c + "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(c)
		if err != nil {
			logger.Warn("invalid IP in allowlist", plumelog.Fields{"ip": cidr})
			continue
		}
		nets = append(nets, ipNet)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realClientIP(r)
			if !ipAllowed(ip, nets) {
				logger.Warn("IP allowlist blocked", plumelog.Fields{"ip": ip})
				logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeForbidden).
					Message("access denied").Build()))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func realClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func ipAllowed(ipStr string, nets []*net.IPNet) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
