# Database Tenant Isolation

> **Package**: `github.com/spcent/plumego/store/db` | **Feature**: Automatic tenant scoping

## Overview

`TenantDB` wraps a standard `*sql.DB` to automatically inject `tenant_id` filters into all queries. This prevents cross-tenant data access without requiring every query to manually include tenant filtering.

## Quick Start

```go
import "github.com/spcent/plumego/store/db"

// Create tenant-aware database wrapper
tenantDB := db.NewTenantDB(database)

// Query with context that has tenant ID
ctx := tenant.ContextWithTenantID(r.Context(), "tenant-123")

// tenant_id = 'tenant-123' is automatically added
rows, err := tenantDB.QueryFromContext(ctx,
    "SELECT id, email, name FROM users WHERE active = ?", true,
)
// Executes: SELECT id, email, name FROM users WHERE active = ? AND tenant_id = 'tenant-123'
```

## How It Works

`TenantDB` uses query rewriting to automatically inject `tenant_id` conditions:

```
Original:  SELECT * FROM users WHERE active = true
Rewritten: SELECT * FROM users WHERE active = true AND tenant_id = 'tenant-123'

Original:  INSERT INTO orders (id, user_id) VALUES (?, ?)
Rewritten: INSERT INTO orders (id, user_id, tenant_id) VALUES (?, ?, ?)
```

The tenant ID is extracted from the context using `tenant.TenantIDFromContext(ctx)`.

## API

### Query (Multiple Rows)

```go
rows, err := tenantDB.QueryFromContext(ctx,
    "SELECT id, email, name, created_at FROM users ORDER BY created_at DESC LIMIT ?", 50,
)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var u User
    rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
    users = append(users, u)
}
```

### QueryRow (Single Row)

```go
var user User
err := tenantDB.QueryRowFromContext(ctx,
    "SELECT id, email, name FROM users WHERE id = ?", userID,
).Scan(&user.ID, &user.Email, &user.Name)

if err == sql.ErrNoRows {
    return nil, ErrUserNotFound
}
```

### Exec (Insert / Update / Delete)

```go
// Insert: tenant_id is automatically added
_, err := tenantDB.ExecFromContext(ctx,
    "INSERT INTO documents (id, title, content) VALUES (?, ?, ?)",
    docID, title, content,
)

// Update: WHERE tenant_id = ? is automatically added
_, err = tenantDB.ExecFromContext(ctx,
    "UPDATE documents SET title = ? WHERE id = ?",
    newTitle, docID,
)

// Delete: WHERE tenant_id = ? is automatically added
_, err = tenantDB.ExecFromContext(ctx,
    "DELETE FROM documents WHERE id = ?", docID,
)
```

### Transaction

```go
tx, err := tenantDB.BeginFromContext(ctx)
if err != nil {
    return err
}
defer tx.Rollback()

_, err = tx.ExecFromContext(ctx, "INSERT INTO orders ...")
if err != nil {
    return err
}

_, err = tx.ExecFromContext(ctx, "UPDATE inventory ...")
if err != nil {
    return err
}

return tx.Commit()
```

## Schema Requirements

Tables accessed through `TenantDB` must have a `tenant_id` column:

```sql
-- Required column on all tenant-scoped tables
CREATE TABLE users (
    id          VARCHAR(36)  PRIMARY KEY,
    tenant_id   VARCHAR(36)  NOT NULL,  -- Required for isolation
    email       VARCHAR(255) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP    DEFAULT NOW(),

    INDEX idx_users_tenant (tenant_id),
    UNIQUE KEY uk_users_email_tenant (tenant_id, email)
);

CREATE TABLE orders (
    id          VARCHAR(36) PRIMARY KEY,
    tenant_id   VARCHAR(36) NOT NULL,
    user_id     VARCHAR(36) NOT NULL,
    total       DECIMAL(10,2),
    status      VARCHAR(50),
    created_at  TIMESTAMP DEFAULT NOW(),

    INDEX idx_orders_tenant (tenant_id),
    INDEX idx_orders_user (tenant_id, user_id)
);
```

## Cross-Tenant Operations (Admin)

For admin operations that span all tenants, use the raw database handle:

```go
// Raw DB bypasses tenant isolation
rawDB := tenantDB.RawDB()

// Query all tenants (admin only)
rows, err := rawDB.QueryContext(ctx,
    "SELECT tenant_id, COUNT(*) FROM users GROUP BY tenant_id",
)

// Global aggregations
var totalUsers int
rawDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
```

**Security**: Only use `RawDB()` in admin routes protected by admin authentication. Never expose it to tenant API routes.

## Context Helper Functions

```go
import "github.com/spcent/plumego/tenant"

// Add tenant ID to context
ctx := tenant.ContextWithTenantID(context.Background(), "tenant-123")

// Extract tenant ID from context
tenantID := tenant.TenantIDFromContext(ctx)
if tenantID == "" {
    return errors.New("no tenant ID in context")
}
```

## Testing Tenant Isolation

```go
func TestTenantIsolation(t *testing.T) {
    database := openTestDB(t)
    tenantDB := db.NewTenantDB(database)

    // Setup two tenants
    ctxA := tenant.ContextWithTenantID(context.Background(), "tenant-a")
    ctxB := tenant.ContextWithTenantID(context.Background(), "tenant-b")

    // Create data for tenant-a
    tenantDB.ExecFromContext(ctxA,
        "INSERT INTO users (id, email) VALUES (?, ?)",
        "user-1", "alice@a.com",
    )

    // Tenant-a sees their data
    var countA int
    tenantDB.QueryRowFromContext(ctxA, "SELECT COUNT(*) FROM users").Scan(&countA)
    assert.Equal(t, 1, countA)

    // Tenant-b sees NO data (isolation working)
    var countB int
    tenantDB.QueryRowFromContext(ctxB, "SELECT COUNT(*) FROM users").Scan(&countB)
    assert.Equal(t, 0, countB)
}
```

## Tenant Configuration in DB

The `db` package also includes `DBTenantConfigManager` for storing tenant configurations in the database:

```go
configMgr := db.NewDBTenantConfigManager(database,
    db.WithTenantCache(1000, 5*time.Minute), // LRU cache with 5min TTL
)

// Get config (cached)
config, err := configMgr.GetTenantConfig(ctx, "tenant-123")

// Update config (invalidates cache)
err = configMgr.SetTenantConfig("tenant-123", tenant.Config{
    TenantID: "tenant-123",
    Quota:    tenant.QuotaConfig{RequestsPerMinute: 1000},
})
```

**Schema** for tenant config table:
```sql
CREATE TABLE tenant_configs (
    tenant_id   VARCHAR(36)  PRIMARY KEY,
    config_json TEXT         NOT NULL,
    updated_at  TIMESTAMP    DEFAULT NOW()
);
```

## Related Documentation

- [Tenant Overview](../../tenant/README.md) — Multi-tenancy architecture
- [Tenant Database Isolation](../../tenant/database-isolation.md) — Detailed patterns
- [Database Module](README.md) — DB wrapper configuration
