# Middleware Module

> **Package**: `github.com/spcent/plumego/middleware`
> **Stability**: High - Middleware API is stable
> **Go Version**: 1.24+

The `middleware` package provides request processing middleware with 19 built-in middleware and support for custom middleware chains.

---

## Overview

Middleware intercepts HTTP requests before they reach handlers, enabling:

- **Request Processing**: Logging, authentication, rate limiting
- **Response Modification**: Compression, headers, transformation
- **Error Handling**: Recovery, error logging
- **Cross-Cutting Concerns**: CORS, security headers, observability

---

## Quick Start

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware/auth"
    "github.com/spcent/plumego/middleware/ratelimit"
)

app := core.New(
    // Built-in middleware
    core.WithRecommendedMiddleware(), // RequestID + Logging + Recovery

    // Additional middleware
    core.WithCORS(cors.DefaultConfig()),

    // Custom middleware
    core.WithComponent(&auth.Component{}),
)
```

---

## Middleware Signature

```go
type Middleware func(http.Handler) http.Handler
```

### Example

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before handler
        log.Printf("%s %s", r.Method, r.URL.Path)

        // Call next handler
        next.ServeHTTP(w, r)

        // After handler
        log.Println("Request completed")
    })
}
```

---

## Built-in Middleware (19)

### 1. Authentication & Authorization

| Middleware | Package | Description |
|------------|---------|-------------|
| **Auth** | `middleware/auth` | JWT/session authentication |

### 2. Request Processing

| Middleware | Package | Description |
|------------|---------|-------------|
| **Bind** | `middleware/bind` | Automatic request binding |
| **Limits** | `middleware/limits` | Request size/duration limits |
| **Timeout** | `middleware/timeout` | Request timeout handling |

### 3. Caching & Performance

| Middleware | Package | Description |
|------------|---------|-------------|
| **Cache** | `middleware/cache` | HTTP caching (ETag, Cache-Control) |
| **Coalesce** | `middleware/coalesce` | Request coalescing/deduplication |
| **Compression** | `middleware/compression` | Response compression (gzip, brotli) |

### 4. Resilience

| Middleware | Package | Description |
|------------|---------|-------------|
| **CircuitBreaker** | `middleware/circuitbreaker` | Circuit breaker pattern |
| **RateLimit** | `middleware/ratelimit` | Rate limiting |
| **Recovery** | `middleware/recovery` | Panic recovery |

### 5. Cross-Origin & Protocol

| Middleware | Package | Description |
|------------|---------|-------------|
| **CORS** | `middleware/cors` | Cross-Origin Resource Sharing |
| **Protocol** | `middleware/protocol` | Protocol adaptation |
| **Versioning** | `middleware/versioning` | API versioning |

### 6. Observability

| Middleware | Package | Description |
|------------|---------|-------------|
| **Observability** | `middleware/observability` | Metrics, tracing, logging |

### 7. Proxy & Routing

| Middleware | Package | Description |
|------------|---------|-------------|
| **Proxy** | `middleware/proxy` | Reverse proxy |
| **Tenant** | `middleware/tenant` | Multi-tenancy routing |

### 8. Security

| Middleware | Package | Description |
|------------|---------|-------------|
| **Security** | `middleware/security` | Security headers (CSP, HSTS, etc.) |

### 9. Transformation

| Middleware | Package | Description |
|------------|---------|-------------|
| **Transform** | `middleware/transform` | Response transformation |

### 10. Development

| Middleware | Package | Description |
|------------|---------|-------------|
| **Debug** | `middleware/debug` | Debug utilities |

---

## Documentation Structure

| Document | Description |
|----------|-------------|
| **[Chain](chain.md)** | Middleware chain and composition |
| **[Custom Middleware](custom-middleware.md)** | Building custom middleware |
| **[Execution Order](execution-order.md)** | Middleware ordering rules |
| **[Built-in Middleware](built-in-middleware.md)** | All 19 middleware detailed |
| **[Error Handling](error-handling.md)** | Error handling in middleware |
| **[Testing](testing-middleware.md)** | Testing middleware |
| **[Best Practices](best-practices.md)** | Patterns and anti-patterns |

---

## Usage Examples

### Recommended Stack

```go
app := core.New(
    core.WithRecommendedMiddleware(), // RequestID + Logging + Recovery
)
```

Equivalent to:
```go
app := core.New()
app.Use(requestid.New())
app.Use(logging.New())
app.Use(recovery.New())
```

### Production Stack

```go
import (
    "github.com/spcent/plumego/middleware/cors"
    "github.com/spcent/plumego/middleware/compression"
    "github.com/spcent/plumego/middleware/ratelimit"
    "github.com/spcent/plumego/middleware/security"
)

app := core.New(
    // Identity & Logging
    core.WithRequestID(),
    core.WithLogging(),

    // Security
    core.WithSecurityHeadersEnabled(true),
    core.WithCORS(cors.Config{
        AllowOrigins: []string{"https://example.com"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    }),

    // Performance
    core.WithComponent(&compression.Component{}),
    core.WithComponent(&ratelimit.Component{
        RequestsPerSecond: 100,
    }),

    // Resilience
    core.WithRecovery(),
)
```

