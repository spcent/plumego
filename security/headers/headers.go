// Package headers provides HTTP security header management and middleware.
//
// This package implements a comprehensive security header policy system supporting:
//   - HSTS (HTTP Strict Transport Security)
//   - CSP (Content Security Policy)
//   - X-Frame-Options
//   - X-Content-Type-Options
//   - Referrer-Policy
//   - Permissions-Policy
//
// Features:
//   - Declarative policy configuration
//   - HTTP middleware for automatic header injection
//   - CSP nonce generation for inline scripts
//   - Policy validation and error reporting
//
// Example usage:
//
//	import "github.com/spcent/plumego/security/headers"
//
//	policy := headers.Policy{
//		StrictTransportSecurity: &headers.HSTSOptions{
//			MaxAge:            365 * 24 * time.Hour,
//			IncludeSubDomains: true,
//			Preload:           true,
//		},
//		ContentSecurityPolicy: &headers.CSPOptions{
//			DefaultSrc: []string{"'self'"},
//			ScriptSrc:  []string{"'self'", "'unsafe-inline'"},
//		},
//		XFrameOptions: "DENY",
//	}
//
//	// Apply as middleware
//	middleware := policy.Middleware()
//	app.Use(middleware)
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

// Middleware returns an HTTP middleware that applies the security policy to all responses.
//
// Example:
//
//	import (
//		"github.com/spcent/plumego/security/headers"
//		"net/http"
//	)
//
//	policy := headers.DefaultPolicy()
//	handler := policy.Middleware(http.HandlerFunc(myHandler))
func (p Policy) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.Apply(w, r)
		next.ServeHTTP(w, r)
	})
}

// CSPBuilder helps construct Content-Security-Policy headers.
//
// Example:
//
//	import "github.com/spcent/plumego/security/headers"
//
//	csp := headers.NewCSPBuilder().
//		DefaultSrc("'self'").
//		ScriptSrc("'self'", "'unsafe-inline'", "https://cdn.example.com").
//		StyleSrc("'self'", "'unsafe-inline'").
//		ImgSrc("'self'", "data:", "https:").
//		Build()
type CSPBuilder struct {
	directives map[string][]string
}

// NewCSPBuilder creates a new CSP builder.
func NewCSPBuilder() *CSPBuilder {
	return &CSPBuilder{
		directives: make(map[string][]string),
	}
}

// DefaultSrc sets the default-src directive.
func (b *CSPBuilder) DefaultSrc(sources ...string) *CSPBuilder {
	b.directives["default-src"] = sources
	return b
}

// ScriptSrc sets the script-src directive.
func (b *CSPBuilder) ScriptSrc(sources ...string) *CSPBuilder {
	b.directives["script-src"] = sources
	return b
}

// StyleSrc sets the style-src directive.
func (b *CSPBuilder) StyleSrc(sources ...string) *CSPBuilder {
	b.directives["style-src"] = sources
	return b
}

// ImgSrc sets the img-src directive.
func (b *CSPBuilder) ImgSrc(sources ...string) *CSPBuilder {
	b.directives["img-src"] = sources
	return b
}

// FontSrc sets the font-src directive.
func (b *CSPBuilder) FontSrc(sources ...string) *CSPBuilder {
	b.directives["font-src"] = sources
	return b
}

// ConnectSrc sets the connect-src directive.
func (b *CSPBuilder) ConnectSrc(sources ...string) *CSPBuilder {
	b.directives["connect-src"] = sources
	return b
}

// FrameSrc sets the frame-src directive.
func (b *CSPBuilder) FrameSrc(sources ...string) *CSPBuilder {
	b.directives["frame-src"] = sources
	return b
}

// ObjectSrc sets the object-src directive.
func (b *CSPBuilder) ObjectSrc(sources ...string) *CSPBuilder {
	b.directives["object-src"] = sources
	return b
}

