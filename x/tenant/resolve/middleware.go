package resolve

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	tenanttransport "github.com/spcent/plumego/x/tenant/transport"
)

// Options configures tenant resolution.
type Options struct {
	// HeaderName is the HTTP header to extract tenant ID from (default: "X-Tenant-ID").
	HeaderName string
	// Extractor is a custom function to extract tenant ID from the request.
	// When set, takes precedence over HeaderName.
	Extractor tenantcore.TenantExtractor
	// AllowMissing allows requests to proceed when no tenant ID is found.
	AllowMissing bool
	// Hooks provides callbacks for tenant resolution events.
	Hooks tenantcore.Hooks
	// OnMissing is called when tenant ID is missing and AllowMissing is false.
	// If nil, a standard 401 JSON error is returned.
	OnMissing func(http.ResponseWriter, *http.Request)
}

// Middleware resolves tenant id from request and stores it in context.
// Resolution order: custom Extractor, then HeaderName header.
func Middleware(options Options) middleware.Middleware {
	header := tenanttransport.HeaderOrDefault(options.HeaderName, tenanttransport.DefaultTenantHeader)
	requireTenant := !options.AllowMissing

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tenantID string
			source := ""

			if options.Extractor != nil {
				if id, err := options.Extractor(r); err == nil && id != "" {
					tenantID = id
					source = "extractor"
				}
			} else if headerValue := r.Header.Get(header); headerValue != "" {
				tenantID = headerValue
				source = "header"
			}

			if tenantID == "" {
				if !requireTenant {
					next.ServeHTTP(w, r)
					return
				}
				if options.OnMissing != nil {
					options.OnMissing(w, r)
					return
				}
				tenanttransport.WriteError(w, r, http.StatusUnauthorized, tenanttransport.CodeRequired, "tenant id is required", contract.CategoryAuth)
				return
			}

			if err := tenantcore.ValidateTenantID(tenantID); err != nil {
				tenanttransport.WriteError(w, r, http.StatusBadRequest, tenanttransport.CodeInvalidID, "invalid tenant ID format", contract.CategoryAuth)
				return
			}

			r = tenantcore.RequestWithTenantID(r, tenantID)
			options.Hooks.Resolve(r.Context(), tenantcore.ResolveInfo{
				TenantID: tenantID,
				Source:   source,
			})

			next.ServeHTTP(w, r)
		})
	}
}
