package tenant

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/tenant"
)

// TenantResolverOptions configures tenant resolution.
type TenantResolverOptions struct {
	// HeaderName is the HTTP header to extract tenant ID from (default: "X-Tenant-ID").
	HeaderName string
	// Extractor is a custom function to extract tenant ID from the request.
	// When set, takes precedence over HeaderName. Principal extraction still runs first.
	Extractor tenant.TenantExtractor
	// DisablePrincipal disables extracting tenant ID from the authenticated Principal.
	DisablePrincipal bool
	// AllowMissing allows requests to proceed when no tenant ID is found.
	AllowMissing bool
	// Hooks provides callbacks for tenant resolution events.
	Hooks tenant.Hooks
	// OnMissing is called when tenant ID is missing and AllowMissing is false.
	// If nil, a standard 401 JSON error is returned.
	OnMissing func(http.ResponseWriter, *http.Request)
}

// TenantResolver resolves tenant id from request and stores it in context.
// Resolution order: Principal → custom Extractor or Header.
// The resolved tenant ID is validated before being stored in context.
func TenantResolver(options TenantResolverOptions) middleware.Middleware {
	header := headerOrDefault(options.HeaderName, defaultTenantHeader)
	requireTenant := !options.AllowMissing
	allowFromPrincipal := !options.DisablePrincipal

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tenantID string
			source := ""

			// 1. Try to extract from authenticated Principal (JWT/session).
			if allowFromPrincipal {
				if p := contract.PrincipalFromRequest(r); p != nil && p.TenantID != "" {
					tenantID = p.TenantID
					source = "principal"
				}
			}

			// 2. Fall back to custom extractor or header.
			if tenantID == "" {
				if options.Extractor != nil {
					if id, err := options.Extractor(r); err == nil && id != "" {
						tenantID = id
						source = "extractor"
					}
				} else if headerValue := r.Header.Get(header); headerValue != "" {
					tenantID = headerValue
					source = "header"
				}
			}

			// 3. Handle missing tenant.
			if tenantID == "" {
				if !requireTenant {
					next.ServeHTTP(w, r)
					return
				}
				if options.OnMissing != nil {
					options.OnMissing(w, r)
					return
				}
				writeTenantError(w, r, http.StatusUnauthorized, "tenant_required", "tenant id is required", contract.CategoryAuthentication)
				return
			}

			// 4. Validate tenant ID format before storing in context.
			if err := tenant.ValidateTenantID(tenantID); err != nil {
				writeTenantError(w, r, http.StatusBadRequest, "invalid_tenant_id", "invalid tenant ID format", contract.CategoryAuthentication)
				return
			}

			// 5. Store in context and invoke hook.
			r = tenant.RequestWithTenantID(r, tenantID)
			options.Hooks.Resolve(r.Context(), tenant.ResolveInfo{
				TenantID: tenantID,
				Source:   source,
			})

			next.ServeHTTP(w, r)
		})
	}
}