// MediaSrc sets the media-src directive.
func (b *CSPBuilder) MediaSrc(sources ...string) *CSPBuilder {
	b.directives["media-src"] = sources
	return b
}

// ChildSrc sets the child-src directive.
func (b *CSPBuilder) ChildSrc(sources ...string) *CSPBuilder {
	b.directives["child-src"] = sources
	return b
}

// FormAction sets the form-action directive.
func (b *CSPBuilder) FormAction(sources ...string) *CSPBuilder {
	b.directives["form-action"] = sources
	return b
}

// FrameAncestors sets the frame-ancestors directive.
func (b *CSPBuilder) FrameAncestors(sources ...string) *CSPBuilder {
	b.directives["frame-ancestors"] = sources
	return b
}

// BaseURI sets the base-uri directive.
func (b *CSPBuilder) BaseURI(sources ...string) *CSPBuilder {
	b.directives["base-uri"] = sources
	return b
}

// ManifestSrc sets the manifest-src directive.
func (b *CSPBuilder) ManifestSrc(sources ...string) *CSPBuilder {
	b.directives["manifest-src"] = sources
	return b
}

// WorkerSrc sets the worker-src directive.
func (b *CSPBuilder) WorkerSrc(sources ...string) *CSPBuilder {
	b.directives["worker-src"] = sources
	return b
}

// ReportURI sets the report-uri directive.
func (b *CSPBuilder) ReportURI(uri string) *CSPBuilder {
	b.directives["report-uri"] = []string{uri}
	return b
}

// ReportTo sets the report-to directive.
func (b *CSPBuilder) ReportTo(group string) *CSPBuilder {
	b.directives["report-to"] = []string{group}
	return b
}

// UpgradeInsecureRequests adds the upgrade-insecure-requests directive.
func (b *CSPBuilder) UpgradeInsecureRequests() *CSPBuilder {
	b.directives["upgrade-insecure-requests"] = []string{}
	return b
}

// BlockAllMixedContent adds the block-all-mixed-content directive.
func (b *CSPBuilder) BlockAllMixedContent() *CSPBuilder {
	b.directives["block-all-mixed-content"] = []string{}
	return b
}

// Sandbox adds the sandbox directive.
func (b *CSPBuilder) Sandbox(values ...string) *CSPBuilder {
	b.directives["sandbox"] = values
	return b
}

// Build constructs the CSP header value.
func (b *CSPBuilder) Build() string {
	if len(b.directives) == 0 {
		return ""
	}

	parts := make([]string, 0, len(b.directives))

	// Maintain consistent order for testing
	directiveOrder := []string{
		"default-src", "script-src", "style-src", "img-src", "font-src",
		"connect-src", "frame-src", "object-src", "media-src", "child-src",
		"form-action", "frame-ancestors", "base-uri", "manifest-src", "worker-src",
		"report-uri", "report-to", "upgrade-insecure-requests", "block-all-mixed-content",
		"sandbox",
	}

	for _, directive := range directiveOrder {
		sources, ok := b.directives[directive]
		if !ok {
			continue
		}

		if len(sources) == 0 {
			parts = append(parts, directive)
		} else {
			parts = append(parts, directive+" "+strings.Join(sources, " "))
		}
	}

	return strings.Join(parts, "; ")
}

// StrictCSP returns a strict CSP policy suitable for modern applications.
//
// Example:
//
//	import "github.com/spcent/plumego/security/headers"
//
//	csp := headers.StrictCSP()
func StrictCSP() string {
	return NewCSPBuilder().
		DefaultSrc("'self'").
		ScriptSrc("'self'").
		StyleSrc("'self'").
		ImgSrc("'self'", "data:", "https:").
		FontSrc("'self'").
		ConnectSrc("'self'").
		FrameSrc("'none'").
		ObjectSrc("'none'").
		BaseURI("'self'").
		FormAction("'self'").
		FrameAncestors("'none'").
		UpgradeInsecureRequests().
		Build()
}