### Custom Middleware

```go
func customMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pre-processing
        start := time.Now()

        // Call next
        next.ServeHTTP(w, r)

        // Post-processing
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
    })
}

app := core.New()
app.Use(customMiddleware)
```

---

## Middleware Chain

### Global Middleware

```go
app := core.New()

// Applied to all routes
app.Use(middleware1)
app.Use(middleware2)
app.Use(middleware3)
```

### Group Middleware

```go
api := app.Group("/api", authMiddleware, rateLimitMiddleware)
api.Get("/users", listUsers) // Both middleware apply
```

### Per-Route Middleware

```go
app.Get("/admin",
    adminMiddleware(
        http.HandlerFunc(adminHandler),
    ),
)
```

### Execution Flow

```
Request
  ↓
middleware1 (before)
  ↓
middleware2 (before)
  ↓
middleware3 (before)
  ↓
handler
  ↓
middleware3 (after)
  ↓
middleware2 (after)
  ↓
middleware1 (after)
  ↓
Response
```

---

## Common Patterns

### Authentication Flow

```go
app := core.New(
    core.WithRequestID(),
    core.WithLogging(),
)

// Public routes
public := app.Group("/api/public")
public.Get("/status", getStatus)

// Protected routes
api := app.Group("/api", authMiddleware)
api.Get("/profile", getProfile)

// Admin routes
admin := app.Group("/api/admin", authMiddleware, adminMiddleware)
admin.Get("/users", adminListUsers)
```

### Rate Limiting Tiers

```go
import "github.com/spcent/plumego/middleware/ratelimit"

// Public API (strict)
public := app.Group("/api/public", ratelimit.New(10, time.Minute))

// Authenticated (generous)
api := app.Group("/api", authMiddleware, ratelimit.New(100, time.Minute))

// Premium (no limit)
premium := app.Group("/api/premium", authMiddleware, premiumMiddleware)
```

### CORS Configuration

```go
import "github.com/spcent/plumego/middleware/cors"

app := core.New(
    core.WithCORS(cors.Config{
        AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        ExposeHeaders:    []string{"X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           3600,
    }),
)
```

---

## Performance Considerations

### Middleware Order Matters

```go
// ✅ Efficient order
app.Use(requestID)       // Fast
app.Use(rateLimit)       // Fast, rejects early
app.Use(authentication)  // Medium
app.Use(logging)         // Medium
app.Use(compression)     // Expensive, only after handler
app.Use(recovery)        // Fast

// ❌ Inefficient order
app.Use(compression)     // Expensive, runs even for rejected requests
app.Use(rateLimit)       // Should be earlier
```

### Conditional Middleware

```go
func conditionalMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip expensive operation for health checks
        if r.URL.Path == "/health" {
            next.ServeHTTP(w, r)
            return
        }

        // Expensive operation only when needed
        doExpensiveWork()
        next.ServeHTTP(w, r)
    })
}
```

---

## Security Best Practices

### 1. Security Headers

```go
app := core.New(
    core.WithSecurityHeadersEnabled(true),
)
```

Adds:
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000`
- `Content-Security-Policy: default-src 'self'`

### 2. CORS Configuration

```go
// ✅ Specific origins
AllowOrigins: []string{"https://example.com"}

// ❌ Wildcard in production
AllowOrigins: []string{"*"}
```

### 3. Rate Limiting

```go
import "github.com/spcent/plumego/middleware/ratelimit"

app := core.New(
    core.WithComponent(&ratelimit.Component{
        RequestsPerSecond: 100,
        Burst:             200,
    }),
)
```

---

## Troubleshooting

### Middleware Not Executing

```go
// ❌ Middleware added after routes
app.Get("/users", handler)
app.Use(loggingMiddleware) // Won't apply to existing routes!

// ✅ Middleware before routes
app.Use(loggingMiddleware)
app.Get("/users", handler)
```

### Response Already Written

```go
// ❌ Writing response twice
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Middleware response"))
        next.ServeHTTP(w, r) // Handler also writes!
    })
}

// ✅ Conditional writing
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !authorized {
            w.Write([]byte("Unauthorized"))
            return // Don't call next
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## Next Steps

- **[Chain](chain.md)** - Middleware chain composition
- **[Custom Middleware](custom-middleware.md)** - Build your own
- **[Execution Order](execution-order.md)** - Ordering rules
- **[Built-in Middleware](built-in-middleware.md)** - Complete reference

---

## Related Modules

- **[Core](../core/)** - Application configuration
- **[Router](../router/)** - Middleware binding
- **[Contract](../contract/)** - Request/response handling

---

**Stability**: High - Breaking changes require major version bump
**Maintainers**: Plumego Core Team
**Last Updated**: 2026-02-11
