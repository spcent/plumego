package headers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/security/input"
)

// HSTSOptions configures Strict-Transport-Security.
type HSTSOptions struct {
	MaxAge            time.Duration
	IncludeSubDomains bool
	Preload           bool
}

// Policy defines security headers that should be applied to responses.
type Policy struct {
	FrameOptions              string
	ContentTypeOptions        string
	ReferrerPolicy            string
	PermissionsPolicy         string
	ContentSecurityPolicy     string
	CrossOriginOpenerPolicy   string
	CrossOriginResourcePolicy string
	CrossOriginEmbedderPolicy string
	StrictTransportSecurity   *HSTSOptions
	Additional                map[string]string
}

// DefaultPolicy returns a baseline header policy with safe defaults.
func DefaultPolicy() Policy {
	return Policy{
		FrameOptions:       "SAMEORIGIN",
		ContentTypeOptions: "nosniff",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}
}

// StrictPolicy returns a hardened policy suitable for locked-down deployments.
func StrictPolicy() Policy {
	return Policy{
		FrameOptions:              "DENY",
		ContentTypeOptions:        "nosniff",
		ReferrerPolicy:            "no-referrer",
		PermissionsPolicy:         "geolocation=(), camera=(), microphone=()",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
		CrossOriginEmbedderPolicy: "require-corp",
		StrictTransportSecurity: &HSTSOptions{
			MaxAge:            365 * 24 * time.Hour,
			IncludeSubDomains: true,
			Preload:           true,
		},
	}
}

// Apply attaches the configured headers to the response.
func (p Policy) Apply(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	setHeader(headers, "X-Frame-Options", p.FrameOptions)
	setHeader(headers, "X-Content-Type-Options", p.ContentTypeOptions)
	setHeader(headers, "Referrer-Policy", p.ReferrerPolicy)
	setHeader(headers, "Permissions-Policy", p.PermissionsPolicy)
	setHeader(headers, "Content-Security-Policy", p.ContentSecurityPolicy)
	setHeader(headers, "Cross-Origin-Opener-Policy", p.CrossOriginOpenerPolicy)
	setHeader(headers, "Cross-Origin-Resource-Policy", p.CrossOriginResourcePolicy)
	setHeader(headers, "Cross-Origin-Embedder-Policy", p.CrossOriginEmbedderPolicy)

	if p.StrictTransportSecurity != nil && p.StrictTransportSecurity.MaxAge > 0 && isHTTPSRequest(r) {
		setHeader(headers, "Strict-Transport-Security", formatHSTS(*p.StrictTransportSecurity))
	}

	for name, value := range p.Additional {
		if !input.IsHeaderName(name) {
			continue
		}
		setHeader(headers, http.CanonicalHeaderKey(name), value)
	}
}

func setHeader(headers http.Header, name, value string) {
	if value == "" {
		return
	}
	if !input.IsHeaderValue(value) {
		return
	}
	headers.Set(name, value)
}

func formatHSTS(opts HSTSOptions) string {
	maxAge := int64(opts.MaxAge.Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	parts := []string{fmt.Sprintf("max-age=%d", maxAge)}
	if opts.IncludeSubDomains {
		parts = append(parts, "includeSubDomains")
	}
	if opts.Preload {
		parts = append(parts, "preload")
	}
	return strings.Join(parts, "; ")
}

func isHTTPSRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.TLS != nil {
		return true
	}

	proto := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0])
	if strings.EqualFold(proto, "https") {
		return true
	}

	if strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Ssl")), "on") {
		return true
	}

	forwarded := strings.ToLower(r.Header.Get("Forwarded"))
	if strings.Contains(forwarded, "proto=https") {
		return true
	}

	return false
}
