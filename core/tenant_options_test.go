package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/tenant"
)

func TestWithTenantConfigManager(t *testing.T) {
	manager := tenant.NewInMemoryConfigManager()
	manager.SetTenantConfig(tenant.Config{
		TenantID: "test-tenant",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 100,
		},
	})

	app := New(
		WithTenantConfigManager(manager),
	)

	// Verify component was added
	if len(app.components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(app.components))
	}

	// Verify it's a TenantConfigComponent
	tenantComp, ok := app.components[0].(*TenantConfigComponent)
	if !ok {
		t.Fatal("expected TenantConfigComponent")
	}

	if tenantComp.Manager != manager {
		t.Error("expected manager to be set")
	}
}

func TestWithTenantMiddleware_Full(t *testing.T) {
	manager := tenant.NewInMemoryConfigManager()
	manager.SetTenantConfig(tenant.Config{
		TenantID: "test-tenant",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 10,
		},
		Policy: tenant.PolicyConfig{
			AllowedModels: []string{"gpt-4"},
		},
	})

	quotaMgr := tenant.NewInMemoryQuotaManager(manager)
	policyEval := tenant.NewConfigPolicyEvaluator(manager)

	app := New(
		WithTenantMiddleware(TenantMiddlewareOptions{
			HeaderName:      "X-Custom-Tenant",
			QuotaManager:    quotaMgr,
			PolicyEvaluator: policyEval,
		}),
	)

	// Add a test route
	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		tenantID := tenant.TenantIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tenantID))
	})

	// Setup server to initialize handler
	if err := app.setupServer(); err != nil {
		t.Fatalf("setupServer failed: %v", err)
	}

	// Test with tenant header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom-Tenant", "test-tenant")
	rec := httptest.NewRecorder()

	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "test-tenant" {
		t.Errorf("expected 'test-tenant', got '%s'", rec.Body.String())
	}
}

func TestWithTenantMiddleware_QuotaOnly(t *testing.T) {
	manager := tenant.NewInMemoryConfigManager()
	manager.SetTenantConfig(tenant.Config{
		TenantID: "test-tenant",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 2,
		},
	})

	quotaMgr := tenant.NewInMemoryQuotaManager(manager)

	app := New(
		WithTenantMiddleware(TenantMiddlewareOptions{
			QuotaManager: quotaMgr,
		}),
	)

	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Setup server
	if err := app.setupServer(); err != nil {
		t.Fatalf("setupServer failed: %v", err)
	}

	// First request - should be allowed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Tenant-ID", "test-tenant")
	rec1 := httptest.NewRecorder()
	app.handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("first request should be allowed, got %d", rec1.Code)
	}

	// Second request - should be allowed
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Tenant-ID", "test-tenant")
	rec2 := httptest.NewRecorder()
	app.handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("second request should be allowed, got %d", rec2.Code)
	}

	// Third request - should be denied (quota exceeded)
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req3.Header.Set("X-Tenant-ID", "test-tenant")
	rec3 := httptest.NewRecorder()
	app.handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("third request should be denied, got %d", rec3.Code)
	}
}

func TestWithTenantMiddleware_PolicyOnly(t *testing.T) {
	manager := tenant.NewInMemoryConfigManager()
	manager.SetTenantConfig(tenant.Config{
		TenantID: "restricted-tenant",
		Policy: tenant.PolicyConfig{
			AllowedModels: []string{"gpt-3.5-turbo"},
		},
	})

	policyEval := tenant.NewConfigPolicyEvaluator(manager)

	app := New(
		WithTenantMiddleware(TenantMiddlewareOptions{
			PolicyEvaluator: policyEval,
		}),
	)

	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Setup server to initialize handler
	if err := app.setupServer(); err != nil {
		t.Fatalf("setupServer failed: %v", err)
	}

	// Allowed model
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Tenant-ID", "restricted-tenant")
	req1.Header.Set("X-Model", "gpt-3.5-turbo")
	rec1 := httptest.NewRecorder()
	app.handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("allowed model should pass, got %d", rec1.Code)
	}

	// Disallowed model
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Tenant-ID", "restricted-tenant")
	req2.Header.Set("X-Model", "gpt-4")
	rec2 := httptest.NewRecorder()
	app.handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("disallowed model should be forbidden, got %d", rec2.Code)
	}
}

func TestWithTenantMiddleware_AllowMissing(t *testing.T) {
	app := New(
		WithTenantMiddleware(TenantMiddlewareOptions{
			AllowMissing: true,
		}),
	)

	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		tenantID := tenant.TenantIDFromContext(r.Context())
		if tenantID == "" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("no-tenant"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(tenantID))
		}
	})

	// Setup server to initialize handler
	if err := app.setupServer(); err != nil {
		t.Fatalf("setupServer failed: %v", err)
	}

	// Request without tenant header - should be allowed
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "no-tenant" {
		t.Errorf("expected 'no-tenant', got '%s'", rec.Body.String())
	}
}

func TestWithTenantMiddleware_Integration(t *testing.T) {
	manager := tenant.NewInMemoryConfigManager()
	manager.SetTenantConfig(tenant.Config{
		TenantID: "full-tenant",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 5,
		},
		Policy: tenant.PolicyConfig{
			AllowedModels: []string{"gpt-4"},
		},
	})

	quotaMgr := tenant.NewInMemoryQuotaManager(manager)
	policyEval := tenant.NewConfigPolicyEvaluator(manager)

	resolveCount := 0
	quotaCount := 0
	policyCount := 0

	app := New(
		WithTenantConfigManager(manager),
		WithTenantMiddleware(TenantMiddlewareOptions{
			QuotaManager:    quotaMgr,
			PolicyEvaluator: policyEval,
			Hooks: tenant.Hooks{
				OnResolve: func(ctx context.Context, info tenant.ResolveInfo) {
					resolveCount++
				},
				OnQuota: func(ctx context.Context, info tenant.QuotaDecision) {
					quotaCount++
				},
				OnPolicy: func(ctx context.Context, info tenant.PolicyDecision) {
					policyCount++
				},
			},
		}),
	)

	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Setup server to initialize handler
	if err := app.setupServer(); err != nil {
		t.Fatalf("setupServer failed: %v", err)
	}

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Tenant-ID", "full-tenant")
	req.Header.Set("X-Model", "gpt-4")
	rec := httptest.NewRecorder()
	app.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Verify hooks were called
	if resolveCount != 1 {
		t.Errorf("expected 1 resolve hook call, got %d", resolveCount)
	}
	if quotaCount != 1 {
		t.Errorf("expected 1 quota hook call, got %d", quotaCount)
	}
	if policyCount != 1 {
		t.Errorf("expected 1 policy hook call, got %d", policyCount)
	}
}
