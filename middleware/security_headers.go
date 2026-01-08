package middleware

import (
	"net/http"

	"github.com/spcent/plumego/security/headers"
)

// SecurityHeaders applies a security header policy to responses.
// When policy is nil, headers.DefaultPolicy() is used.
func SecurityHeaders(policy *headers.Policy) Middleware {
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
