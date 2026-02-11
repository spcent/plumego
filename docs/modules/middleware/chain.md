# Middleware Chain

> **Package**: `github.com/spcent/plumego/middleware`

Middleware chains allow composing multiple middleware into a single processing pipeline.

---

## Overview

A middleware chain:
- Executes middleware in registration order
- Allows grouping related middleware
- Supports conditional execution
- Enables reusable middleware stacks

---

## Basic Chain

### Creating a Chain

```go
import "github.com/spcent/plumego/middleware"

chain := middleware.NewChain().
    Use(requestIDMiddleware).
    Use(loggingMiddleware).
    Use(recoveryMiddleware)
```

### Applying to Handler

```go
handler := http.HandlerFunc(myHandler)
wrapped := chain.Apply(handler)

// Or directly
app.Get("/users", chain.Apply(http.HandlerFunc(listUsers)))
```

---

## Chain Composition

### Combining Chains

```go
// Base chain
base := middleware.NewChain().
    Use(requestIDMiddleware).
    Use(loggingMiddleware)

// Auth chain (extends base)
auth := base.Append(authMiddleware)

// Admin chain (extends auth)
admin := auth.Append(adminMiddleware)

// Use different chains
app.Get("/public", base.Apply(publicHandler))
app.Get("/api", auth.Apply(apiHandler))
app.Get("/admin", admin.Apply(adminHandler))
```

### Conditional Chains

```go
func buildChain(config Config) *middleware.Chain {
    chain := middleware.NewChain().
        Use(requestIDMiddleware).
        Use(loggingMiddleware)

    if config.EnableAuth {
        chain = chain.Use(authMiddleware)
    }

    if config.EnableRateLimit {
        chain = chain.Use(rateLimitMiddleware)
    }

    return chain
}
```

---

## Global vs. Route Chains

### Global Application

```go
app := core.New()

// Apply to all routes
app.Use(middleware1)
app.Use(middleware2)

app.Get("/users", handler)
app.Get("/products", handler)
// Both routes use middleware1 → middleware2
```

### Per-Route Application

```go
// Create chains
publicChain := middleware.NewChain().Use(loggingMiddleware)
authChain := middleware.NewChain().Use(loggingMiddleware).Use(authMiddleware)

// Apply per route
app.Get("/public", publicChain.Apply(publicHandler))
app.Get("/api", authChain.Apply(apiHandler))
```

---

## Complete Example

```go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware"
)

func main() {
    // Create reusable chains
    baseChain := middleware.NewChain().
        Use(requestIDMiddleware).
        Use(loggingMiddleware).
        Use(recoveryMiddleware)

    authChain := baseChain.
        Append(authMiddleware).
        Append(rateLimitMiddleware)

    adminChain := authChain.
        Append(adminMiddleware)

    app := core.New()

    // Public routes (base chain)
    app.Get("/health", baseChain.Apply(healthHandler))
    app.Get("/status", baseChain.Apply(statusHandler))

    // API routes (auth chain)
    app.Get("/api/profile", authChain.Apply(profileHandler))
    app.Get("/api/orders", authChain.Apply(ordersHandler))

    // Admin routes (admin chain)
    app.Get("/admin/users", adminChain.Apply(adminUsersHandler))
    app.Get("/admin/settings", adminChain.Apply(adminSettingsHandler))

    app.Boot()
}

func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Request-ID", generateID())
        next.ServeHTTP(w, r)
    })
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        log.Printf("[START] %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
        log.Printf("[END] %s %s (%v)", r.Method, r.URL.Path, time.Since(start))
    })
}

func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic recovered: %v", err)
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", 401)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func adminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check if user is admin
        isAdmin := checkAdmin(r)
        if !isAdmin {
            http.Error(w, "Forbidden", 403)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !checkRateLimit(r) {
            http.Error(w, "Rate limit exceeded", 429)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## Best Practices

### ✅ Do

1. **Reuse Chains**
   ```go
   baseChain := middleware.NewChain().Use(common...)
   authChain := baseChain.Append(auth)
   ```

2. **Name Chains Clearly**
   ```go
   publicChain := ...
   authenticatedChain := ...
   adminChain := ...
   ```

3. **Order Matters**
   ```go
   chain := middleware.NewChain().
       Use(requestID).  // First
       Use(logging).    // Second
       Use(auth).       // Third
   ```

### ❌ Don't

1. **Don't Create Circular References**
   ```go
   // ❌ Circular
   chain1 := middleware.NewChain().Use(middleware2)
   chain2 := middleware.NewChain().Use(middleware1)
   ```

2. **Don't Ignore Order**
   ```go
   // ❌ Auth before logging (won't log auth failures)
   chain := middleware.NewChain().
       Use(auth).
       Use(logging)

   // ✅ Logging first
   chain := middleware.NewChain().
       Use(logging).
       Use(auth)
   ```

---

**Next**: [Custom Middleware](custom-middleware.md)
