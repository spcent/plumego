package tenant

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spcent/plumego/contract"
)

// ErrInvalidTenantID indicates a tenant ID failed validation.
var ErrInvalidTenantID = errors.New("invalid tenant ID")

// maxTenantIDLength is the maximum allowed length for a tenant ID.
const maxTenantIDLength = 255

// ValidateTenantID checks that a tenant ID is well-formed.
// Valid tenant IDs contain only alphanumeric characters, hyphens,
// underscores, and periods, with a maximum length of 255 bytes.
func ValidateTenantID(id string) error {
	if id == "" {
		return ErrTenantNotFound
	}
	if len(id) > maxTenantIDLength {
		return fmt.Errorf("%w: exceeds maximum length of %d", ErrInvalidTenantID, maxTenantIDLength)
	}
	if !utf8.ValidString(id) {
		return fmt.Errorf("%w: contains invalid UTF-8", ErrInvalidTenantID)
	}
	for _, c := range id {
		if !isValidTenantIDChar(c) {
			return fmt.Errorf("%w: contains invalid character %q", ErrInvalidTenantID, c)
		}
	}
	return nil
}

func isValidTenantIDChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '.'
}

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

			// Validate tenant ID format
			if err := ValidateTenantID(tenantID); err != nil {
				writeTenantError(w, http.StatusBadRequest, "Invalid tenant ID format")
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
					w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.RemainingRequests))
					if result.RetryAfter > 0 {
						w.Header().Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
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

// fromRequestValue creates a TenantExtractor from a function that reads
// a string value from the request. This eliminates duplication across
// header/query/cookie extractors.
func fromRequestValue(extract func(*http.Request) string) TenantExtractor {
	return func(r *http.Request) (string, error) {
		v := extract(r)
		if v == "" {
			return "", ErrTenantNotFound
		}
		return v, nil
	}
}

// FromHeader creates an extractor that reads from a header
func FromHeader(headerName string) TenantExtractor {
	return fromRequestValue(func(r *http.Request) string {
		return r.Header.Get(headerName)
	})
}

// FromQuery creates an extractor that reads from a query parameter
func FromQuery(paramName string) TenantExtractor {
	return fromRequestValue(func(r *http.Request) string {
		return r.URL.Query().Get(paramName)
	})
}

// FromCookie creates an extractor that reads from a cookie
func FromCookie(cookieName string) TenantExtractor {
	return fromRequestValue(func(r *http.Request) string {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return ""
		}
		return cookie.Value
	})
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

// FromContextValue creates an extractor that reads from a context value.
// This is useful for integrating with authentication middleware that stores
// parsed claims (e.g., JWT) in the request context.
//
// Example:
//
//	// With a JWT middleware that stores claims under a context key:
//	tenant.FromContextValue(jwtClaimsKey, func(v any) string {
//		claims, ok := v.(map[string]any)
//		if !ok { return "" }
//		id, _ := claims["tenant_id"].(string)
//		return id
//	})
func FromContextValue(key any, extract func(any) string) TenantExtractor {
	return func(r *http.Request) (string, error) {
		v := r.Context().Value(key)
		if v == nil {
			return "", ErrTenantNotFound
		}
		id := extract(v)
		if id == "" {
			return "", ErrTenantNotFound
		}
		return id, nil
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

// writeTenantError writes a tenant-related error response using the standard
// contract.WriteError format for consistency across the codebase.
func writeTenantError(w http.ResponseWriter, statusCode int, message string) {
	contract.WriteError(w, nil, contract.APIError{
		Status:   statusCode,
		Code:     http.StatusText(statusCode),
		Message:  message,
		Category: contract.CategoryForStatus(statusCode),
	})
}
