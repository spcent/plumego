package handler

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"
	"github.com/spcent/plumego/security/jwt"
	"mini-saas-api/internal/domain/access"
	"mini-saas-api/internal/domain/session"
)

// TokenVerifier verifies an access token. *jwt.JWTManager satisfies this.
type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string, tokenType jwt.TokenType) (*jwt.TokenClaims, error)
}

// RequireAuth is a per-route guard: it verifies the bearer token, lifts the
// tenant ID and role from the claims onto an authn.Principal, and installs the
// principal in the request context. Handlers downstream read it back with
// authn.PrincipalFromContext. Fail closed: any verification problem is 401.
func RequireAuth(verifier TokenVerifier, logger plumelog.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := authn.ExtractBearerToken(r)
			if token == "" {
				writeUnauthorized(w, r, logger, "auth.missing_token", "missing bearer token")
				return
			}
			claims, err := verifier.VerifyToken(r.Context(), token, jwt.TokenTypeAccess)
			if err != nil {
				writeUnauthorized(w, r, logger, "auth.invalid_token", "invalid or expired access token")
				return
			}
			principal := jwt.PrincipalFromClaims(claims)
			principal.TenantID = session.TenantFromClaims(claims)
			if principal.Subject == "" || principal.TenantID == "" {
				writeUnauthorized(w, r, logger, "auth.invalid_token", "token is missing subject or tenant")
				return
			}
			next.ServeHTTP(w, r.WithContext(authn.WithPrincipal(r.Context(), principal)))
		})
	}
}

// RequireRole is a per-route guard enforcing a minimum membership role.
// It must run inside RequireAuth. The role is the one captured at token issue
// time; role changes take effect when the client refreshes its access token.
func RequireRole(min access.Role, logger plumelog.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := authn.PrincipalFromContext(r.Context())
			if p == nil || len(p.Roles) == 0 {
				writeUnauthorized(w, r, logger, "auth.missing_principal", "authentication required")
				return
			}
			if !access.Role(p.Roles[0]).AtLeast(min) {
				logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeForbidden).
					Code("auth.insufficient_role").
					Message("this action requires the "+string(min)+" role").
					Build()))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request, logger plumelog.StructuredLogger, code, msg string) {
	logWriteErr(logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeUnauthorized).
		Code(code).
		Message(msg).
		Build()))
}
