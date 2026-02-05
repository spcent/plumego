package tenant

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/tenant"
)

// TenantPolicyOptions configures tenant policy checks.
type TenantPolicyOptions struct {
	Evaluator   tenant.PolicyEvaluator
	ModelHeader string
	ToolHeader  string
	Hooks       tenant.Hooks
	OnDenied    func(http.ResponseWriter, *http.Request, tenant.PolicyResult)
}

// TenantPolicy enforces tenant policy checks.
func TenantPolicy(options TenantPolicyOptions) middleware.Middleware {
	modelHeader := headerOrDefault(options.ModelHeader, defaultModelHeader)
	toolHeader := headerOrDefault(options.ToolHeader, defaultToolHeader)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if options.Evaluator == nil {
				next.ServeHTTP(w, r)
				return
			}

			tenantID := tenant.TenantIDFromContext(r.Context())
			req := tenant.PolicyRequest{
				Model:  r.Header.Get(modelHeader),
				Tool:   r.Header.Get(toolHeader),
				Method: r.Method,
				Path:   r.URL.Path,
			}
			result, err := options.Evaluator.Evaluate(r.Context(), tenantID, req)
			allowed := err == nil && result.Allowed
			status := http.StatusForbidden

			options.Hooks.Policy(r.Context(), tenant.PolicyDecision{
				TenantID: tenantID,
				Allowed:  allowed,
				Reason:   result.Reason,
				Model:    req.Model,
				Tool:     req.Tool,
				Method:   req.Method,
				Path:     req.Path,
				Status:   status,
			})

			if allowed {
				next.ServeHTTP(w, r)
				return
			}

			if options.OnDenied != nil {
				options.OnDenied(w, r, result)
				return
			}

			writeTenantError(w, r, status, "policy_denied", "tenant policy denied request", contract.CategoryAuthentication)
		})
	}
}
