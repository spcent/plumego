package csrf

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

const headerName = "X-Requested-With"

// Middleware rejects state-changing requests (POST/PUT/PATCH/DELETE) that do not
// carry the X-Requested-With header. This defends against cross-site form submissions
// while being transparent to fetch/XHR clients that set this header automatically.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			if r.Header.Get(headerName) == "" {
				_ = contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeForbidden).
					Message("missing X-Requested-With header").
					Build())
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
