package handler

import (
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
