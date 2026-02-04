package security

import (
	"net/http"

	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/security/headers"
)

// SecurityHeaders applies a security header policy to responses.
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
//	handler := security.SecurityHeaders(nil)(myHandler)
//
//	// Or with custom policy
//	policy := &headers.Policy{
//		FrameOptions: "DENY",
//		XSSProtection: "1; mode=block",
//		ContentTypeOptions: "nosniff",
//	}
//	handler := security.SecurityHeaders(policy)(myHandler)
//
// The default policy includes:
//   - X-Frame-Options: DENY (prevents clickjacking)
//   - X-Content-Type-Options: nosniff (prevents MIME sniffing)
//   - X-XSS-Protection: 1; mode=block (enables XSS protection)
//   - Referrer-Policy: strict-origin-when-cross-origin
//   - Content-Security-Policy: default-src 'self'
//
// Note: This middleware should be applied early in the middleware chain
// to ensure security headers are set for all responses.
func SecurityHeaders(policy *headers.Policy) middleware.Middleware {
	effective := headers.DefaultPolicy()
	if policy != nil {
		effective = *policy
		if policy.Additional != nil {
			effective.Additional = make(map[string]string, len(policy.Additional))
			for key, value := range policy.Additional {
				effective.Additional[key] = value
			}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			effective.Apply(w, r)
			next.ServeHTTP(w, r)
		})
	}
}
