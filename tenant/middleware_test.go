package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateTenantID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{"valid alphanumeric", "tenant-123", nil},
		{"valid with dots", "org.team.svc", nil},
		{"valid with underscores", "my_tenant_1", nil},
		{"valid uuid-like", "550e8400-e29b-41d4-a716-446655440000", nil},
		{"empty", "", ErrTenantNotFound},
		{"too long", strings.Repeat("a", 256), ErrInvalidTenantID},
		{"max length", strings.Repeat("a", 255), nil},
		{"contains space", "tenant 123", ErrInvalidTenantID},
		{"contains slash", "tenant/123", ErrInvalidTenantID},
		{"contains newline", "tenant\n123", ErrInvalidTenantID},
		{"contains null byte", "tenant\x00123", ErrInvalidTenantID},
		{"contains angle bracket", "tenant<script>", ErrInvalidTenantID},
		{"contains colon", "tenant:123", ErrInvalidTenantID},
		{"unicode characters", "tenant-日本語", ErrInvalidTenantID},
		{"emoji", "tenant-😀", ErrInvalidTenantID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTenantID(tt.id)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateTenantID(%q) = %v, want nil", tt.id, err)
				}
				return
			}
			if err == nil {
				t.Errorf("ValidateTenantID(%q) = nil, want %v", tt.id, tt.wantErr)
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateTenantID(%q) = %v, want %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestMiddleware_ExtractFromHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := TenantIDFromContext(r.Context())
		w.Write([]byte(tenantID))
	})

	mw := Middleware(MiddlewareConfig{})
	wrapped := mw(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "test-tenant")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "test-tenant" {
		t.Errorf("expected body 'test-tenant', got %q", rr.Body.String())
	}
}

func TestMiddleware_MissingTenantID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	mw := Middleware(MiddlewareConfig{})
	wrapped := mw(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestMiddleware_AllowMissing(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		tenantID := TenantIDFromContext(r.Context())
		if tenantID != "" {
			t.Errorf("expected empty tenant ID, got %q", tenantID)
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(MiddlewareConfig{AllowMissing: true})
	wrapped := mw(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if !called {
		t.Error("handler should be called when AllowMissing is true")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestMiddleware_InvalidTenantID_Rejected(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for invalid tenant ID")
	})

	mw := Middleware(MiddlewareConfig{})
	wrapped := mw(handler)

	tests := []struct {
		name     string
		tenantID string
	}{
		{"contains space", "tenant 123"},
		{"contains slash", "tenant/123"},
		{"contains newline", "tenant\n123"},
		{"contains angle bracket", "<script>alert(1)</script>"},
		{"very long", strings.Repeat("x", 300)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Tenant-ID", tt.tenantID)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected status 400 for tenant ID %q, got %d", tt.tenantID, rr.Code)
			}
		})
	}
}

func TestMiddleware_QuotaHeaders(t *testing.T) {
	cfgMgr := NewInMemoryConfigManager()
	cfgMgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Quota: QuotaConfig{
			RequestsPerMinute: 1,
		},
	})
	quotaMgr := NewInMemoryQuotaManager(cfgMgr)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(MiddlewareConfig{
		QuotaManager: quotaMgr,
	})
	wrapped := mw(handler)

	// First request - allowed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Tenant-ID", "test-tenant")
	rr1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", rr1.Code)
	}

	// Second request - denied with proper headers
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Tenant-ID", "test-tenant")
	rr2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rr2.Code)
	}

	// Verify X-RateLimit-Remaining is a valid number (not garbled rune conversion)
	remaining := rr2.Header().Get("X-RateLimit-Remaining")
	if remaining != "0" {
		t.Errorf("expected X-RateLimit-Remaining '0', got %q", remaining)
	}

	retryAfter := rr2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header to be set")
	}
	// Verify it's a valid numeric string, not a garbled rune
	for _, c := range retryAfter {
		if c < '0' || c > '9' {
			t.Errorf("Retry-After should be numeric, got %q", retryAfter)
			break
		}
	}
}

func TestMiddleware_PolicyDenied(t *testing.T) {
	cfgMgr := NewInMemoryConfigManager()
	cfgMgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Policy: PolicyConfig{
			AllowedModels: []string{"gpt-4"},
		},
	})
	policyEval := NewConfigPolicyEvaluator(cfgMgr)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(MiddlewareConfig{
		PolicyEvaluator: policyEval,
	})
	wrapped := mw(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "test-tenant")
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	// The middleware uses empty model/tool in PolicyRequest, which should be allowed
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 (empty model/tool allowed), got %d", rr.Code)
	}
}

func TestMiddleware_QueryParam(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := TenantIDFromContext(r.Context())
		w.Write([]byte(tenantID))
	})

	mw := Middleware(MiddlewareConfig{
		QueryParam: "tenant_id",
	})
	wrapped := mw(handler)

	req := httptest.NewRequest("GET", "/test?tenant_id=query-tenant", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "query-tenant" {
		t.Errorf("expected 'query-tenant', got %q", rr.Body.String())
	}
}

