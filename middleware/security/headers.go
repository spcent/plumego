// Package security provides the HTTP security-header middleware. It adapts the
// policy types in [security/headers] into the [middleware.Middleware] signature
// so callers can wire a security-header policy into the standard middleware
// chain without importing security/headers directly.
//
// This package and the top-level security/ package are distinct:
//   - security/ owns auth primitives, JWT helpers, input safety, and abuse guards.
//   - middleware/security (this package) owns only the HTTP response-header middleware adapter.
package security

import (
	"net/http"

	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/security/headers"
)

// Config controls security-header middleware behavior.
type Config struct {
	Policy *headers.Policy
}

// Middleware applies a security header policy to responses.
//
// This middleware adds security-related HTTP headers to responses to protect
// against common web vulnerabilities such as XSS, clickjacking, and MIME sniffing.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/security"
//	import "github.com/spcent/plumego/security/headers"
//
//	// Use default security policy
//	mw, err := security.Middleware(security.Config{})
//
//	// Or with custom policy
//	policy := &headers.Policy{
//		FrameOptions: "DENY",
//		ContentTypeOptions: "nosniff",
//		ContentSecurityPolicy: headers.StrictCSP(),
//	}
//	mw, err = security.Middleware(security.Config{Policy: policy})
//	_ = err
//	handler := mw(myHandler)
//
// The default policy includes:
//   - X-Frame-Options: SAMEORIGIN (prevents clickjacking)
//   - X-Content-Type-Options: nosniff (prevents MIME sniffing)
//   - Referrer-Policy: strict-origin-when-cross-origin
//
// For stronger XSS protection, set Content-Security-Policy explicitly,
// e.g. use headers.StrictCSP() or headers.StrictPolicy().
//
// Invalid custom policies fail during construction.
//
// Note: This middleware should be applied early in the middleware chain to
// ensure security headers are set for all responses.
func Middleware(config Config) (middleware.Middleware, error) {
	effective := headers.DefaultPolicy()
	if config.Policy != nil {
		effective = *config.Policy
		if config.Policy.Additional != nil {
			effective.Additional = make(map[string]string, len(config.Policy.Additional))
			for key, value := range config.Policy.Additional {
				effective.Additional[key] = value
			}
		}
		if err := effective.Validate(); err != nil {
			return nil, err
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			effective.Apply(w, r)
			next.ServeHTTP(w, r)
		})
	}, nil
}
