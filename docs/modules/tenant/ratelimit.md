# Tenant Rate Limiting

> **Package**: `github.com/spcent/plumego/tenant` | **Feature**: Per-tenant request rate limiting

## Overview

Per-tenant rate limiting ensures fair usage among tenants, preventing any single tenant from degrading service for others. Unlike global rate limiting, tenant rate limits scale with the tenant's plan.

## RateLimiter Interface

```go
type RateLimiter interface {
    Allow(ctx context.Context, tenantID string) bool
    AllowN(ctx context.Context, tenantID string, n int) bool
}
```

## Setup

### In-Memory Rate Limiter

```go
import "github.com/spcent/plumego/tenant"

// Create limiter using tenant quota config
rateLimiter := tenant.NewConfigRateLimiter(configMgr)

app := core.New(
    core.WithTenantConfigManager(configMgr),
    core.WithTenantMiddleware(core.TenantMiddlewareOptions{
        HeaderName:  "X-Tenant-ID",
        RateLimiter: rateLimiter,
    }),
)
```

### Custom Rate Limiter

```go
type TenantRateLimiter struct {
    limiters map[string]*rate.Limiter
    configMgr tenant.ConfigManager
    mu       sync.RWMutex
}

func (l *TenantRateLimiter) getLimiter(ctx context.Context, tenantID string) *rate.Limiter {
    l.mu.RLock()
    limiter, ok := l.limiters[tenantID]
    l.mu.RUnlock()

    if !ok {
        config, _ := l.configMgr.GetTenantConfig(ctx, tenantID)
        rps := rate.Limit(float64(config.Quota.RequestsPerMinute) / 60.0)
        limiter = rate.NewLimiter(rps, config.Quota.RequestsPerMinute/10) // burst = 10% of minute quota

        l.mu.Lock()
        l.limiters[tenantID] = limiter
        l.mu.Unlock()
    }

    return limiter
}

func (l *TenantRateLimiter) Allow(ctx context.Context, tenantID string) bool {
    return l.getLimiter(ctx, tenantID).Allow()
}
```

## Rate Limit by Plan

```go
var planLimits = map[string]int{
    "free":       60,    // 60 req/min
    "starter":    300,   // 300 req/min
    "pro":        1000,  // 1000 req/min
    "enterprise": 10000, // 10000 req/min
}

func createLimiterForTenant(tenantID, plan string) *rate.Limiter {
    rpm := planLimits[plan]
    rps := rate.Limit(float64(rpm) / 60.0)
    burst := rpm / 5 // Allow burst of 20% of minute limit

    return rate.NewLimiter(rps, burst)
}
```

## Rate Limit Middleware

```go
func tenantRateLimitMiddleware(rateLimiter tenant.RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenantID := tenant.TenantIDFromContext(r.Context())

            if !rateLimiter.Allow(r.Context(), tenantID) {
                // Add rate limit headers
                w.Header().Set("X-RateLimit-Limit", "1000")
                w.Header().Set("X-RateLimit-Remaining", "0")
                w.Header().Set("Retry-After", "60")

                http.Error(w, "Rate limit exceeded. Please retry after 60 seconds.", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Endpoint-Specific Limits

Apply different limits per endpoint:

```go
type EndpointLimiter struct {
    limits    map[string]float64 // endpoint -> multiplier
    base      tenant.RateLimiter
}

func (l *EndpointLimiter) AllowEndpoint(ctx context.Context, tenantID, endpoint string) bool {
    multiplier, ok := l.limits[endpoint]
    if !ok {
        multiplier = 1.0
    }

    // Use higher cost for expensive endpoints
    n := int(1.0 / multiplier)
    if n < 1 {
        n = 1
    }

    return l.base.AllowN(ctx, tenantID, n)
}

// Usage
limiter := &EndpointLimiter{
    limits: map[string]float64{
        "/api/export": 0.1,   // Costs 10x (expensive)
        "/api/search": 0.5,   // Costs 2x
        "/api/users":  1.0,   // Normal cost
        "/api/health": 100.0, // Free
    },
}
```

## Distributed Rate Limiting

For multi-instance deployments, use Redis:

```go
import "github.com/go-redis/redis_rate/v10"

type RedisRateLimiter struct {
    limiter   *redis_rate.Limiter
    configMgr tenant.ConfigManager
}

func (l *RedisRateLimiter) Allow(ctx context.Context, tenantID string) bool {
    config, _ := l.configMgr.GetTenantConfig(ctx, tenantID)
    rpm := config.Quota.RequestsPerMinute

    res, err := l.limiter.Allow(ctx, "tenant:"+tenantID, redis_rate.PerMinute(rpm))
    if err != nil {
        return true // Fail open on Redis error
    }

    return res.Allowed > 0
}
```

## Response Headers

```go
// Standard rate limit headers
w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rpm))
w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

// If exceeded
w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))
```

## Best Practices

### 1. Use Token Bucket for Bursting

Allow short bursts to handle traffic spikes without false positives.

### 2. Fail Open on Errors

```go
func (l *RateLimiter) Allow(ctx context.Context, tenantID string) bool {
    result, err := l.checkRedis(ctx, tenantID)
    if err != nil {
        log.Error("Rate limit check failed, allowing request", "error", err)
        return true // Fail open
    }
    return result
}
```

### 3. Log Violations

```go
if !rateLimiter.Allow(ctx, tenantID) {
    log.Warn("Rate limit exceeded",
        "tenant_id", tenantID,
        "path", r.URL.Path,
        "method", r.Method,
    )
}
```

## Related Documentation

- [Tenant Overview](README.md) — Multi-tenancy overview
- [Quota Management](quota.md) — Resource quotas
- [Middleware: Rate Limiting](../middleware/ratelimit.md) — Global rate limiting