func TestMiddleware_ErrorResponseFormat(t *testing.T) {
	mw := Middleware(MiddlewareConfig{})
	wrapped := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	// Response uses the standard contract.WriteError format:
	// {"error": {"status": 400, "code": "...", "message": "..."}}
	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	errObj, ok := body["error"]
	if !ok {
		t.Fatal("expected 'error' field in response")
	}
	errMap, ok := errObj.(map[string]interface{})
	if !ok {
		t.Fatal("expected 'error' field to be an object")
	}
	if _, ok := errMap["message"]; !ok {
		t.Error("expected 'message' field in error object")
	}
	if _, ok := errMap["code"]; !ok {
		t.Error("expected 'code' field in error object")
	}
}

// Extractor tests

func TestFromHeader(t *testing.T) {
	extractor := FromHeader("X-Custom-Tenant")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Custom-Tenant", "my-tenant")

	id, err := extractor(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "my-tenant" {
		t.Errorf("expected 'my-tenant', got %q", id)
	}
}

func TestFromHeader_Missing(t *testing.T) {
	extractor := FromHeader("X-Custom-Tenant")
	req := httptest.NewRequest("GET", "/", nil)

	_, err := extractor(req)
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestFromQuery(t *testing.T) {
	extractor := FromQuery("tid")
	req := httptest.NewRequest("GET", "/?tid=q-tenant", nil)

	id, err := extractor(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "q-tenant" {
		t.Errorf("expected 'q-tenant', got %q", id)
	}
}

func TestFromCookie(t *testing.T) {
	extractor := FromCookie("tenant_cookie")
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "tenant_cookie", Value: "cookie-tenant"})

	id, err := extractor(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "cookie-tenant" {
		t.Errorf("expected 'cookie-tenant', got %q", id)
	}
}

func TestFromCookie_Missing(t *testing.T) {
	extractor := FromCookie("tenant_cookie")
	req := httptest.NewRequest("GET", "/", nil)

	_, err := extractor(req)
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestFromSubdomain(t *testing.T) {
	extractor := FromSubdomain()

	tests := []struct {
		name     string
		host     string
		expected string
		wantErr  bool
	}{
		{"standard", "acme.example.com", "acme", false},
		{"with port", "acme.example.com:8080", "acme", false},
		{"no subdomain", "localhost", "", true},
		{"bare domain no dot", "example", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Host = tt.host

			id, err := extractor(req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

type testContextKey struct{}

func TestFromContextValue(t *testing.T) {
	extractor := FromContextValue(testContextKey{}, func(v any) string {
		claims, ok := v.(map[string]string)
		if !ok {
			return ""
		}
		return claims["tenant_id"]
	})

	// With valid context value
	ctx := context.WithValue(context.Background(), testContextKey{}, map[string]string{
		"tenant_id": "jwt-tenant",
	})
	req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)

	id, err := extractor(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "jwt-tenant" {
		t.Errorf("expected 'jwt-tenant', got %q", id)
	}

	// Without context value
	req2 := httptest.NewRequest("GET", "/", nil)
	_, err = extractor(req2)
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestChain(t *testing.T) {
	extractor := Chain(
		FromHeader("X-Tenant-ID"),
		FromQuery("tenant_id"),
	)

	// Should find from header
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Header.Set("X-Tenant-ID", "header-tenant")
	id, err := extractor(req1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "header-tenant" {
		t.Errorf("expected 'header-tenant', got %q", id)
	}

	// Should fallback to query
	req2 := httptest.NewRequest("GET", "/?tenant_id=query-tenant", nil)
	id, err = extractor(req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "query-tenant" {
		t.Errorf("expected 'query-tenant', got %q", id)
	}

	// Should fail if neither present
	req3 := httptest.NewRequest("GET", "/", nil)
	_, err = extractor(req3)
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestMiddleware_QuotaRemainingHeaderValues(t *testing.T) {
	// This test specifically verifies the fix for the string(rune(int)) bug.
	// Before the fix, remaining=10 would produce header value "\n" (newline).
	cfgMgr := NewInMemoryConfigManager()
	cfgMgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Quota: QuotaConfig{
			RequestsPerMinute: 2,
		},
	})
	quotaMgr := NewInMemoryQuotaManager(cfgMgr)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(MiddlewareConfig{QuotaManager: quotaMgr})
	wrapped := mw(handler)

	now := time.Now().UTC()

	// Use up the quota
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-ID", "test-tenant")
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
	}

	// This request should be denied
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "test-tenant")
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}

	// The critical assertion: header value must be a decimal number, not a Unicode character
	remaining := rr.Header().Get("X-RateLimit-Remaining")
	if remaining != "0" {
		t.Errorf("X-RateLimit-Remaining: expected '0', got %q (len=%d)", remaining, len(remaining))
	}

	_ = now // used to keep imports clean
}
