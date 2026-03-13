package tenant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContextWithTenantID(t *testing.T) {
	ctx := context.Background()
	tenantID := "test-tenant-123"

	// Add tenant ID to context
	newCtx := ContextWithTenantID(ctx, tenantID)

	// Verify it was added
	retrieved := TenantIDFromContext(newCtx)
	if retrieved != tenantID {
		t.Errorf("expected %s, got %s", tenantID, retrieved)
	}
}

func TestTenantIDFromContext_Present(t *testing.T) {
	ctx := context.Background()
	expected := "my-tenant"

	ctx = ContextWithTenantID(ctx, expected)
	actual := TenantIDFromContext(ctx)

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestTenantIDFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	// Context without tenant ID should return empty string
	tenantID := TenantIDFromContext(ctx)
	if tenantID != "" {
		t.Errorf("expected empty string, got %s", tenantID)
	}
}

func TestTenantIDFromContext_Nil(t *testing.T) {
	// Nil context should return empty string (not panic)
	tenantID := TenantIDFromContext(t.Context())
	if tenantID != "" {
		t.Errorf("expected empty string for nil context, got %s", tenantID)
	}
}

func TestContextWithTenantID_Nil(t *testing.T) {
	// Nil context should create background context
	tenantID := "test-tenant"
	ctx := ContextWithTenantID(t.Context(), tenantID)

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	retrieved := TenantIDFromContext(ctx)
	if retrieved != tenantID {
		t.Errorf("expected %s, got %s", tenantID, retrieved)
	}
}

func TestContextWithTenantID_Override(t *testing.T) {
	ctx := context.Background()

	// Set first tenant
	ctx = ContextWithTenantID(ctx, "tenant-1")

	// Override with second tenant
	ctx = ContextWithTenantID(ctx, "tenant-2")

	// Should have the second tenant
	retrieved := TenantIDFromContext(ctx)
	if retrieved != "tenant-2" {
		t.Errorf("expected tenant-2, got %s", retrieved)
	}
}

func TestContextWithTenantID_EmptyString(t *testing.T) {
	ctx := context.Background()

	// Set empty tenant ID
	ctx = ContextWithTenantID(ctx, "")

	// Should be able to retrieve empty string
	retrieved := TenantIDFromContext(ctx)
	if retrieved != "" {
		t.Errorf("expected empty string, got %s", retrieved)
	}
}

func TestRequestWithTenantID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	tenantID := "request-tenant"

	// Add tenant to request
	newReq := RequestWithTenantID(req, tenantID)

	// Should return new request
	if newReq == req {
		t.Error("expected new request instance")
	}

	// Verify tenant is in context
	retrieved := TenantIDFromContext(newReq.Context())
	if retrieved != tenantID {
		t.Errorf("expected %s, got %s", tenantID, retrieved)
	}

	// Original request should be unchanged
	original := TenantIDFromContext(req.Context())
	if original != "" {
		t.Errorf("original request should not have tenant, got %s", original)
	}
}

func TestRequestWithTenantID_Nil(t *testing.T) {
	// Nil request should return nil (not panic)
	newReq := RequestWithTenantID(nil, "tenant")
	if newReq != nil {
		t.Errorf("expected nil for nil request, got %v", newReq)
	}
}

func TestRequestWithTenantID_ChainedCalls(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Chain multiple calls
	req = RequestWithTenantID(req, "tenant-1")
	req = RequestWithTenantID(req, "tenant-2")
	req = RequestWithTenantID(req, "tenant-3")

	// Should have the last tenant
	retrieved := TenantIDFromContext(req.Context())
	if retrieved != "tenant-3" {
		t.Errorf("expected tenant-3, got %s", retrieved)
	}
}

func TestTenantIDFromContext_WrongType(t *testing.T) {
	ctx := context.Background()

	// Manually add wrong type to context (simulate corruption)
	ctx = context.WithValue(ctx, tenantIDContextKeyVar, 12345) // int instead of string

	// Should handle gracefully and return empty string
	tenantID := TenantIDFromContext(ctx)
	if tenantID != "" {
		t.Errorf("expected empty string for wrong type, got %s", tenantID)
	}
}

func TestContextPreservation(t *testing.T) {
	// Create context with existing values
	ctx := context.Background()
	type otherKey struct{}
	ctx = context.WithValue(ctx, otherKey{}, "other-value")

	// Add tenant ID
	ctx = ContextWithTenantID(ctx, "my-tenant")

	// Verify both values are present
	tenantID := TenantIDFromContext(ctx)
	if tenantID != "my-tenant" {
		t.Errorf("expected my-tenant, got %s", tenantID)
	}

	otherValue := ctx.Value(otherKey{})
	if otherValue != "other-value" {
		t.Errorf("expected other-value, got %v", otherValue)
	}
}

func TestRequestPreservation(t *testing.T) {
	// Create request with existing context values
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	type customKey struct{}
	ctx := context.WithValue(req.Context(), customKey{}, "custom")
	req = req.WithContext(ctx)

	// Add tenant ID
	req = RequestWithTenantID(req, "tenant-id")

	// Both values should be present
	tenantID := TenantIDFromContext(req.Context())
	if tenantID != "tenant-id" {
		t.Errorf("expected tenant-id, got %s", tenantID)
	}

	customValue := req.Context().Value(customKey{})
	if customValue != "custom" {
		t.Errorf("expected custom, got %v", customValue)
	}

	// Request properties should be preserved
	if req.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", req.Method)
	}
	if req.URL.Path != "/test" {
		t.Errorf("expected /test, got %s", req.URL.Path)
	}
}

func BenchmarkContextWithTenantID(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_ = ContextWithTenantID(ctx, "tenant-123")
	}
}

func BenchmarkTenantIDFromContext(b *testing.B) {
	ctx := ContextWithTenantID(context.Background(), "tenant-123")
	for i := 0; i < b.N; i++ {
		_ = TenantIDFromContext(ctx)
	}
}

func BenchmarkRequestWithTenantID(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for i := 0; i < b.N; i++ {
		_ = RequestWithTenantID(req, "tenant-123")
	}
}
