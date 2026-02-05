package tenant

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
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
func TenantResolver(options TenantResolverOptions) middleware.Middleware {
	header := headerOrDefault(options.HeaderName, defaultTenantHeader)
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

				writeTenantError(w, r, http.StatusUnauthorized, "tenant_required", "tenant id is required", contract.CategoryAuthentication)
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
