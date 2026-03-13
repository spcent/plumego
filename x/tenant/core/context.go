package tenant

import (
	"context"
	"net/http"
)

type tenantIDContextKey struct{}

var tenantIDContextKeyVar tenantIDContextKey

// ContextWithTenantID attaches a tenant id to context.
func ContextWithTenantID(ctx context.Context, tenantID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, tenantIDContextKeyVar, tenantID)
}

// TenantIDFromContext extracts tenant id from context.
func TenantIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(tenantIDContextKeyVar); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// RequestWithTenantID returns a shallow copy of r with tenant id attached.
func RequestWithTenantID(r *http.Request, tenantID string) *http.Request {
	if r == nil {
		return nil
	}
	return r.WithContext(ContextWithTenantID(r.Context(), tenantID))
}
