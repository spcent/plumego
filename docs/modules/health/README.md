# Health Module

> **Package Path**: `github.com/spcent/plumego/health` | **Stability**: High | **Priority**: P1

## Overview

The `health/` package provides health check endpoints for monitoring application status, essential for Kubernetes, Docker, and load balancer integrations. Health checks enable orchestration systems to detect failures and route traffic appropriately.

**Key Features**:
- **Liveness Probes**: Detect if application is running
- **Readiness Probes**: Detect if application can handle traffic
- **Custom Health Checks**: Register application-specific checks
- **Health Status API**: Structured health information
- **Graceful Degradation**: Partial failure handling

## Quick Start

### Basic Health Checks

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/health"
)

app := core.New(
    core.WithAddr(":8080"),
)

// Liveness endpoint: Is app running?
app.Get("/health/live", health.LivenessHandler())

// Readiness endpoint: Can app handle traffic?
app.Get("/health/ready", health.ReadinessHandler())

app.Boot()
```

### With Custom Checks

```go
// Create health checker
checker := health.NewChecker()

// Add database check
checker.AddReadiness("database", func() health.HealthStatus {
    if err := db.Ping(); err != nil {
        return health.StatusUnhealthy("database unreachable")
    }
    return health.StatusHealthy()
})

// Add cache check
checker.AddReadiness("cache", func() health.HealthStatus {
    if err := cache.Ping(); err != nil {
        return health.StatusUnhealthy("cache unreachable")
    }
    return health.StatusHealthy()
})

// Register handlers
app.Get("/health/live", checker.LivenessHandler())
app.Get("/health/ready", checker.ReadinessHandler())
```

## Core Types

### HealthStatus

```go
type HealthStatus struct {
    Status  string                 `json:"status"`  // "healthy", "unhealthy", "degraded"
    Message string                 `json:"message,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

**Status Values**:
- `healthy`: Component is functioning normally
- `unhealthy`: Component has failed
- `degraded`: Component is partially functional

### HealthCheck

Function signature for health checks:

```go
type HealthCheck func() HealthStatus
```

### Checker

Main health checker instance:

```go
type Checker struct {
    // ... internal fields
}

func NewChecker() *Checker
func (c *Checker) AddLiveness(name string, check HealthCheck)
func (c *Checker) AddReadiness(name string, check HealthCheck)
func (c *Checker) LivenessHandler() http.HandlerFunc
func (c *Checker) ReadinessHandler() http.HandlerFunc
func (c *Checker) HealthHandler() http.HandlerFunc
```

## Health Check Types

### Liveness Checks

**Purpose**: Is the application running?

**Typical Checks**:
- Application process is alive
- No deadlocks or hangs
- Core goroutines are responsive

**Kubernetes Action**: Restart pod if liveness fails

**Example**:
```go
checker.AddLiveness("app", func() health.HealthStatus {
    // Simple: Just return healthy (process is alive)
    return health.StatusHealthy()
})
```

### Readiness Checks

**Purpose**: Can the application handle traffic?

**Typical Checks**:
- Database connectivity
- Cache availability
- External service dependencies
- Initialization complete

**Kubernetes Action**: Remove pod from service endpoints if readiness fails

**Example**:
```go
checker.AddReadiness("database", func() health.HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return health.StatusUnhealthy(fmt.Sprintf("db ping failed: %v", err))
    }

    return health.StatusHealthy()
})
```

## Response Format

### Healthy Response

```json
{
  "status": "healthy",
  "checks": {
    "database": {
      "status": "healthy"
    },
    "cache": {
      "status": "healthy"
    }
  }
}
```

HTTP Status: `200 OK`

### Unhealthy Response

```json
{
  "status": "unhealthy",
  "checks": {
    "database": {
      "status": "healthy"
    },
    "cache": {
      "status": "unhealthy",
      "message": "connection timeout"
    }
  }
}
```

HTTP Status: `503 Service Unavailable`

### Degraded Response

```json
{
  "status": "degraded",
  "checks": {
    "database": {
      "status": "healthy"
    },
    "cache": {
      "status": "degraded",
      "message": "reduced capacity",
      "details": {
        "available_nodes": 2,
        "total_nodes": 5
      }
    }
  }
}
```

HTTP Status: `200 OK` (still serving traffic)

## Kubernetes Integration

### Deployment YAML

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        ports:
        - containerPort: 8080

        # Liveness probe
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 1
          failureThreshold: 3

        # Readiness probe
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 1
          failureThreshold: 3
```

### Probe Configuration

| Parameter | Liveness | Readiness | Description |
|-----------|----------|-----------|-------------|
| `initialDelaySeconds` | 10-30 | 5-10 | Wait before first check |
| `periodSeconds` | 10-30 | 5-10 | Check interval |
| `timeoutSeconds` | 1-5 | 1-5 | Check timeout |
| `failureThreshold` | 3-5 | 2-3 | Failures before action |

## Common Health Checks

### 1. Database Connectivity

```go
checker.AddReadiness("database", func() health.HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return health.StatusUnhealthy(fmt.Sprintf("db unreachable: %v", err))
    }

    // Check connection pool
    stats := db.Stats()
    if stats.OpenConnections == stats.MaxOpenConnections {
        return health.StatusDegraded("connection pool exhausted")
    }

    return health.StatusHealthy()
})
```

