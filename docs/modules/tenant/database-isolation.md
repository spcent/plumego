# Database Isolation

> **Package**: `github.com/spcent/plumego/x/tenant/store/db` | **Feature**: Automatic tenant query scoping

## Overview

Database isolation ensures tenant data is kept separate through automatic query scoping. Plumego provides an `x/tenant/store/db` `TenantDB` wrapper that automatically injects `tenant_id` conditions into SQL queries, preventing data leakage between tenants.

## Setup

```go
import tenantdb "github.com/spcent/plumego/x/tenant/store/db"

// Wrap standard database/sql connection
tenantDB := tenantdb.NewTenantDB(database)
```

## Automatic Query Scoping

### SELECT Queries

```go
// Handler: tenant-a makes request
tenantID := tenant.TenantIDFromContext(ctx) // "tenant-a"

// Automatic tenant_id injection
rows, err := tenantDB.QueryFromContext(ctx,
    "SELECT id, name, email FROM users WHERE active = ?", true)

// Executed as:
// SELECT id, name, email FROM users WHERE tenant_id = 'tenant-a' AND active = true
```

### INSERT Queries

```go
_, err := tenantDB.ExecFromContext(ctx,
    "INSERT INTO users (name, email) VALUES (?, ?)",
    "John Doe", "john@example.com",
)

// Executed as:
// INSERT INTO users (tenant_id, name, email) VALUES ('tenant-a', 'John Doe', 'john@example.com')
```

### UPDATE Queries

```go
_, err := tenantDB.ExecFromContext(ctx,
    "UPDATE users SET name = ? WHERE id = ?",
    "Jane Doe", userID,
)

// Executed as:
// UPDATE users SET name = 'Jane Doe' WHERE tenant_id = 'tenant-a' AND id = ?
```

### DELETE Queries

```go
_, err := tenantDB.ExecFromContext(ctx,
    "DELETE FROM sessions WHERE expired = ?", true)

// Executed as:
// DELETE FROM sessions WHERE tenant_id = 'tenant-a' AND expired = true
```

## Database Schema

### Adding tenant_id to Tables

```sql
-- Add tenant_id to all multi-tenant tables
CREATE TABLE users (
    id          VARCHAR(36) PRIMARY KEY,
    tenant_id   VARCHAR(64) NOT NULL,   -- Required for isolation
    email       VARCHAR(255) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Ensure tenant_id in all indexes
    UNIQUE KEY uk_tenant_email (tenant_id, email),
    INDEX idx_tenant (tenant_id)
);

CREATE TABLE orders (
    id          VARCHAR(36) PRIMARY KEY,
    tenant_id   VARCHAR(64) NOT NULL,
    user_id     VARCHAR(36) NOT NULL,
    total       DECIMAL(10,2) NOT NULL,
    status      VARCHAR(50) NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_tenant (tenant_id),
    INDEX idx_tenant_user (tenant_id, user_id)
);
```

## Using TenantDB

### In Repository Pattern

```go
type UserRepository struct {
    db *tenantdb.TenantDB
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    var user User

    err := r.db.QueryRowFromContext(ctx,
        "SELECT id, email, name FROM users WHERE id = ?", id,
    ).Scan(&user.ID, &user.Email, &user.Name)

    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }

    return &user, err
    // Automatic: tenant_id scoped to current tenant from context
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    _, err := r.db.ExecFromContext(ctx,
        "INSERT INTO users (id, email, name) VALUES (?, ?, ?)",
        user.ID, user.Email, user.Name,
    )
    // Automatic: tenant_id injected from context
    return err
}

func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]*User, error) {
    rows, err := r.db.QueryFromContext(ctx,
        "SELECT id, email, name FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?",
        pageSize, (page-1)*pageSize,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        var u User
        rows.Scan(&u.ID, &u.Email, &u.Name)
        users = append(users, &u)
    }

    return users, nil
}
```

### Accessing Raw DB

For cross-tenant operations (admin, analytics):

```go
// Access underlying db/sql connection (no tenant scoping)
rawDB := tenantDB.RawDB()

// Count all tenants (admin operation)
var total int
rawDB.QueryRow("SELECT COUNT(DISTINCT tenant_id) FROM users").Scan(&total)

// Cross-tenant analytics (careful with data access!)
rows, _ := rawDB.Query("SELECT tenant_id, COUNT(*) FROM users GROUP BY tenant_id")
```

## Tenant Isolation Strategies

### Strategy 1: Shared Schema (Recommended)

