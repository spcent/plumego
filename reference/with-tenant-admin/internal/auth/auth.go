package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
)

const HeaderAdminToken = "X-Admin-Token"

func RequireAdminToken(adminToken string) func(http.Handler) http.Handler {
	expected := sha256.Sum256([]byte(strings.TrimSpace(adminToken)))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := sha256.Sum256([]byte(strings.TrimSpace(r.Header.Get(HeaderAdminToken))))
			if !hmac.Equal(got[:], expected[:]) {
				_ = contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeUnauthorized).
					Code(contract.CodeUnauthorized).
					Message("admin token is required").
					Build())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
