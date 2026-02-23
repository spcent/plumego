# Readiness Probes

> **Package**: `github.com/spcent/plumego/health` | **Purpose**: Detect if application can handle traffic

## Overview

Readiness probes determine whether an application is ready to accept traffic. If a readiness probe fails, Kubernetes (or other orchestrators) will remove the pod from service endpoints, preventing traffic from being routed to it.

**When to use**: Check if dependencies are available and app is fully initialized

**Orchestrator action**: Remove from load balancer on failure (no restart)

## Quick Start

### Basic Readiness Endpoint

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/health"
)

checker := health.NewChecker()

// Check database connectivity
checker.AddReadiness("database", func() health.HealthStatus {
    if err := db.Ping(); err != nil {
        return health.StatusUnhealthy("database unreachable")
    }
    return health.StatusHealthy()
})

// Check cache connectivity
checker.AddReadiness("cache", func() health.HealthStatus {
    if err := cache.Ping(); err != nil {
        return health.StatusDegraded("cache unavailable")
    }
    return health.StatusHealthy()
})

app := core.New()
app.Get("/health/ready", checker.ReadinessHandler())
app.Boot()
```

## Common Readiness Checks

### 1. Database Connectivity

```go
checker.AddReadiness("database", func() health.HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Check connectivity
    if err := db.PingContext(ctx); err != nil {
        return health.StatusUnhealthy(fmt.Sprintf("db ping failed: %v", err))
    }

    // Check connection pool
    stats := db.Stats()
    if stats.OpenConnections >= stats.MaxOpenConnections {
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

    // Ping cache
    if err := cache.Ping(ctx); err != nil {
        // Cache down: degraded (can still serve without cache)
        return health.StatusDegraded("cache unreachable")
    }

    // Check memory usage
    info := cache.Info()
    memUsage := float64(info.UsedMemory) / float64(info.MaxMemory) * 100

    if memUsage > 95 {
        return health.StatusDegraded(fmt.Sprintf("cache memory %.1f%% full", memUsage))
    }

    return health.StatusHealthy()
})
```

### 3. External Service Dependencies

```go
checker.AddReadiness("payment-service", func() health.HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(
        ctx,
        "GET",
        "https://api.payment.com/health",
        nil,
    )

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return health.StatusUnhealthy(fmt.Sprintf("payment service unreachable: %v", err))
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return health.StatusUnhealthy(fmt.Sprintf("payment service unhealthy: %d", resp.StatusCode))
    }

    return health.StatusHealthy()
})
```

### 4. Message Queue Connectivity

```go
checker.AddReadiness("message-queue", func() health.HealthStatus {
    if !messageQueue.IsConnected() {
        return health.StatusUnhealthy("message queue disconnected")
    }

    // Check queue depth
    depth := messageQueue.Depth()
    if depth > 10000 {
        return health.StatusDegraded(fmt.Sprintf("high queue depth: %d", depth))
    }

    return health.StatusHealthy()
})
```

### 5. Application Initialization

```go
type App struct {
    ready atomic.Bool
}

func (a *App) Initialize() {
    // Load configuration
    loadConfig()

    // Connect to databases
    connectDB()

    // Warm caches
    warmCaches()

    // Mark as ready
    a.ready.Store(true)
}

// Readiness check
checker.AddReadiness("initialization", func() health.HealthStatus {
    if !app.ready.Load() {
        return health.StatusUnhealthy("initialization in progress")
    }
    return health.StatusHealthy()
})
```

### 6. License/Configuration Validation

```go
checker.AddReadiness("license", func() health.HealthStatus {
    license, err := loadLicense()
    if err != nil {
        return health.StatusUnhealthy("license not found")
    }

    if license.IsExpired() {
        return health.StatusUnhealthy("license expired")
    }

    if license.ExpiresIn() < 7*24*time.Hour {
        return health.StatusDegraded("license expiring soon")
    }

    return health.StatusHealthy()
})
```

## Kubernetes Configuration

### Basic Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5     # Wait 5s before first check
  periodSeconds: 5            # Check every 5s
  timeoutSeconds: 3           # 3s timeout per check
  failureThreshold: 3         # Remove from LB after 3 failures
  successThreshold: 1         # Add to LB after 1 success
```

