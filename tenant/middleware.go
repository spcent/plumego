package tenant

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Middleware creates an HTTP middleware for multi-tenancy
//
// Example:
//
//	app.Use(tenant.Middleware(tenant.MiddlewareConfig{
//		HeaderName:   "X-Tenant-ID",
//		QueryParam:   "tenant_id",
//		Extractor:    tenant.FromJWT("tenant_id"),
//		ConfigManager: configMgr,
//		QuotaManager:  quotaMgr,
//		PolicyEvaluator: policyEval,
//	}))
func Middleware(config MiddlewareConfig) func(http.Handler) http.Handler {
	// Apply defaults
	if config.HeaderName == "" {
		config.HeaderName = "X-Tenant-ID"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract tenant ID
			tenantID, err := config.extractTenantID(r)
			if err != nil {
				if !config.AllowMissing {
					writeTenantError(w, http.StatusBadRequest, "Missing or invalid tenant ID")
					return
				}
				// Continue without tenant context
				next.ServeHTTP(w, r)
				return
			}

			// Add tenant ID to context
			ctx := ContextWithTenantID(r.Context(), tenantID)

			// Check quota if configured
			if config.QuotaManager != nil {
				quotaReq := QuotaRequest{
					Requests: 1,
					Tokens:   0,
				}
				result, err := config.QuotaManager.Allow(ctx, tenantID, quotaReq)
				if err != nil || !result.Allowed {
					w.Header().Set("X-RateLimit-Remaining", string(rune(result.RemainingRequests)))
					if result.RetryAfter > 0 {
						w.Header().Set("Retry-After", string(rune(result.RetryAfter.Seconds())))
					}

					writeTenantError(w, http.StatusTooManyRequests, "Quota exceeded")
					return
				}
			}

			// Evaluate policy if configured
			if config.PolicyEvaluator != nil {
				policyReq := PolicyRequest{
					Method: r.Method,
					Path:   r.URL.Path,
				}

				result, err := config.PolicyEvaluator.Evaluate(ctx, tenantID, policyReq)
				if err != nil || !result.Allowed {
					writeTenantError(w, http.StatusForbidden, result.Reason)
					return
				}
			}

			// Continue with tenant context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MiddlewareConfig configures the tenant middleware
type MiddlewareConfig struct {
	// HeaderName is the HTTP header containing the tenant ID
	// Default: X-Tenant-ID
	HeaderName string

	// QueryParam is the query parameter containing the tenant ID
	// Optional, checked if header is not present
	QueryParam string

	// Extractor is a custom function to extract tenant ID
	// If set, takes precedence over HeaderName and QueryParam
	Extractor TenantExtractor

	// AllowMissing allows requests without a tenant ID
	// Default: false
	AllowMissing bool

	// ConfigManager provides tenant configuration
	// Optional
	ConfigManager ConfigManager

	// QuotaManager enforces tenant quotas
	// Optional
	QuotaManager QuotaManager

	// PolicyEvaluator enforces tenant policies
	// Optional
	PolicyEvaluator PolicyEvaluator

	// Hooks provides callbacks for tenant events
	// Optional
	Hooks *Hooks
}

// extractTenantID extracts the tenant ID from the request
func (c *MiddlewareConfig) extractTenantID(r *http.Request) (string, error) {
	// Use custom extractor if provided
	if c.Extractor != nil {
		return c.Extractor(r)
	}

	// Check header
	if c.HeaderName != "" {
		if tenantID := r.Header.Get(c.HeaderName); tenantID != "" {
			return tenantID, nil
		}
	}

	// Check query parameter
	if c.QueryParam != "" {
		if tenantID := r.URL.Query().Get(c.QueryParam); tenantID != "" {
			return tenantID, nil
		}
	}

	return "", ErrTenantNotFound
}

// TenantExtractor extracts tenant ID from a request
type TenantExtractor func(*http.Request) (string, error)

// FromHeader creates an extractor that reads from a header
func FromHeader(headerName string) TenantExtractor {
	return func(r *http.Request) (string, error) {
		tenantID := r.Header.Get(headerName)
		if tenantID == "" {
			return "", ErrTenantNotFound
		}
		return tenantID, nil
	}
}

// FromQuery creates an extractor that reads from a query parameter
func FromQuery(paramName string) TenantExtractor {
	return func(r *http.Request) (string, error) {
		tenantID := r.URL.Query().Get(paramName)
		if tenantID == "" {
			return "", ErrTenantNotFound
		}
		return tenantID, nil
	}
}

// FromSubdomain creates an extractor that parses from subdomain
// Example: tenant1.example.com -> tenant1
func FromSubdomain() TenantExtractor {
	return func(r *http.Request) (string, error) {
		host := r.Host
		// Remove port if present
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// Extract subdomain
		parts := strings.Split(host, ".")
		if len(parts) < 2 {
			return "", ErrTenantNotFound
		}

		return parts[0], nil
	}
}

// FromJWT creates an extractor that reads from JWT claims
// Note: This requires the request to have been authenticated first
// The JWT claims should be stored in the request context
func FromJWT(claimName string) TenantExtractor {
	return func(r *http.Request) (string, error) {
		// This is a placeholder - actual implementation depends on
		// how JWT claims are stored in the context
		// Users should implement this based on their auth middleware

		// Example if using a specific context key:
		// claims, ok := r.Context().Value("jwt_claims").(map[string]interface{})
		// if !ok {
		//     return "", ErrTenantNotFound
		// }
		// tenantID, ok := claims[claimName].(string)
		// if !ok {
		//     return "", ErrTenantNotFound
		// }
		// return tenantID, nil

		return "", ErrTenantNotFound
	}
}

// FromCookie creates an extractor that reads from a cookie
func FromCookie(cookieName string) TenantExtractor {
	return func(r *http.Request) (string, error) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return "", ErrTenantNotFound
		}
		if cookie.Value == "" {
			return "", ErrTenantNotFound
		}
		return cookie.Value, nil
	}
}

// Chain chains multiple extractors, trying each in order
func Chain(extractors ...TenantExtractor) TenantExtractor {
	return func(r *http.Request) (string, error) {
		for _, extractor := range extractors {
			tenantID, err := extractor(r)
			if err == nil && tenantID != "" {
				return tenantID, nil
			}
		}
		return "", ErrTenantNotFound
	}
}

// writeTenantError writes a tenant-related error response
func writeTenantError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": message,
	})
}
