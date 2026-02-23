# Multi-Tenant SaaS Examples

> **Package**: `github.com/spcent/plumego/tenant` | **Complete working examples**

## Overview

Complete examples demonstrating multi-tenant SaaS application patterns with Plumego.

## Example 1: Complete SaaS Application

Full-featured multi-tenant SaaS with admin API, tenant onboarding, and data isolation.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/tenant"
    "github.com/spcent/plumego/store/db"
    "github.com/google/uuid"
)

// Plan definitions
var plans = map[string]tenant.QuotaConfig{
    "free":       {RequestsPerMinute: 60, MaxUsers: 3, StorageLimitGB: 1},
    "starter":    {RequestsPerMinute: 300, MaxUsers: 10, StorageLimitGB: 5},
    "pro":        {RequestsPerMinute: 1000, MaxUsers: 100, StorageLimitGB: 50},
    "enterprise": {RequestsPerMinute: 10000, MaxUsers: 10000, StorageLimitGB: 1000},
}

var planFeatures = map[string][]string{
    "free":       {"api"},
    "starter":    {"api", "webhooks"},
    "pro":        {"api", "webhooks", "analytics", "exports"},
    "enterprise": {"api", "webhooks", "analytics", "exports", "sso", "audit_logs"},
}

func main() {
    // Setup
    database := connectDB()
    configMgr := db.NewDBTenantConfigManager(database,
        db.WithTenantCache(1000, 5*time.Minute),
    )
    quotaMgr := tenant.NewInMemoryQuotaManager(configMgr)
    policyEval := tenant.NewConfigPolicyEvaluator(configMgr)
    tenantDB := db.NewTenantDB(database)

    app := core.New(
        core.WithAddr(":8080"),
        core.WithTenantConfigManager(configMgr),
        core.WithTenantMiddleware(core.TenantMiddlewareOptions{
            HeaderName:      "X-Tenant-ID",
            AllowMissing:    false,
            QuotaManager:    quotaMgr,
            PolicyEvaluator: policyEval,
            Hooks: tenant.Hooks{
                OnQuota: func(ctx context.Context, d tenant.QuotaDecision) {
                    log.Printf("QUOTA: tenant=%s resource=%s used=%d limit=%d",
                        d.TenantID, d.Resource, d.Used, d.Limit)
                },
                OnPolicy: func(ctx context.Context, d tenant.PolicyDecision) {
                    log.Printf("POLICY DENIED: tenant=%s policy=%s reason=%s",
                        d.TenantID, d.PolicyName, d.Reason)
                },
            },
        }),
    )

    // Admin API (no tenant middleware)
    admin := app.Group("/admin", adminAuthMiddleware())
    admin.Post("/tenants", handleCreateTenant(configMgr))
    admin.Get("/tenants/:id", handleGetTenant(configMgr))
    admin.Put("/tenants/:id/plan", handleUpdatePlan(configMgr))
    admin.Delete("/tenants/:id", handleDeleteTenant(configMgr))

    // Tenant API (with tenant middleware)
    api := app.Group("/api")
    api.Get("/users", handleListUsers(tenantDB))
    api.Post("/users", handleCreateUser(tenantDB, quotaMgr))
    api.Get("/analytics", handleAnalytics(policyEval))
    api.Post("/exports", handleExport(policyEval))

    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
}

// Admin: Create tenant
func handleCreateTenant(configMgr tenant.ConfigManager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            Name  string `json:"name"`
            Email string `json:"email"`
            Plan  string `json:"plan"`
        }
        json.NewDecoder(r.Body).Decode(&req)

        // Validate plan
        quota, ok := plans[req.Plan]
        if !ok {
            http.Error(w, "Invalid plan", http.StatusBadRequest)
            return
        }

        tenantID := uuid.New().String()
        config := tenant.Config{
            TenantID: tenantID,
            Quota:    quota,
            Policy: tenant.PolicyConfig{
                AllowedFeatures: planFeatures[req.Plan],
            },
            Metadata: map[string]string{
                "name":  req.Name,
                "email": req.Email,
                "plan":  req.Plan,
            },
            UpdatedAt: time.Now(),
        }

        if err := configMgr.SetTenantConfig(tenantID, config); err != nil {
            http.Error(w, "Failed to create tenant", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{
            "tenant_id": tenantID,
            "api_key":   generateAPIKey(tenantID),
        })
    }
}