### Fast Readiness Probe

For quick recovery after transient failures:

```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 2
  periodSeconds: 2
  timeoutSeconds: 1
  failureThreshold: 2
  successThreshold: 1
```

### Conservative Readiness Probe

For services with slow startup or flaky dependencies:

```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 5
  successThreshold: 2
```

## Graceful Degradation

### Using Degraded Status

Allow partial functionality when non-critical dependencies fail:

```go
// Critical: Database (unhealthy if down)
checker.AddReadiness("database", func() health.HealthStatus {
    if err := db.Ping(); err != nil {
        return health.StatusUnhealthy("database unreachable")
    }
    return health.StatusHealthy()
})

// Non-critical: Cache (degraded if down)
checker.AddReadiness("cache", func() health.HealthStatus {
    if err := cache.Ping(); err != nil {
        // Still ready, but degraded
        return health.StatusDegraded("cache unavailable")
    }
    return health.StatusHealthy()
})

// Non-critical: Analytics (degraded if down)
checker.AddReadiness("analytics", func() health.HealthStatus {
    if err := analytics.Ping(); err != nil {
        return health.StatusDegraded("analytics unavailable")
    }
    return health.StatusHealthy()
})
```

**Behavior**:
- Database down → `503 Service Unavailable` (removed from LB)
- Cache down → `200 OK` with degraded status (stays in LB)

## Best Practices

### 1. Check All Critical Dependencies

```go
// ✅ Good: Check all dependencies that affect functionality
checker.AddReadiness("database", dbCheck)
checker.AddReadiness("cache", cacheCheck)
checker.AddReadiness("auth-service", authCheck)

// ❌ Bad: Missing critical dependency
// (not checking database)
```

### 2. Use Fast Checks with Timeouts

```go
// ✅ Good: Fast check with timeout
checker.AddReadiness("database", func() health.HealthStatus {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return health.StatusUnhealthy("db unreachable")
    }
    return health.StatusHealthy()
})

// ❌ Bad: No timeout (may hang)
checker.AddReadiness("database", func() health.HealthStatus {
    if err := db.Ping(); err != nil {
        return health.StatusUnhealthy("db unreachable")
    }
    return health.StatusHealthy()
})
```

### 3. Distinguish Critical vs Non-Critical

```go
// Critical: unhealthy if down
checker.AddReadiness("database", func() health.HealthStatus {
    if err := db.Ping(); err != nil {
        return health.StatusUnhealthy("db down")
    }
    return health.StatusHealthy()
})

// Non-critical: degraded if down
checker.AddReadiness("cache", func() health.HealthStatus {
    if err := cache.Ping(); err != nil {
        return health.StatusDegraded("cache down")
    }
    return health.StatusHealthy()
})
```

### 4. Don't Check Non-Dependencies

```go
// ❌ Bad: Checking analytics (not needed for core functionality)
checker.AddReadiness("analytics", analyticsCheck)

// ✅ Good: Only check what affects user-facing features
checker.AddReadiness("database", dbCheck)
checker.AddReadiness("payment-api", paymentCheck)
```

## Common Patterns

### 1. Circuit Breaker Integration

```go
type CircuitBreaker struct {
    failures    int
    lastFailure time.Time
    threshold   int
    mu          sync.Mutex
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    // Reset after timeout
    if time.Since(cb.lastFailure) > time.Minute {
        cb.failures = 0
    }

    // Check if circuit open
    if cb.failures >= cb.threshold {
        return errors.New("circuit open")
    }

    err := fn()
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
    }

    return err
}

// Use in readiness check
var apiBreaker = &CircuitBreaker{threshold: 5}

checker.AddReadiness("external-api", func() health.HealthStatus {
    err := apiBreaker.Call(func() error {
        return pingExternalAPI()
    })

    if err != nil {
        return health.StatusUnhealthy("external API unavailable")
    }

    return health.StatusHealthy()
})
```

### 2. Startup Sequence Tracking

