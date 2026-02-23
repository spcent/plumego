# Route Groups

> **Package**: `github.com/spcent/plumego/router`

Route groups provide hierarchical organization of routes with shared prefixes and middleware. This document covers group creation, nesting, and middleware application.

---

## Table of Contents

- [Overview](#overview)
- [Creating Groups](#creating-groups)
- [Nested Groups](#nested-groups)
- [Group Middleware](#group-middleware)
- [Common Patterns](#common-patterns)
- [Best Practices](#best-practices)

---

## Overview

### What are Route Groups?

Route groups allow you to:
- Share a common path prefix
- Apply middleware to multiple routes
- Organize routes hierarchically
- Version APIs
- Separate concerns (public vs. admin)

### Benefits

- **DRY Principle**: Don't repeat prefixes
- **Organization**: Logical route grouping
- **Middleware**: Shared authentication, logging, etc.
- **Versioning**: Easy API version management

---

## Creating Groups

### Basic Group

```go
app := core.New()

// Create group with prefix
api := app.Group("/api")

// Register routes in group
api.Get("/users", listUsers)        // → /api/users
api.Get("/products", listProducts)  // → /api/products
```

### Group Methods

All HTTP methods are available on groups:

```go
api := app.Group("/api")

api.Get("/users", listUsers)
api.Post("/users", createUser)
api.Put("/users/:id", updateUser)
api.Delete("/users/:id", deleteUser)

// Context-aware variants
api.GetCtx("/products", listProducts)
api.PostCtx("/products", createProduct)
```

### Named Routes in Groups

```go
api := app.Group("/api")

api.GetNamed("api.users.index", "/users", listUsers)
api.GetNamed("api.users.show", "/users/:id", getUser)
```

---

## Nested Groups

Groups can be nested to create hierarchical structures.

### Two Levels

```go
app := core.New()

// Level 1: API group
api := app.Group("/api")

// Level 2: Version groups
v1 := api.Group("/v1")
v1.Get("/users", listUsersV1)  // → /api/v1/users

v2 := api.Group("/v2")
v2.Get("/users", listUsersV2)  // → /api/v2/users
```

### Three Levels

```go
// Level 1: API
api := app.Group("/api")

// Level 2: Version
v1 := api.Group("/v1")

// Level 3: Resource
users := v1.Group("/users")
users.Get("", listUsers)           // → /api/v1/users
users.Get("/:id", getUser)         // → /api/v1/users/:id
users.Post("", createUser)         // → /api/v1/users
users.Put("/:id", updateUser)      // → /api/v1/users/:id
users.Delete("/:id", deleteUser)   // → /api/v1/users/:id
```

### Complex Hierarchy

```go
app := core.New()

// Public API
public := app.Group("/api/public")
public.Get("/status", getStatus)
public.Get("/version", getVersion)

// Authenticated API
api := app.Group("/api")
v1 := api.Group("/v1")

// User resources
users := v1.Group("/users")
users.Get("", listUsers)
users.Get("/:id", getUser)

// Admin resources
admin := v1.Group("/admin")
admin.Get("/dashboard", adminDashboard)
admin.Get("/users", adminListUsers)

// Results:
// /api/public/status
// /api/public/version
// /api/v1/users
// /api/v1/users/:id
// /api/v1/admin/dashboard
// /api/v1/admin/users
```

---

## Group Middleware

Apply middleware to all routes in a group.

### Basic Middleware

```go
// Middleware function
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

// Apply to group
api := app.Group("/api", loggingMiddleware)
api.Get("/users", listUsers)  // Logging applies
```

### Multiple Middleware

```go
api := app.Group("/api",
    loggingMiddleware,
    authMiddleware,
    rateLimitMiddleware,
)
```

### Middleware Order

Middleware executes in registration order:

```go
api := app.Group("/api",
    middleware1,  // Runs first
    middleware2,  // Runs second
    middleware3,  // Runs third
)
```

### Adding Middleware to Existing Group

```go
api := app.Group("/api")

// Add middleware later
api.Use(loggingMiddleware)
api.Use(authMiddleware)

// Register routes
api.Get("/users", listUsers)
```

---

## Common Patterns

### API Versioning

```go
app := core.New()

// Version 1
v1 := app.Group("/api/v1")
v1.Get("/users", listUsersV1)
v1.Get("/products", listProductsV1)

// Version 2
v2 := app.Group("/api/v2")
v2.Get("/users", listUsersV2)
v2.Get("/products", listProductsV2)

// Version 3 (latest)
v3 := app.Group("/api/v3")
v3.Get("/users", listUsersV3)
v3.Get("/products", listProductsV3)

// Default to latest version
app.Get("/api/users", listUsersV3)
```

### Public vs. Authenticated

```go
// Public routes (no auth)
public := app.Group("/api/public")
public.Get("/status", getStatus)
public.Post("/login", login)
public.Post("/register", register)

// Authenticated routes
api := app.Group("/api", authMiddleware)
api.Get("/profile", getProfile)
api.Get("/orders", listOrders)
api.Post("/orders", createOrder)

// Admin routes (auth + admin check)
admin := app.Group("/api/admin", authMiddleware, adminMiddleware)
admin.Get("/users", adminListUsers)
admin.Delete("/users/:id", adminDeleteUser)
```

### Language/Locale Groups

```go
// English
en := app.Group("/en")
en.Get("/", homeEN)
en.Get("/about", aboutEN)

// Chinese
zh := app.Group("/zh")
zh.Get("/", homeZH)
zh.Get("/about", aboutZH)

// French
fr := app.Group("/fr")
fr.Get("/", homeFR)
fr.Get("/about", aboutFR)
```

### Tenant-Based Routing

```go
// Multi-tenant SaaS
tenant := app.Group("/tenants/:tenant_id", tenantMiddleware)
tenant.Get("/dashboard", tenantDashboard)
tenant.Get("/users", tenantListUsers)
tenant.Get("/settings", tenantSettings)

// Example URLs:
// /tenants/acme-corp/dashboard
// /tenants/acme-corp/users
// /tenants/widgets-inc/dashboard
```

### RESTful Resource Groups

```go
// Users resource
users := app.Group("/users")
users.Get("", listUsers)               // GET /users
users.Post("", createUser)             // POST /users
users.Get("/:id", getUser)             // GET /users/:id
users.Put("/:id", updateUser)          // PUT /users/:id
users.Delete("/:id", deleteUser)       // DELETE /users/:id

// Nested: User's posts
posts := users.Group("/:user_id/posts")
posts.Get("", listUserPosts)           // GET /users/:user_id/posts
posts.Get("/:id", getUserPost)         // GET /users/:user_id/posts/:id
posts.Post("", createUserPost)         // POST /users/:user_id/posts
```

---

## Complete Examples

### API Gateway

```go
package main

import (
    "net/http"
    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New()

    // Health endpoints (no auth)
    health := app.Group("/health")
    health.Get("", healthCheck)
    health.Get("/ready", readinessCheck)
    health.Get("/live", livenessCheck)

    // Public API
    public := app.Group("/api/public")
    public.Post("/login", login)
    public.Post("/register", register)
    public.Get("/status", getStatus)

    // Authenticated API v1
    v1 := app.Group("/api/v1", authMiddleware, rateLimitMiddleware)

    // User endpoints
    users := v1.Group("/users")
    users.Get("", listUsers)
    users.Get("/:id", getUser)
    users.Put("/:id", updateUser)
    users.Delete("/:id", deleteUser)

    // Product endpoints
    products := v1.Group("/products")
    products.Get("", listProducts)
    products.Get("/:id", getProduct)
    products.Post("", createProduct)
    products.Put("/:id", updateProduct)
    products.Delete("/:id", deleteProduct)

    // Order endpoints
    orders := v1.Group("/orders")
    orders.Get("", listOrders)
    orders.Get("/:id", getOrder)
    orders.Post("", createOrder)
    orders.Put("/:id/cancel", cancelOrder)

    // Admin endpoints (additional admin middleware)
    admin := v1.Group("/admin", adminMiddleware)
    admin.Get("/dashboard", adminDashboard)
    admin.Get("/users", adminListUsers)
    admin.Get("/analytics", adminAnalytics)
    admin.Post("/settings", updateSettings)

    app.Boot()
}
```

### Multi-Tenant SaaS

```go
package main

import (
    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New()

    // Global endpoints
    app.Get("/", homepage)
    app.Post("/signup", signup)

    // Tenant-scoped endpoints
    tenant := app.Group("/t/:tenant_id",
        tenantResolverMiddleware,
        tenantAuthMiddleware,
    )

    // Tenant dashboard
    tenant.Get("/dashboard", tenantDashboard)

    // Tenant users
    users := tenant.Group("/users")
    users.Get("", listTenantUsers)
    users.Post("", createTenantUser)
    users.Get("/:id", getTenantUser)
    users.Put("/:id", updateTenantUser)
    users.Delete("/:id", deleteTenantUser)

    // Tenant settings
    settings := tenant.Group("/settings")
    settings.Get("", getTenantSettings)
    settings.Put("", updateTenantSettings)

    // Tenant billing
    billing := tenant.Group("/billing")
    billing.Get("/invoices", listInvoices)
    billing.Get("/invoices/:id", getInvoice)
    billing.Post("/payment-method", updatePaymentMethod)

    app.Boot()
}
```

### Microservices Gateway

```go
package main

import (
    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New()

    // User service
    userService := app.Group("/services/users",
        serviceAuthMiddleware,
        proxyMiddleware("http://users-service:8080"),
    )
    userService.Get("/*path", proxyHandler)
    userService.Post("/*path", proxyHandler)
    userService.Put("/*path", proxyHandler)
    userService.Delete("/*path", proxyHandler)

    // Product service
    productService := app.Group("/services/products",
        serviceAuthMiddleware,
        proxyMiddleware("http://products-service:8080"),
    )
    productService.Get("/*path", proxyHandler)
    productService.Post("/*path", proxyHandler)

    // Order service
    orderService := app.Group("/services/orders",
        serviceAuthMiddleware,
        proxyMiddleware("http://orders-service:8080"),
    )
    orderService.Get("/*path", proxyHandler)
    orderService.Post("/*path", proxyHandler)

    app.Boot()
}
```

---

## Best Practices

### ✅ Do

1. **Use Groups for Organization**
   ```go
   // ✅ Organized
   api := app.Group("/api")
   api.Get("/users", listUsers)
   api.Get("/products", listProducts)
   ```

2. **Version Your APIs**
   ```go
   v1 := app.Group("/api/v1")
   v2 := app.Group("/api/v2")
   ```

3. **Apply Middleware at Group Level**
   ```go
   // ✅ Efficient
   api := app.Group("/api", authMiddleware)
   api.Get("/users", listUsers)
   api.Get("/products", listProducts)
   ```

4. **Use Nested Groups for Hierarchy**
   ```go
   api := app.Group("/api")
   v1 := api.Group("/v1")
   users := v1.Group("/users")
   ```

### ❌ Don't

1. **Don't Create Unnecessary Nesting**
   ```go
   // ❌ Too deep
   a := app.Group("/a")
   b := a.Group("/b")
   c := b.Group("/c")
   d := c.Group("/d")

   // ✅ Keep it reasonable
   api := app.Group("/api/v1")
   api.Get("/resource", handler)
   ```

2. **Don't Duplicate Middleware**
   ```go
   // ❌ Redundant
   api := app.Group("/api", authMiddleware)
   api.Get("/users", authMiddleware(listUsers))

   // ✅ Apply once at group level
   api := app.Group("/api", authMiddleware)
   api.Get("/users", listUsers)
   ```

3. **Don't Mix Concerns in One Group**
   ```go
   // ❌ Mixed public and private
   api := app.Group("/api")
   api.Get("/status", publicStatus)  // Should be public
   api.Get("/profile", getProfile)   // Needs auth

   // ✅ Separate groups
   public := app.Group("/api/public")
   public.Get("/status", publicStatus)

   api := app.Group("/api", authMiddleware)
   api.Get("/profile", getProfile)
   ```

---

## Testing Groups

```go
package main

import (
    "net/http/httptest"
    "testing"
    "github.com/spcent/plumego/core"
)

func TestGroups(t *testing.T) {
    app := core.New()

    api := app.Group("/api")
    v1 := api.Group("/v1")
    v1.Get("/users", listUsers)

    tests := []struct {
        path string
        want int
    }{
        {"/api/v1/users", 200},
        {"/api/users", 404},
        {"/v1/users", 404},
    }

    for _, tt := range tests {
        req := httptest.NewRequest("GET", tt.path, nil)
        w := httptest.NewRecorder()

        app.Router().ServeHTTP(w, req)

        if w.Code != tt.want {
            t.Errorf("GET %s: got %d, want %d", tt.path, w.Code, tt.want)
        }
    }
}
```

---

## Next Steps

- **[Path Parameters](path-parameters.md)** - Extract URL parameters
- **[Middleware Binding](middleware-binding.md)** - Advanced middleware patterns
- **[Basic Routing](basic-routing.md)** - Route registration fundamentals

---

**Related**:
- [Router Overview](README.md)
- [Middleware Module](../middleware/)
- [Core Module](../core/)