// Admin: Update plan
func handleUpdatePlan(configMgr tenant.ConfigManager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tenantID := r.PathValue("id")

        var req struct {
            Plan string `json:"plan"`
        }
        json.NewDecoder(r.Body).Decode(&req)

        quota, ok := plans[req.Plan]
        if !ok {
            http.Error(w, "Invalid plan", http.StatusBadRequest)
            return
        }

        config, err := configMgr.GetTenantConfig(r.Context(), tenantID)
        if err != nil {
            http.Error(w, "Tenant not found", http.StatusNotFound)
            return
        }

        config.Quota = quota
        config.Policy.AllowedFeatures = planFeatures[req.Plan]
        config.Metadata["plan"] = req.Plan
        config.UpdatedAt = time.Now()

        configMgr.SetTenantConfig(tenantID, config)

        json.NewEncoder(w).Encode(map[string]string{"status": "updated", "plan": req.Plan})
    }
}

// Tenant: List users (auto-scoped by tenant)
func handleListUsers(tenantDB *db.TenantDB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        rows, err := tenantDB.QueryFromContext(r.Context(),
            "SELECT id, email, name, created_at FROM users ORDER BY created_at DESC LIMIT 50",
        )
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var users []map[string]interface{}
        for rows.Next() {
            var id, email, name string
            var createdAt time.Time
            rows.Scan(&id, &email, &name, &createdAt)
            users = append(users, map[string]interface{}{
                "id": id, "email": email, "name": name,
                "created_at": createdAt,
            })
        }

        json.NewEncoder(w).Encode(map[string]interface{}{
            "users": users,
            "count": len(users),
        })
    }
}

// Tenant: Create user (with quota check)
func handleCreateUser(tenantDB *db.TenantDB, quotaMgr tenant.QuotaManager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tenantID := tenant.TenantIDFromContext(r.Context())

        var req struct {
            Email string `json:"email"`
            Name  string `json:"name"`
        }
        json.NewDecoder(r.Body).Decode(&req)

        // Check user quota
        decision, err := quotaMgr.CheckQuota(r.Context(), tenantID, "users", 1)
        if err != nil || !decision.Allowed {
            http.Error(w, fmt.Sprintf("User limit reached (%d/%d). Upgrade your plan.",
                decision.Used, decision.Limit), http.StatusForbidden)
            return
        }

        // Create user (auto-scoped)
        userID := uuid.New().String()
        _, err = tenantDB.ExecFromContext(r.Context(),
            "INSERT INTO users (id, email, name) VALUES (?, ?, ?)",
            userID, req.Email, req.Name,
        )
        if err != nil {
            http.Error(w, "Failed to create user", http.StatusInternalServerError)
            return
        }

        // Record usage
        quotaMgr.RecordUsage(r.Context(), tenantID, "users", 1)

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
    }
}

// Tenant: Analytics (feature-gated)
func handleAnalytics(policyEval tenant.PolicyEvaluator) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tenantID := tenant.TenantIDFromContext(r.Context())

        decision, _ := policyEval.EvaluatePolicy(r.Context(), tenantID, "feature.analytics", nil)
        if !decision.Allowed {
            http.Error(w, "Analytics not available on your plan. Upgrade to Pro.", http.StatusForbidden)
            return
        }

        // Return analytics data
        json.NewEncoder(w).Encode(map[string]interface{}{
            "page_views": 12450,
            "sessions":   3201,
            "bounce_rate": 0.42,
        })
    }
}

func adminAuthMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Header.Get("X-Admin-Key") != "secret-admin-key" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func generateAPIKey(tenantID string) string {
    return fmt.Sprintf("pk_%s_%s", tenantID[:8], uuid.New().String()[:8])
}
```

## Example 2: Tenant Onboarding Flow

```go
type OnboardingService struct {
    configMgr tenant.ConfigManager
    db        *db.TenantDB
    email     EmailService
}