```go
type StartupPhases struct {
    configLoaded   atomic.Bool
    dbConnected    atomic.Bool
    cacheConnected atomic.Bool
    routesReady    atomic.Bool
}

func (sp *StartupPhases) AllReady() bool {
    return sp.configLoaded.Load() &&
        sp.dbConnected.Load() &&
        sp.cacheConnected.Load() &&
        sp.routesReady.Load()
}

// Readiness check
checker.AddReadiness("startup", func() health.HealthStatus {
    if !phases.AllReady() {
        var pending []string
        if !phases.configLoaded.Load() {
            pending = append(pending, "config")
        }
        if !phases.dbConnected.Load() {
            pending = append(pending, "database")
        }
        if !phases.cacheConnected.Load() {
            pending = append(pending, "cache")
        }
        if !phases.routesReady.Load() {
            pending = append(pending, "routes")
        }

        return health.StatusUnhealthy(fmt.Sprintf("pending: %v", pending))
    }
    return health.StatusHealthy()
})
```

### 3. Dependency Health Aggregation

```go
type DependencyChecker struct {
    checks map[string]HealthCheck
    mu     sync.RWMutex
}

func (dc *DependencyChecker) Add(name string, check HealthCheck) {
    dc.mu.Lock()
    dc.checks[name] = check
    dc.mu.Unlock()
}

func (dc *DependencyChecker) CheckAll() health.HealthStatus {
    dc.mu.RLock()
    defer dc.mu.RUnlock()

    allHealthy := true
    var unhealthy []string

    for name, check := range dc.checks {
        status := check()
        if status.Status != "healthy" {
            allHealthy = false
            unhealthy = append(unhealthy, name)
        }
    }

    if !allHealthy {
        return health.StatusUnhealthy(
            fmt.Sprintf("dependencies unhealthy: %v", unhealthy),
        )
    }

    return health.StatusHealthy()
}
```

## Testing

### Unit Test

```go
func TestReadinessCheck(t *testing.T) {
    checker := health.NewChecker()

    // Add mock check
    checker.AddReadiness("test", func() health.HealthStatus {
        return health.StatusHealthy()
    })

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

### Integration Test

```go
func TestReadinessWithDatabase(t *testing.T) {
    // Start test database
    db := startTestDB(t)
    defer db.Close()

    // Create checker with database check
    checker := health.NewChecker()
    checker.AddReadiness("database", func() health.HealthStatus {
        if err := db.Ping(); err != nil {
            return health.StatusUnhealthy("db unreachable")
        }
        return health.StatusHealthy()
    })

    // Test healthy state
    req := httptest.NewRequest("GET", "/health/ready", nil)
    rec := httptest.NewRecorder()
    checker.ReadinessHandler()(rec, req)

    if rec.Code != http.StatusOK {
        t.Error("Expected healthy when DB is up")
    }

    // Stop database
    db.Close()

    // Test unhealthy state
    req = httptest.NewRequest("GET", "/health/ready", nil)
    rec = httptest.NewRecorder()
    checker.ReadinessHandler()(rec, req)

    if rec.Code != http.StatusServiceUnavailable {
        t.Error("Expected unhealthy when DB is down")
    }
}
```

## Monitoring

### Track Readiness State

```go
var (
    readinessStatus = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "readiness_status",
        Help: "Current readiness status (1=healthy, 0=unhealthy)",
    })
    readinessFailures = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "readiness_check_failures_total",
            Help: "Total readiness check failures by component",
        },
        []string{"component"},
    )
)

// Wrap checks with metrics
func wrapWithMetrics(name string, check HealthCheck) HealthCheck {
    return func() health.HealthStatus {
        status := check()

        if status.Status == "unhealthy" {
            readinessFailures.WithLabelValues(name).Inc()
            readinessStatus.Set(0)
        } else {
            readinessStatus.Set(1)
        }

        return status
    }
}
```

## Related Documentation

- [Health Overview](README.md) — Health module overview
- [Liveness Probes](liveness.md) — Liveness configuration

## Reference

- [Kubernetes Readiness Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
