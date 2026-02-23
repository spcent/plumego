# Middleware Binding

> **Package**: `github.com/spcent/plumego/router`

Middleware can be applied at different levels in Plumego: globally, per-group, or per-route. This document covers how to bind middleware to routes and groups.

---

## Table of Contents

- [Overview](#overview)
- [Global Middleware](#global-middleware)
- [Group Middleware](#group-middleware)
- [Per-Route Middleware](#per-route-middleware)
- [Middleware Order](#middleware-order)
- [Common Patterns](#common-patterns)
- [Best Practices](#best-practices)

---

## Overview

### Middleware Levels

Plumego supports three levels of middleware application:

1. **Global**: Applies to all routes
2. **Group**: Applies to all routes in a group
3. **Per-Route**: Applies to specific routes only

### Middleware Signature

```go
type Middleware func(http.Handler) http.Handler
```

---

## Global Middleware

Global middleware applies to **all routes** in the application.

### Using Built-in Options

```go
app := core.New(
    core.WithRecommendedMiddleware(), // RequestID + Logging + Recovery
)

// Or individual middleware
app := core.New(
    core.WithRequestID(),
    core.WithLogging(),
    core.WithRecovery(),
)
```

### Using Use() Method

```go
app := core.New()

// Add middleware
app.Use(requestIDMiddleware)
app.Use(loggingMiddleware)
app.Use(recoveryMiddleware)

// Register routes (middleware applies to all)
app.Get("/users", listUsers)
app.Get("/products", listProducts)
```

### Custom Global Middleware

```go
func customMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pre-processing
        log.Printf("Request: %s %s", r.Method, r.URL.Path)

        // Call next handler
        next.ServeHTTP(w, r)

        // Post-processing
        log.Println("Response sent")
    })
}

app := core.New()
app.Use(customMiddleware)
```

---

## Group Middleware

Group middleware applies to **all routes in a group**.

### During Group Creation

```go
app := core.New()

// Create group with middleware
api := app.Group("/api", authMiddleware, rateLimitMiddleware)

// All routes in this group have auth + rate limit
api.Get("/users", listUsers)
api.Get("/products", listProducts)
```

### Multiple Middleware

```go
api := app.Group("/api",
    loggingMiddleware,
    authMiddleware,
    rateLimitMiddleware,
    corsMiddleware,
)
```

### Adding Middleware After Creation

```go
api := app.Group("/api")

// Add middleware
api.Use(authMiddleware)
api.Use(rateLimitMiddleware)

// Register routes
api.Get("/users", listUsers)
```

### Nested Group Middleware

```go
app := core.New()

// Level 1: Logging
api := app.Group("/api", loggingMiddleware)

// Level 2: Auth (inherits logging)
v1 := api.Group("/v1", authMiddleware)

// Level 3: Admin (inherits logging + auth)
admin := v1.Group("/admin", adminMiddleware)

// Middleware stack for /api/v1/admin/users:
// loggingMiddleware → authMiddleware → adminMiddleware → handler
```

---

## Per-Route Middleware

Per-route middleware applies to **specific routes only**.

### Wrapping Handler

```go
app := core.New()

// Apply middleware to single route
app.Get("/admin/users",
    adminMiddleware(
        http.HandlerFunc(adminListUsers),
    ),
)

// Route without middleware
app.Get("/public/status", publicStatusHandler)
```

### Multiple Middleware

```go
// Chain multiple middleware
app.Get("/protected",
    middleware1(
        middleware2(
            middleware3(
                http.HandlerFunc(handler),
            ),
        ),
    ),
)
```

### Middleware Helper

```go
func chain(h http.HandlerFunc, middleware ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middleware) - 1; i >= 0; i-- {
        h = middleware[i](h).(http.HandlerFunc)
    }
    return h
}

// Usage
app.Get("/protected",
    chain(handler, authMiddleware, rateLimitMiddleware),
)
```

---

## Middleware Order

### Execution Order

Middleware executes in registration order (outside-in):

```go
app.Use(middleware1)  // Runs first
app.Use(middleware2)  // Runs second
app.Use(middleware3)  // Runs third

app.Get("/test", handler)  // Runs last

// Request flow:
// middleware1 → middleware2 → middleware3 → handler
// Response flow:
// handler → middleware3 → middleware2 → middleware1
```

### Combined Order

```go
// Global middleware
app.Use(globalMiddleware)

// Group middleware
api := app.Group("/api", groupMiddleware)

// Route-specific middleware
api.Get("/protected",
    routeMiddleware(http.HandlerFunc(handler)),
)

// Execution order for /api/protected:
// 1. globalMiddleware
// 2. groupMiddleware
// 3. routeMiddleware
// 4. handler
```

### Example

```go
func logger(name string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log.Printf("[%s] Start", name)
            next.ServeHTTP(w, r)
            log.Printf("[%s] End", name)
        })
    }
}

app := core.New()
app.Use(logger("global"))

api := app.Group("/api", logger("group"))
api.Get("/test", logger("route")(http.HandlerFunc(handler)))

// Output for GET /api/test:
// [global] Start
// [group] Start
// [route] Start
// [route] End
// [group] End
// [global] End
```

---

## Common Patterns

### Authentication Tiers

```go
app := core.New()

// Public routes (no auth)
public := app.Group("/api/public")
public.Get("/status", getStatus)
public.Post("/login", login)

// User routes (basic auth)
user := app.Group("/api", authMiddleware)
user.Get("/profile", getProfile)
user.Get("/orders", listOrders)

// Admin routes (auth + admin)
admin := app.Group("/api/admin", authMiddleware, adminMiddleware)
admin.Get("/users", adminListUsers)
admin.Delete("/users/:id", adminDeleteUser)
```

### Rate Limiting by Route

```go
app := core.New()

// Public routes (aggressive rate limit)
public := app.Group("/api/public", rateLimit(10, time.Minute))
public.Get("/status", getStatus)

// Authenticated routes (generous rate limit)
api := app.Group("/api", authMiddleware, rateLimit(100, time.Minute))
api.Get("/users", listUsers)

// Premium routes (no rate limit)
premium := app.Group("/api/premium", authMiddleware, premiumMiddleware)
premium.Get("/analytics", getAnalytics)
```

### Conditional Middleware

```go
func conditionalAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip auth for health checks
        if r.URL.Path == "/health" {
            next.ServeHTTP(w, r)
            return
        }

        // Require auth for other routes
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

app := core.New()
app.Use(conditionalAuth)
```

### CORS by Route

```go
app := core.New()

// CORS for public API
public := app.Group("/api/public", corsMiddleware(cors.Config{
    AllowOrigins: []string{"*"},
}))
public.Get("/data", publicData)

// Strict CORS for private API
private := app.Group("/api", authMiddleware, corsMiddleware(cors.Config{
    AllowOrigins: []string{"https://example.com"},
    AllowMethods: []string{"GET", "POST"},
}))
private.Get("/users", listUsers)
```

---

## Complete Examples

### Multi-Tier API

```go
package main

import (
    "log"
    "net/http"
    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New(
        // Global middleware
        core.WithRequestID(),
        core.WithLogging(),
        core.WithRecovery(),
    )

    // Public endpoints (no auth)
    app.Get("/health", healthCheck)
    app.Post("/login", login)

    // API v1
    v1 := app.Group("/api/v1")

    // Public v1 routes
    public := v1.Group("/public")
    public.Get("/status", getStatus)
    public.Get("/version", getVersion)

    // Authenticated v1 routes
    auth := v1.Group("",
        authMiddleware,
        rateLimitMiddleware,
    )
    auth.Get("/profile", getProfile)
    auth.Get("/orders", listOrders)

    // Admin v1 routes
    admin := v1.Group("/admin",
        authMiddleware,
        adminMiddleware,
        auditMiddleware,
    )
    admin.Get("/users", adminListUsers)
    admin.Post("/settings", updateSettings)

    app.Boot()
}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        // Validate token...
        next.ServeHTTP(w, r)
    })
}

func adminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check if user is admin...
        isAdmin := true // From token claims
        if !isAdmin {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check rate limit...
        next.ServeHTTP(w, r)
    })
}

func auditMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Audit: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
```

### Content Negotiation

```go
package main

import (
    "net/http"
    "strings"
    "github.com/spcent/plumego/core"
)

func contentNegotiationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        accept := r.Header.Get("Accept")

        // Only allow JSON
        if !strings.Contains(accept, "application/json") {
            http.Error(w, "Only JSON is supported", http.StatusNotAcceptable)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}

func main() {
    app := core.New()

    // API routes require JSON
    api := app.Group("/api", contentNegotiationMiddleware)
    api.Get("/users", listUsers)
    api.Get("/products", listProducts)

    // HTML routes (no middleware)
    app.Get("/", homeHandler)
    app.Get("/about", aboutHandler)

    app.Boot()
}
```

---

## Best Practices

### ✅ Do

1. **Apply Middleware at Appropriate Level**
   ```go
   // ✅ Global for all routes
   app.Use(loggingMiddleware)

   // ✅ Group for related routes
   api := app.Group("/api", authMiddleware)

   // ✅ Per-route for specific needs
   app.Get("/admin", adminMiddleware(handler))
   ```

2. **Order Middleware Logically**
   ```go
   // ✅ Logging → Auth → Business Logic
   app.Use(loggingMiddleware)
   api := app.Group("/api", authMiddleware)
   api.Get("/users", handler)
   ```

3. **Use Groups for Middleware Sharing**
   ```go
   // ✅ Share middleware via groups
   admin := app.Group("/admin", authMiddleware, adminMiddleware)
   admin.Get("/users", handler1)
   admin.Get("/settings", handler2)
   ```

### ❌ Don't

1. **Don't Duplicate Middleware**
   ```go
   // ❌ Redundant
   api := app.Group("/api", authMiddleware)
   api.Get("/users", authMiddleware(handler))

   // ✅ Apply once
   api := app.Group("/api", authMiddleware)
   api.Get("/users", handler)
   ```

2. **Don't Apply Heavy Middleware Globally**
   ```go
   // ❌ Expensive middleware on all routes
   app.Use(heavyAnalyticsMiddleware)

   // ✅ Only on API routes
   api := app.Group("/api", heavyAnalyticsMiddleware)
   ```

3. **Don't Ignore Middleware Order**
   ```go
   // ❌ Auth before logging (won't log auth failures)
   app.Use(authMiddleware)
   app.Use(loggingMiddleware)

   // ✅ Logging first
   app.Use(loggingMiddleware)
   app.Use(authMiddleware)
   ```

---

## Testing Middleware

```go
package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/spcent/plumego/core"
)

func TestMiddlewareBinding(t *testing.T) {
    app := core.New()

    // Middleware that adds header
    testMiddleware := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("X-Test", "middleware-applied")
            next.ServeHTTP(w, r)
        })
    }

    // Route with middleware
    app.Get("/with-middleware",
        testMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
        })),
    )

    // Route without middleware
    app.Get("/without-middleware", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    tests := []struct {
        path        string
        wantHeader  bool
    }{
        {"/with-middleware", true},
        {"/without-middleware", false},
    }

    for _, tt := range tests {
        req := httptest.NewRequest("GET", tt.path, nil)
        w := httptest.NewRecorder()

        app.Router().ServeHTTP(w, req)

        header := w.Header().Get("X-Test")
        hasHeader := header != ""

        if hasHeader != tt.wantHeader {
            t.Errorf("GET %s: header present = %v, want %v", tt.path, hasHeader, tt.wantHeader)
        }
    }
}
```

---

## Next Steps

- **[Middleware Module](../middleware/)** - Built-in middleware
- **[Route Groups](route-groups.md)** - Group organization
- **[Basic Routing](basic-routing.md)** - Route registration

---

**Related**:
- [Router Overview](README.md)
- [Middleware Module](../middleware/)
- [Core Module](../core/)