### 2. Cache Availability

```go
checker.AddReadiness("cache", func() health.HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    if err := cache.Ping(ctx); err != nil {
        return health.StatusUnhealthy("cache unreachable")
    }

    return health.StatusHealthy()
})
```

### 3. External API Dependency

```go
checker.AddReadiness("payment-api", func() health.HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.payment.com/health", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return health.StatusUnhealthy("payment API unreachable")
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return health.StatusUnhealthy(fmt.Sprintf("payment API status: %d", resp.StatusCode))
    }

    return health.StatusHealthy()
})
```

### 4. Disk Space

```go
checker.AddLiveness("disk-space", func() health.HealthStatus {
    var stat syscall.Statfs_t
    if err := syscall.Statfs("/", &stat); err != nil {
        return health.StatusUnhealthy("cannot check disk space")
    }

    available := stat.Bavail * uint64(stat.Bsize)
    total := stat.Blocks * uint64(stat.Bsize)
    percentUsed := float64(total-available) / float64(total) * 100

    if percentUsed > 95 {
        return health.StatusUnhealthy(fmt.Sprintf("disk %.1f%% full", percentUsed))
    }

    if percentUsed > 85 {
        return health.StatusDegraded(fmt.Sprintf("disk %.1f%% full", percentUsed))
    }

    return health.StatusHealthy()
})
```

### 5. Memory Usage

```go
checker.AddLiveness("memory", func() health.HealthStatus {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    allocMB := m.Alloc / 1024 / 1024
    sysMB := m.Sys / 1024 / 1024

    if allocMB > 1024 { // 1 GB
        return health.StatusDegraded(fmt.Sprintf("high memory usage: %d MB", allocMB))
    }

    return health.StatusHealthy()
})
```

## Component Integration

Components can register their own health checks:

```go
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
}

// Example component
type DatabaseComponent struct {
    db *sql.DB
}

func (c *DatabaseComponent) Health() (string, health.HealthStatus) {
    if err := c.db.Ping(); err != nil {
        return "database", health.StatusUnhealthy("ping failed")
    }
    return "database", health.StatusHealthy()
}

// Register component health checks
for _, component := range app.Components() {
    name, status := component.Health()
    checker.AddReadiness(name, func() health.HealthStatus {
        _, status := component.Health()
        return status
    })
}
```

## Best Practices

### 1. Keep Checks Fast

```go
// ✅ Good: Fast check with timeout
checker.AddReadiness("db", func() health.HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return health.StatusUnhealthy("ping failed")
    }
    return health.StatusHealthy()
})

// ❌ Bad: Slow check without timeout
checker.AddReadiness("db", func() health.HealthStatus {
    if err := db.Ping(); err != nil { // May hang
        return health.StatusUnhealthy("ping failed")
    }
    return health.StatusHealthy()
})
```

### 2. Separate Liveness and Readiness

```go
// ✅ Liveness: Simple check (is app alive?)
checker.AddLiveness("app", func() health.HealthStatus {
    return health.StatusHealthy()
})

// ✅ Readiness: Dependency checks (can app serve traffic?)
checker.AddReadiness("database", dbCheck)
checker.AddReadiness("cache", cacheCheck)
```

### 3. Use Degraded Status Appropriately

```go
checker.AddReadiness("cache", func() health.HealthStatus {
    if err := cache.Ping(); err != nil {
        // Cache down: degraded (can still serve with DB only)
        return health.StatusDegraded("cache unavailable")
    }
    return health.StatusHealthy()
})
```

### 4. Don't Include Non-Critical Dependencies

```go
// ❌ Bad: Analytics service is not critical
checker.AddReadiness("analytics", analyticsCheck)

// ✅ Good: Only check critical dependencies
checker.AddReadiness("database", dbCheck)
checker.AddReadiness("auth-service", authCheck)
```

## Module Documentation

Detailed documentation for health check features:

- **[Liveness Probes](liveness.md)** — Liveness check configuration and patterns
- **[Readiness Probes](readiness.md)** — Readiness check configuration and patterns

## Testing

### Unit Tests

```go
func TestHealthChecks(t *testing.T) {
    checker := health.NewChecker()

    // Add test check
    checker.AddReadiness("test", func() health.HealthStatus {
        return health.StatusHealthy()
    })

    // Test readiness handler
    req := httptest.NewRequest("GET", "/health/ready", nil)
    rec := httptest.NewRecorder()

    checker.ReadinessHandler()(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", rec.Code)
    }

    var response map[string]interface{}
    json.NewDecoder(rec.Body).Decode(&response)

    if response["status"] != "healthy" {
        t.Errorf("Expected healthy, got %v", response["status"])
    }
}
```

## Related Documentation

- [Core: Components](../core/components.md) — Component health checks
- [Liveness Probes](liveness.md) — Liveness configuration
- [Readiness Probes](readiness.md) — Readiness configuration

## Reference Implementation

See examples:
- `examples/reference/` — Health checks in production app
- `examples/api-gateway/` — Gateway health checks

---

**Next Steps**:
1. Read [Liveness Probes](liveness.md) for liveness check patterns
2. Read [Readiness Probes](readiness.md) for readiness check patterns
3. Configure Kubernetes probes for your deployment
