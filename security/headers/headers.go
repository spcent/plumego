package headers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/security/input"
)

// HSTSOptions configures Strict-Transport-Security.
//
// HSTS (HTTP Strict Transport Security) forces browsers to use HTTPS for all
// future requests to the domain, preventing SSL stripping attacks.
//
// Example:
//
//	import "github.com/spcent/plumego/security/headers"
//
//	hsts := headers.HSTSOptions{
//		MaxAge:            365 * 24 * time.Hour, // 1 year
//		IncludeSubDomains: true,
//		Preload:           true,
//	}
//	policy := headers.Policy{
//		StrictTransportSecurity: &hsts,
//	}
type HSTSOptions struct {
	// MaxAge is the duration (in seconds) that the browser should remember
	// to only use HTTPS for this domain
	MaxAge time.Duration

	// IncludeSubDomains indicates whether the HSTS policy applies to all subdomains
	IncludeSubDomains bool

	// Preload indicates whether the domain can be included in browser HSTS preload lists
	Preload bool
}

// Policy defines security headers that should be applied to responses.
//
// Policy configures various security headers to protect against common web
// vulnerabilities such as clickjacking, XSS, MIME sniffing, and more.
//
// Example:
//
//	import "github.com/spcent/plumego/security/headers"
//
//	policy := headers.Policy{
//		FrameOptions:          "DENY",
//		ContentTypeOptions:    "nosniff",
//		ReferrerPolicy:        "strict-origin-when-cross-origin",
//		ContentSecurityPolicy: "default-src 'self'",
//	}
//
// The policy can be applied using the Apply method:
//
//	policy.Apply(w, r)
type Policy struct {
	// FrameOptions controls whether the page can be displayed in a frame
	// Values: DENY, SAMEORIGIN
	FrameOptions string

	// ContentTypeOptions prevents MIME sniffing attacks
	// Values: nosniff
	ContentTypeOptions string

	// ReferrerPolicy controls how much referrer information is sent
	// Values: no-referrer, strict-origin-when-cross-origin, etc.
	ReferrerPolicy string

	// PermissionsPolicy controls browser features and capabilities
	// Values: geolocation=(), camera=(), microphone=(), etc.
	PermissionsPolicy string

	// ContentSecurityPolicy controls which resources can be loaded
	// Values: default-src 'self', script-src 'self' 'unsafe-inline', etc.
	ContentSecurityPolicy string

	// CrossOriginOpenerPolicy controls cross-origin opener isolation
	// Values: same-origin, same-origin-allow-popups, unsafe-none
	CrossOriginOpenerPolicy string

	// CrossOriginResourcePolicy controls which resources can be loaded
	// Values: same-site, same-origin, cross-origin
	CrossOriginResourcePolicy string

	// CrossOriginEmbedderPolicy controls cross-origin embedding
	// Values: require-corp, unsafe-none
	CrossOriginEmbedderPolicy string

	// StrictTransportSecurity configures HSTS
	StrictTransportSecurity *HSTSOptions

	// Additional allows custom security headers
	Additional map[string]string
}

// DefaultPolicy returns a baseline header policy with safe defaults.
//
// The default policy includes:
//   - X-Frame-Options: SAMEORIGIN (prevents clickjacking)
//   - X-Content-Type-Options: nosniff (prevents MIME sniffing)
//   - Referrer-Policy: strict-origin-when-cross-origin
//
// Example:
//
//	import "github.com/spcent/plumego/security/headers"
//
//	policy := headers.DefaultPolicy()
//	policy.Apply(w, r)
func DefaultPolicy() Policy {
	return Policy{
		FrameOptions:       "SAMEORIGIN",
		ContentTypeOptions: "nosniff",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}
}

// StrictPolicy returns a hardened policy suitable for locked-down deployments.
//
// The strict policy includes:
//   - X-Frame-Options: DENY (prevents all framing)
//   - X-Content-Type-Options: nosniff
//   - Referrer-Policy: no-referrer
//   - Permissions-Policy: geolocation=(), camera=(), microphone=()
//   - Cross-Origin-Opener-Policy: same-origin
//   - Cross-Origin-Resource-Policy: same-origin
//   - Cross-Origin-Embedder-Policy: require-corp
//   - Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
//
// Example:
//
//	import "github.com/spcent/plumego/security/headers"
//
//	policy := headers.StrictPolicy()
//	policy.Apply(w, r)
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
//
// Example:
//
//	import "github.com/spcent/plumego/security/headers"
//
//	policy := headers.DefaultPolicy()
//	policy.Apply(w, r)
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

	// Check X-Forwarded-Proto header - validate entire proxy chain
	protoHeader := r.Header.Get("X-Forwarded-Proto")
	if protoHeader != "" {
		// Split by comma to get all proxy values
		proxies := strings.Split(protoHeader, ",")
		for _, proxy := range proxies {
			proto := strings.TrimSpace(proxy)
			if strings.EqualFold(proto, "https") {
				return true
			}
			// If we encounter http in the chain, it's not secure
			if strings.EqualFold(proto, "http") {
				return false
			}
		}
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
