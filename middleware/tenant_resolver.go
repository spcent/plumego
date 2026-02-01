package middleware

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/tenant"
)

// TenantResolverOptions configures tenant resolution.
type TenantResolverOptions struct {
	HeaderName       string
	DisablePrincipal bool
	AllowMissing     bool
	Hooks            tenant.Hooks
	OnMissing        func(http.ResponseWriter, *http.Request)
}

// TenantResolver resolves tenant id from request and stores it in context.
func TenantResolver(options TenantResolverOptions) Middleware {
	header := options.HeaderName
	if header == "" {
		header = "X-Tenant-ID"
	}
	requireTenant := !options.AllowMissing
	allowFromPrincipal := !options.DisablePrincipal

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tenantID string
			source := ""

			if allowFromPrincipal {
				if p := contract.PrincipalFromRequest(r); p != nil && p.TenantID != "" {
					tenantID = p.TenantID
					source = "principal"
				}
			}

			if tenantID == "" {
				if headerValue := r.Header.Get(header); headerValue != "" {
					tenantID = headerValue
					source = "header"
				}
			}

			if tenantID == "" && requireTenant {
				if options.OnMissing != nil {
					options.OnMissing(w, r)
					return
				}

				contract.WriteError(w, r, contract.APIError{
					Status:   http.StatusUnauthorized,
					Code:     "tenant_required",
					Message:  "tenant id is required",
					Category: contract.CategoryAuthentication,
				})
				return
			}

			if tenantID != "" {
				r = tenant.RequestWithTenantID(r, tenantID)
				options.Hooks.Resolve(r.Context(), tenant.ResolveInfo{
					TenantID: tenantID,
					Source:   source,
				})
			}

			next.ServeHTTP(w, r)
		})
	}
}