All tenants share same tables with `tenant_id` column. Simplest to manage.

```sql
-- Single table for all tenants
CREATE TABLE users (
    id        VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    ...
    INDEX idx_tenant (tenant_id)
);
```

**Pros**: Simple, low overhead, easy to manage
**Cons**: Less isolation, requires careful query scoping

### Strategy 2: Schema-per-Tenant

Each tenant gets their own database schema.

```sql
-- Create schema per tenant
CREATE SCHEMA tenant_abc;
CREATE TABLE tenant_abc.users (...);
CREATE TABLE tenant_abc.orders (...);
```

```go
// Set search path per request
func getDB(ctx context.Context, tenantID string) *sql.DB {
    db.Exec(fmt.Sprintf("SET search_path TO tenant_%s", tenantID))
    return db
}
```

**Pros**: Strong isolation, easy backup/restore per tenant
**Cons**: More complex, schema migrations harder

### Strategy 3: Database-per-Tenant

Each tenant gets their own database.

```go
type MultiDBManager struct {
    dbs map[string]*sql.DB
    mu  sync.RWMutex
}

func (m *MultiDBManager) GetDB(tenantID string) *sql.DB {
    m.mu.RLock()
    db, ok := m.dbs[tenantID]
    m.mu.RUnlock()

    if !ok {
        db = connectTenantDB(tenantID)
        m.mu.Lock()
        m.dbs[tenantID] = db
        m.mu.Unlock()
    }

    return db
}
```

**Pros**: Maximum isolation, per-tenant scaling
**Cons**: High overhead, complex management

## Migrations

### Running Migrations per Tenant

```go
func runMigrations(tenantID string) error {
    // Get tenant's DB context
    ctx := tenant.ContextWithTenantID(context.Background(), tenantID)

    // Run migrations
    return migrator.Up(ctx, tenantDB)
}

// On new tenant creation
func createTenant(tenantID string) error {
    // Create tenant config
    configMgr.SetTenantConfig(tenantID, defaultConfig)

    // Run schema migrations for new tenant
    return runMigrations(tenantID)
}
```

### Global Migrations

```go
// Add tenant_id column to existing table
rawDB.Exec(`
    ALTER TABLE users
    ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64) NOT NULL DEFAULT '',
    ADD INDEX idx_tenant (tenant_id)
`)
```

## Testing

### Unit Test with Tenant Context

```go
func TestUserRepository(t *testing.T) {
    db := setupTestDB(t)
    tenantDB := tenantdb.NewTenantDB(db)
    repo := &UserRepository{db: tenantDB}

    // Create context with tenant ID
    ctx := tenant.ContextWithTenantID(context.Background(), "test-tenant")

    // Create user
    user := &User{ID: "user-1", Email: "test@example.com", Name: "Test User"}
    err := repo.Create(ctx, user)
    if err != nil {
        t.Fatal(err)
    }

    // Verify isolation: other tenant cannot see this user
    otherCtx := tenant.ContextWithTenantID(context.Background(), "other-tenant")
    found, err := repo.FindByID(otherCtx, "user-1")
    if err != ErrNotFound {
        t.Error("Expected not found for other tenant")
    }
}
```

## Best Practices

### 1. Always Use TenantDB for Tenant Data

```go
// ✅ Always scope tenant data
tenantDB.QueryFromContext(ctx, "SELECT ...")

// ❌ Never bypass scoping for tenant data
rawDB.Query("SELECT ...")
```

### 2. Index tenant_id

```sql
-- Always index tenant_id
CREATE INDEX idx_tenant_id ON users (tenant_id);

-- Composite index for common queries
CREATE INDEX idx_tenant_user ON orders (tenant_id, user_id);
```

### 3. Validate Tenant Ownership

```go
// Even with TenantDB, verify resource belongs to tenant
func getOrder(ctx context.Context, orderID string) (*Order, error) {
    tenantID := tenant.TenantIDFromContext(ctx)

    // TenantDB auto-scopes, but explicit check adds defense-in-depth
    var order Order
    err := tenantDB.QueryRowFromContext(ctx,
        "SELECT * FROM orders WHERE id = ? AND tenant_id = ?",
        orderID, tenantID,
    ).Scan(...)

    return &order, err
}
```

## Related Documentation

- [Tenant Overview](README.md) — Multi-tenancy overview
- [Store: DB](../store/db/) — Database wrapper documentation
- [Examples](examples.md) — Complete database isolation examples
