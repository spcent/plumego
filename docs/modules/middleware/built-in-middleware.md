# Built-in Middleware Reference

> **Package**: `github.com/spcent/plumego/middleware/*`

Complete reference for all 19 built-in middleware packages.

---

## Authentication & Authorization

### 1. Auth (`middleware/auth`)

JWT and session authentication.

```go
import "github.com/spcent/plumego/middleware/auth"

app := core.New(
    core.WithComponent(&auth.Component{
        JWTSecret: os.Getenv("JWT_SECRET"),
    }),
)
```

---

## Request Processing

### 2. Bind (`middleware/bind`)

Automatic request body binding.

```go
import "github.com/spcent/plumego/middleware/bind"

app.Use(bind.New())
```

### 3. Limits (`middleware/limits`)

Request size and duration limits.

```go
import "github.com/spcent/plumego/middleware/limits"

app.Use(limits.New(limits.Config{
    MaxBodySize:    10 << 20, // 10 MB
    MaxHeaderSize:  1 << 20,  // 1 MB
    RequestTimeout: 30 * time.Second,
}))
```

### 4. Timeout (`middleware/timeout`)

Request timeout handling.

```go
import "github.com/spcent/plumego/middleware/timeout"

app.Use(timeout.New(30 * time.Second))
```

---

## Caching & Performance

### 5. Cache (`middleware/cache`)

HTTP caching with ETag and Cache-Control.

```go
import "github.com/spcent/plumego/middleware/cache"

app.Use(cache.New(cache.Config{
    TTL:          5 * time.Minute,
    CacheControl: "public, max-age=300",
}))
```

### 6. Coalesce (`middleware/coalesce`)

Request coalescing/deduplication.

```go
import "github.com/spcent/plumego/middleware/coalesce"

app.Use(coalesce.New())
```

### 7. Compression (`middleware/compression`)

Response compression (gzip, brotli).

```go
import "github.com/spcent/plumego/middleware/compression"

app := core.New(
    core.WithComponent(&compression.Component{
        Level: compression.DefaultCompression,
    }),
)
```

---

## Resilience

### 8. CircuitBreaker (`middleware/circuitbreaker`)

Circuit breaker pattern for fault tolerance.

```go
import "github.com/spcent/plumego/middleware/circuitbreaker"

app.Use(circuitbreaker.New(circuitbreaker.Config{
    MaxRequests:  5,
    Interval:     10 * time.Second,
    Timeout:      30 * time.Second,
}))
```

### 9. RateLimit (`middleware/ratelimit`)

Rate limiting.

```go
import "github.com/spcent/plumego/middleware/ratelimit"

app := core.New(
    core.WithComponent(&ratelimit.Component{
        RequestsPerSecond: 100,
        Burst:             200,
    }),
)
```

### 10. Recovery (`middleware/recovery`)

Panic recovery.

```go
import "github.com/spcent/plumego/middleware/recovery"

app := core.New(
    core.WithRecovery(),
)
```

---

## Cross-Origin & Protocol

### 11. CORS (`middleware/cors`)

Cross-Origin Resource Sharing.

```go
import "github.com/spcent/plumego/middleware/cors"

app := core.New(
    core.WithCORS(cors.Config{
        AllowOrigins:     []string{"https://example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           3600,
    }),
)
```

### 12. Protocol (`middleware/protocol`)

Protocol adaptation (HTTP, gRPC, GraphQL).

```go
import "github.com/spcent/plumego/middleware/protocol"

app.Use(protocol.New())
```

### 13. Versioning (`middleware/versioning`)

API versioning.

```go
import "github.com/spcent/plumego/middleware/versioning"

app.Use(versioning.New(versioning.Config{
    Header: "API-Version",
    Default: "v1",
}))
```

---

## Observability

### 14. Observability (`middleware/observability`)

Metrics, tracing, and logging.

```go
import "github.com/spcent/plumego/middleware/observability"

app := core.New(
    core.WithComponent(&observability.Component{
        MetricsEnabled: true,
        TracingEnabled: true,
    }),
)
```

---

## Proxy & Routing

### 15. Proxy (`middleware/proxy`)

Reverse proxy.

```go
import "github.com/spcent/plumego/middleware/proxy"

target, _ := url.Parse("http://backend:8080")
app.Use(proxy.New(target))
```

### 16. Tenant (`middleware/tenant`)

Multi-tenancy routing.

```go
import "github.com/spcent/plumego/middleware/tenant"

app := core.New(
    core.WithTenantMiddleware(core.TenantMiddlewareOptions{
        HeaderName: "X-Tenant-ID",
    }),
)
```

---

## Security

### 17. Security (`middleware/security`)

Security headers (CSP, HSTS, etc.).

```go
app := core.New(
    core.WithSecurityHeadersEnabled(true),
)
```

---

## Transformation

### 18. Transform (`middleware/transform`)

Response transformation.

```go
import "github.com/spcent/plumego/middleware/transform"

app.Use(transform.New(transform.Config{
    ContentType: "application/json",
}))
```

---

## Development

### 19. Debug (`middleware/debug`)

Debug utilities.

```go
import "github.com/spcent/plumego/middleware/debug"

if config.Debug {
    app.Use(debug.New())
}
```

---

**Next**: [Error Handling](error-handling.md)
