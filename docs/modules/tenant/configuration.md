# Tenant Configuration Management

> **Package**: `github.com/spcent/plumego/x/tenant/core` | **Feature**: Per-tenant configuration

## Overview

Tenant configuration management provides a way to store and retrieve per-tenant settings. Plumego supports both in-memory (testing/development) and database-backed (production) config managers with caching.

## ConfigManager Interface

```go
type ConfigManager interface {
    GetTenantConfig(ctx context.Context, tenantID string) (Config, error)
}
```

## In-Memory Config Manager

Best for testing and development:

```go
import "github.com/spcent/plumego/x/tenant/core"

// Create in-memory manager
configMgr := tenant.NewInMemoryConfigManager()

// Set tenant configurations
configMgr.SetTenantConfig("tenant-a", tenant.Config{
    TenantID: "tenant-a",
    Quota: tenant.QuotaConfig{
        RequestsPerMinute: 1000,
        MaxUsers:          100,
        StorageLimitGB:    10,
    },
    Policy: tenant.PolicyConfig{
        AllowedFeatures: []string{"api", "webhooks", "analytics"},
    },
    Metadata: map[string]string{
        "name":  "Acme Corp",
        "plan":  "pro",
        "email": "admin@acme.com",
    },
    UpdatedAt: time.Now(),
})

configMgr.SetTenantConfig("tenant-b", tenant.Config{
    TenantID: "tenant-b",
    Quota: tenant.QuotaConfig{
        RequestsPerMinute: 100,
        MaxUsers:          5,
        StorageLimitGB:    1,
    },
    Metadata: map[string]string{
        "name": "Startup Inc",
        "plan": "free",
    },
})

// Retrieve config
config, err := configMgr.GetTenantConfig(ctx, "tenant-a")
```

## Database Config Manager

For production use with automatic caching:

```go
import (
    "github.com/spcent/plumego/x/tenant/core"
    tenantconfig "github.com/spcent/plumego/x/tenant/config"
)

// Create with LRU cache (1000 entries, 5 min TTL)
configMgr := tenantconfig.NewDBTenantConfigManager(
    database,
    tenantconfig.WithTenantCache(1000, 5*time.Minute),
)
```

### Database Schema

```sql
CREATE TABLE tenant_configs (
    tenant_id       VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    plan            VARCHAR(50) NOT NULL DEFAULT 'free',
    quota_rpm       INT NOT NULL DEFAULT 60,
    quota_max_users INT NOT NULL DEFAULT 3,
    quota_storage   DECIMAL(10,2) NOT NULL DEFAULT 1.0,
    features        JSON,
    metadata        JSON,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### Refresh on Update

```go
// After updating tenant in database, refresh cache
func handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
    tenantID := plumego.Param(r, "id")

    // Update in database
    if err := db.UpdateTenant(tenantID, updates); err != nil {
        http.Error(w, "Update failed", http.StatusInternalServerError)
        return
    }

    // Invalidate cache so next request fetches fresh config
    configMgr.InvalidateCache(tenantID)

    json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
```

## Custom Config Manager

Implement the interface for any backend:

```go
type RedisConfigManager struct {
    client *redis.Client
    ttl    time.Duration
}

func (m *RedisConfigManager) GetTenantConfig(ctx context.Context, tenantID string) (tenant.Config, error) {
    key := fmt.Sprintf("tenant:config:%s", tenantID)

    data, err := m.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return tenant.Config{}, tenant.ErrTenantNotFound
    }
    if err != nil {
        return tenant.Config{}, err
    }

    var config tenant.Config
    if err := json.Unmarshal(data, &config); err != nil {
        return tenant.Config{}, err
    }

    return config, nil
}

func (m *RedisConfigManager) SetTenantConfig(ctx context.Context, tenantID string, config tenant.Config) error {
    key := fmt.Sprintf("tenant:config:%s", tenantID)
    data, err := json.Marshal(config)
    if err != nil {
        return err
    }
    return m.client.Set(ctx, key, data, m.ttl).Err()
}
```

## Config Structure

### Full Config Example

```go
config := tenant.Config{
    TenantID: "tenant-123",

    Quota: tenant.QuotaConfig{
        RequestsPerMinute: 1000,
        RequestsPerDay:    50000,
        MaxUsers:          100,
        StorageLimitGB:    50.0,
        MaxConnections:    10,
        MaxFileUploadMB:   50,
    },

    Policy: tenant.PolicyConfig{
        AllowedFeatures:  []string{"api", "webhooks", "analytics", "exports"},
        BlockedIPs:       []string{},
        AllowedCountries: []string{"US", "CA", "GB"},
        CustomRules: map[string]interface{}{
            "max_export_rows": 100000,
            "retention_days":  365,
            "sso_enabled":     true,
        },
    },

    Metadata: map[string]string{
        "name":          "Acme Corporation",
        "plan":          "enterprise",
        "billing_email": "billing@acme.com",
        "region":        "us-east-1",
        "created_by":    "admin",
    },

    UpdatedAt: time.Now(),
}
```

## Best Practices

### 1. Cache Config Aggressively

Config rarely changes — cache for minutes:

```go
// 5-minute cache prevents DB hammering
configMgr := tenantconfig.NewDBTenantConfigManager(db,
    tenantconfig.WithTenantCache(1000, 5*time.Minute),
)
```

### 2. Handle Missing Tenants Gracefully

```go
config, err := configMgr.GetTenantConfig(ctx, tenantID)
if err != nil {
    if errors.Is(err, tenant.ErrTenantNotFound) {
        http.Error(w, "Unknown tenant", http.StatusUnauthorized)
        return
    }
    http.Error(w, "Configuration error", http.StatusInternalServerError)
    return
}
```

### 3. Validate Config at Startup

```go
func validateTenantConfig(config tenant.Config) error {
    if config.TenantID == "" {
        return errors.New("tenant_id is required")
    }
    if config.Quota.RequestsPerMinute <= 0 {
        return errors.New("requests_per_minute must be positive")
    }
    return nil
}
```

## Related Documentation

- [Tenant Overview](README.md) — Multi-tenancy overview
- [Quota Management](quota.md) — Quota enforcement
- [Database Isolation](database-isolation.md) — Tenant-scoped queries