func (s *OnboardingService) OnboardTenant(ctx context.Context, req OnboardRequest) (*Tenant, error) {
    tenantID := uuid.New().String()

    // 1. Create tenant config
    config := tenant.Config{
        TenantID: tenantID,
        Quota:    plans[req.Plan],
        Policy: tenant.PolicyConfig{
            AllowedFeatures: planFeatures[req.Plan],
        },
        Metadata: map[string]string{
            "name":  req.CompanyName,
            "email": req.AdminEmail,
            "plan":  req.Plan,
        },
    }

    if err := s.configMgr.SetTenantConfig(tenantID, config); err != nil {
        return nil, fmt.Errorf("create config: %w", err)
    }

    // 2. Create admin user with tenant context
    tenantCtx := tenant.ContextWithTenantID(ctx, tenantID)
    adminID := uuid.New().String()
    passwordHash, _ := password.Hash(req.AdminPassword)

    _, err := s.db.ExecFromContext(tenantCtx,
        "INSERT INTO users (id, email, name, password_hash, role) VALUES (?, ?, ?, ?, 'admin')",
        adminID, req.AdminEmail, req.AdminName, passwordHash,
    )
    if err != nil {
        return nil, fmt.Errorf("create admin: %w", err)
    }

    // 3. Send welcome email
    s.email.SendWelcome(req.AdminEmail, WelcomeData{
        CompanyName: req.CompanyName,
        TenantID:    tenantID,
        LoginURL:    fmt.Sprintf("https://app.example.com/login?tenant=%s", tenantID),
    })

    return &Tenant{
        ID:      tenantID,
        Name:    req.CompanyName,
        Plan:    req.Plan,
        AdminID: adminID,
    }, nil
}
```

## Example 3: Usage Analytics Dashboard

```go
func handleUsageDashboard(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())

    resources := []string{"requests_per_minute", "users", "storage_mb"}
    dashboard := map[string]interface{}{
        "tenant_id": tenantID,
        "timestamp": time.Now(),
        "usage":     map[string]interface{}{},
    }

    for _, resource := range resources {
        used, _ := quotaMgr.GetUsage(r.Context(), tenantID, resource)
        config, _ := configMgr.GetTenantConfig(r.Context(), tenantID)
        limit := getLimit(config.Quota, resource)

        dashboard["usage"].(map[string]interface{})[resource] = map[string]interface{}{
            "used":      used,
            "limit":     limit,
            "remaining": limit - used,
            "percent":   fmt.Sprintf("%.1f%%", float64(used)/float64(limit)*100),
        }
    }

    config, _ := configMgr.GetTenantConfig(r.Context(), tenantID)
    dashboard["plan"] = config.Metadata["plan"]
    dashboard["features"] = config.Policy.AllowedFeatures

    json.NewEncoder(w).Encode(dashboard)
}
```

**Response**:
```json
{
  "tenant_id": "tenant-123",
  "plan": "pro",
  "features": ["api", "webhooks", "analytics", "exports"],
  "usage": {
    "requests_per_minute": {"used": 450, "limit": 1000, "remaining": 550, "percent": "45.0%"},
    "users": {"used": 23, "limit": 100, "remaining": 77, "percent": "23.0%"},
    "storage_mb": {"used": 2048, "limit": 51200, "remaining": 49152, "percent": "4.0%"}
  }
}
```

## Testing Multi-Tenant Apps

```go
func TestTenantIsolation(t *testing.T) {
    configMgr := tenant.NewInMemoryConfigManager()
    tenantDB := db.NewTenantDB(testDB)

    // Create two tenants
    configMgr.SetTenantConfig("tenant-a", tenant.Config{TenantID: "tenant-a"})
    configMgr.SetTenantConfig("tenant-b", tenant.Config{TenantID: "tenant-b"})

    // Create data for tenant-a
    ctxA := tenant.ContextWithTenantID(context.Background(), "tenant-a")
    tenantDB.ExecFromContext(ctxA, "INSERT INTO users (id, email) VALUES ('user-1', 'a@test.com')")

    // Tenant-a can see their own data
    var count int
    tenantDB.QueryRowFromContext(ctxA, "SELECT COUNT(*) FROM users").Scan(&count)
    if count != 1 {
        t.Errorf("Expected 1 user for tenant-a, got %d", count)
    }

    // Tenant-b cannot see tenant-a's data
    ctxB := tenant.ContextWithTenantID(context.Background(), "tenant-b")
    tenantDB.QueryRowFromContext(ctxB, "SELECT COUNT(*) FROM users").Scan(&count)
    if count != 0 {
        t.Errorf("Expected 0 users for tenant-b, got %d", count)
    }
}
```

## Related Documentation

- [Tenant Overview](README.md) — Multi-tenancy overview
- [Configuration](configuration.md) — Tenant config management
- [Quota Management](quota.md) — Quota enforcement
- [Database Isolation](database-isolation.md) — Data isolation

## Reference Implementation

See `examples/multi-tenant-saas/` for complete working application.
