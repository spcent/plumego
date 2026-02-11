# Liveness Probes

> **Package**: `github.com/spcent/plumego/health` | **Purpose**: Detect if application is running

## Overview

Liveness probes determine whether an application is alive and functioning. If a liveness probe fails, Kubernetes (or other orchestrators) will restart the container to attempt recovery from a deadlocked or hung state.

**When to use**: Detect application hangs, deadlocks, or unrecoverable errors

**Orchestrator action**: Restart container on failure

## Quick Start

### Basic Liveness Endpoint

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/health"
)

app := core.New()

// Simple liveness: Just checks if process is alive
app.Get("/health/live", health.LivenessHandler())

app.Boot()
```

**Response**:
```json
{
  "status": "healthy"
}
```

### Custom Liveness Checks

```go
checker := health.NewChecker()

// Check if app is responsive
checker.AddLiveness("app", func() health.HealthStatus {
    // Application is alive if we can execute this function
    return health.StatusHealthy()
})

// Check for deadlocks
checker.AddLiveness("goroutines", func() health.HealthStatus {
    numGoroutines := runtime.NumGoroutine()
    if numGoroutines > 10000 {
        return health.StatusUnhealthy("too many goroutines: possible leak")
    }
    return health.StatusHealthy()
})

app.Get("/health/live", checker.LivenessHandler())
```

## Common Liveness Checks

### 1. Basic Alive Check

The simplest check - if the endpoint responds, the app is alive:

```go
app.Get("/health/live", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

### 2. Goroutine Leak Detection

```go
checker.AddLiveness("goroutines", func() health.HealthStatus {
    numGoroutines := runtime.NumGoroutine()

    // Warning threshold
    if numGoroutines > 5000 {
        return health.StatusDegraded(fmt.Sprintf("high goroutine count: %d", numGoroutines))
    }

    // Critical threshold
    if numGoroutines > 10000 {
        return health.StatusUnhealthy(fmt.Sprintf("goroutine leak detected: %d", numGoroutines))
    }

    return health.StatusHealthy()
})
```

### 3. Memory Exhaustion

```go
checker.AddLiveness("memory", func() health.HealthStatus {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    allocMB := m.Alloc / 1024 / 1024
    sysMB := m.Sys / 1024 / 1024

    // Check if memory usage is excessive
    if allocMB > 2048 { // 2 GB
        return health.StatusUnhealthy(fmt.Sprintf("excessive memory: %d MB allocated", allocMB))
    }

    if allocMB > 1024 { // 1 GB
        return health.StatusDegraded(fmt.Sprintf("high memory: %d MB allocated", allocMB))
    }

    return health.StatusHealthy()
})
```

### 4. Disk Space

```go
checker.AddLiveness("disk", func() health.HealthStatus {
    var stat syscall.Statfs_t
    if err := syscall.Statfs("/", &stat); err != nil {
        return health.StatusUnhealthy("cannot check disk space")
    }

    available := stat.Bavail * uint64(stat.Bsize)
    total := stat.Blocks * uint64(stat.Bsize)
    percentUsed := float64(total-available) / float64(total) * 100

    if percentUsed > 98 {
        return health.StatusUnhealthy(fmt.Sprintf("disk critical: %.1f%% full", percentUsed))
    }

    return health.StatusHealthy()
})
```

### 5. Deadlock Detection

```go
type DeadlockDetector struct {
    lastCheck time.Time
    mu        sync.Mutex
}

func (dd *DeadlockDetector) Heartbeat() {
    dd.mu.Lock()
    dd.lastCheck = time.Now()
    dd.mu.Unlock()
}

func (dd *DeadlockDetector) Check() health.HealthStatus {
    dd.mu.Lock()
    lastCheck := dd.lastCheck
    dd.mu.Unlock()

    if time.Since(lastCheck) > 30*time.Second {
        return health.StatusUnhealthy("no heartbeat in 30s: possible deadlock")
    }

    return health.StatusHealthy()
}

// Register check
detector := &DeadlockDetector{lastCheck: time.Now()}
checker.AddLiveness("deadlock", detector.Check)

// Call Heartbeat() periodically in main goroutine
go func() {
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        detector.Heartbeat()
    }
}()
```

## Kubernetes Configuration

### Basic Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 30    # Wait 30s before first check
  periodSeconds: 10           # Check every 10s
  timeoutSeconds: 5           # 5s timeout per check
  failureThreshold: 3         # Restart after 3 failures
```

### Aggressive Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 1
  failureThreshold: 2
```

### Conservative Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 60
  periodSeconds: 30
  timeoutSeconds: 10
  failureThreshold: 5
```

## Best Practices

### 1. Keep Liveness Checks Simple

```go
// ✅ Good: Simple, fast check
checker.AddLiveness("app", func() health.HealthStatus {
    return health.StatusHealthy()
})

// ❌ Bad: Checking external dependencies
checker.AddLiveness("database", func() health.HealthStatus {
    if err := db.Ping(); err != nil {
        return health.StatusUnhealthy("db down")
    }
    return health.StatusHealthy()
})
// Use readiness checks for dependencies!
```

### 2. Avoid False Positives

```go
// ✅ Good: High thresholds to avoid false restarts
checker.AddLiveness("goroutines", func() health.HealthStatus {
    if runtime.NumGoroutine() > 10000 {
        return health.StatusUnhealthy("goroutine leak")
    }
    return health.StatusHealthy()
})

// ❌ Bad: Too sensitive
checker.AddLiveness("goroutines", func() health.HealthStatus {
    if runtime.NumGoroutine() > 100 { // Too low!
        return health.StatusUnhealthy("goroutine leak")
    }
    return health.StatusHealthy()
})
```

### 3. Set Appropriate Initial Delay

```yaml
# ✅ Good: Allow time for app to start
livenessProbe:
  initialDelaySeconds: 30
  # ... other settings

# ❌ Bad: Check immediately (app may not be ready)
livenessProbe:
  initialDelaySeconds: 0
  # ... other settings
```

### 4. Don't Make Liveness Too Aggressive

```yaml
# ✅ Good: Reasonable failure threshold
livenessProbe:
  failureThreshold: 3        # 3 failures = 30s before restart
  periodSeconds: 10

# ❌ Bad: Restarts too quickly
livenessProbe:
  failureThreshold: 1        # Single failure causes restart
  periodSeconds: 5
```

## Common Pitfalls

### Pitfall 1: Checking Dependencies

**Problem**: Liveness checks database/cache, causing restarts when external service fails

**Solution**: Move dependency checks to readiness

```go
// ❌ DON'T: Check database in liveness
checker.AddLiveness("database", dbCheck)

// ✅ DO: Check database in readiness
checker.AddReadiness("database", dbCheck)
```

### Pitfall 2: Slow Checks

**Problem**: Liveness check takes too long, causing timeout failures

**Solution**: Use fast checks with short timeouts

```go
// ✅ Fast check
checker.AddLiveness("app", func() health.HealthStatus {
    return health.StatusHealthy()
})
```

### Pitfall 3: Restart Loops

**Problem**: App keeps restarting due to failing liveness check

**Indicators**:
- Pod restarts continuously
- App never becomes ready
- Logs show incomplete initialization

**Solution**: Increase `initialDelaySeconds` or fix underlying issue

```yaml
livenessProbe:
  initialDelaySeconds: 60    # Give more time to start
  failureThreshold: 5        # More tolerance
```

## Testing

### Unit Test

```go
func TestLivenessEndpoint(t *testing.T) {
    checker := health.NewChecker()
    checker.AddLiveness("test", func() health.HealthStatus {
        return health.StatusHealthy()
    })

    req := httptest.NewRequest("GET", "/health/live", nil)
    rec := httptest.NewRecorder()

    checker.LivenessHandler()(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", rec.Code)
    }
}
```

### Integration Test

```go
func TestLivenessProbe(t *testing.T) {
    // Start app
    app := startTestApp(t)
    defer app.Stop()

    // Check liveness endpoint
    resp, err := http.Get("http://localhost:8080/health/live")
    if err != nil {
        t.Fatalf("Liveness check failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}
```

## Monitoring

### Track Liveness Failures

```go
var livenessFailures = prometheus.NewCounter(prometheus.CounterOpts{
    Name: "liveness_check_failures_total",
    Help: "Total number of liveness check failures",
})

checker.AddLiveness("monitored", func() health.HealthStatus {
    status := performCheck()
    if status.Status == "unhealthy" {
        livenessFailures.Inc()
    }
    return status
})
```

## Related Documentation

- [Health Overview](README.md) — Health module overview
- [Readiness Probes](readiness.md) — Readiness configuration

## Reference

- [Kubernetes Liveness Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
