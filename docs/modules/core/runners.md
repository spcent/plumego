# Runner System

> **Package**: `github.com/spcent/plumego/core`

Runners are lightweight background services that run alongside your HTTP server. Unlike Components, Runners are simple services with only Start/Stop methods, making them ideal for background tasks, workers, and periodic jobs.

---

## Table of Contents

- [Overview](#overview)
- [Runner Interface](#runner-interface)
- [Creating Runners](#creating-runners)
- [Registration](#registration)
- [Lifecycle](#lifecycle)
- [Patterns](#patterns)
- [Best Practices](#best-practices)
- [Examples](#examples)

---

## Overview

### What is a Runner?

A Runner is a lightweight background service that:
- Runs concurrently with the HTTP server
- Has simple Start/Stop lifecycle
- Does not register routes or middleware
- Does not have dependencies (simpler than Components)

### When to Use Runners

Use Runners for:
- Background workers
- Periodic cleanup tasks
- Message queue consumers
- Cache warmer services
- Health monitors
- Metrics collectors

Use Components when you need:
- HTTP route registration
- Middleware registration
- Dependency management
- Health check reporting

---

## Runner Interface

```go
type Runner interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

### Method Responsibilities

#### Start

- Called during application boot
- Should start background goroutines
- Should **not** block (launch goroutines instead)
- Context is for initialization, not lifetime

```go
func (r *MyRunner) Start(ctx context.Context) error {
    r.done = make(chan struct{})

    // Launch background goroutine
    go r.worker()

    return nil
}
```

#### Stop

- Called during application shutdown
- Should stop all goroutines
- Should respect context timeout
- Should clean up resources

```go
func (r *MyRunner) Stop(ctx context.Context) error {
    close(r.done)

    // Wait for worker to finish or timeout
    select {
    case <-r.workerDone:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

---

## Creating Runners

### Basic Runner

```go
package myrunner

import (
    "context"
    "log"
    "time"
)

type Worker struct {
    interval time.Duration
    done     chan struct{}
}

func New(interval time.Duration) *Worker {
    return &Worker{
        interval: interval,
        done:     make(chan struct{}),
    }
}

func (w *Worker) Start(ctx context.Context) error {
    log.Println("Worker: Starting...")

    go w.worker()

    return nil
}

func (w *Worker) Stop(ctx context.Context) error {
    log.Println("Worker: Stopping...")
    close(w.done)
    return nil
}

func (w *Worker) worker() {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            w.doWork()
        case <-w.done:
            return
        }
    }
}

func (w *Worker) doWork() {
    log.Println("Worker: Doing work...")
}
```

### Runner with Graceful Shutdown

```go
type GracefulWorker struct {
    interval   time.Duration
    done       chan struct{}
    workerDone chan struct{}
}

func (w *GracefulWorker) Start(ctx context.Context) error {
    w.done = make(chan struct{})
    w.workerDone = make(chan struct{})

    go w.worker()

    return nil
}

func (w *GracefulWorker) Stop(ctx context.Context) error {
    close(w.done)

    // Wait for worker to finish gracefully
    select {
    case <-w.workerDone:
        log.Println("Worker stopped gracefully")
        return nil
    case <-ctx.Done():
        log.Println("Worker shutdown timeout")
        return ctx.Err()
    }
}

func (w *GracefulWorker) worker() {
    defer close(w.workerDone)

    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            w.doWork()
        case <-w.done:
            log.Println("Worker received stop signal")
            return
        }
    }
}
```

### Runner with Error Handling

```go
type ResilientWorker struct {
    interval time.Duration
    done     chan struct{}
    errors   chan error
}

func (w *ResilientWorker) Start(ctx context.Context) error {
    w.done = make(chan struct{})
    w.errors = make(chan error, 10)

    go w.worker()
    go w.errorHandler()

    return nil
}

func (w *ResilientWorker) Stop(ctx context.Context) error {
    close(w.done)
    return nil
}

func (w *ResilientWorker) worker() {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := w.doWork(); err != nil {
                select {
                case w.errors <- err:
                default:
                    log.Printf("Error queue full, dropping error: %v", err)
                }
            }
        case <-w.done:
            return
        }
    }
}

func (w *ResilientWorker) errorHandler() {
    for {
        select {
        case err := <-w.errors:
            log.Printf("Worker error: %v", err)
            // Could send to error tracking service
        case <-w.done:
            return
        }
    }
}

func (w *ResilientWorker) doWork() error {
    // Work that might fail
    return nil
}
```

---

## Registration

### Basic Registration

```go
app := core.New(
    core.WithRunner(&MyWorker{}),
)
```

### Multiple Runners

```go
app := core.New(
    core.WithRunner(&CleanupWorker{}),
    core.WithRunner(&MetricsCollector{}),
    core.WithRunner(&CacheWarmer{}),
)
```

### Registration with Configuration

```go
cleanupWorker := &CleanupWorker{
    Interval: 1 * time.Hour,
    RetentionDays: 30,
}

metricsCollector := &MetricsCollector{
    Interval: 10 * time.Second,
}

app := core.New(
    core.WithRunner(cleanupWorker),
    core.WithRunner(metricsCollector),
)
```

---

## Lifecycle

### Startup Sequence

```
Boot():
  1. Start Components (in dependency order)
  2. Start Runners (concurrently)  ← All runners start at the same time
  3. Start HTTP Server
```

**Note**: Unlike Components, Runners:
- Start **concurrently** (no ordering)
- Start **after** all Components
- Have no dependency resolution

### Shutdown Sequence

```
Shutdown():
  1. Stop HTTP Server
  2. Stop Components (reverse order)
  3. Stop Runners (concurrently)  ← All runners stop at the same time
```

### Lifecycle Example

```go
type LoggingWorker struct {
    name string
    done chan struct{}
}

func (w *LoggingWorker) Start(ctx context.Context) error {
    log.Printf("[%s] Starting...", w.name)
    w.done = make(chan struct{})

    go func() {
        <-w.done
        log.Printf("[%s] Worker goroutine stopped", w.name)
    }()

    log.Printf("[%s] Started successfully", w.name)
    return nil
}

func (w *LoggingWorker) Stop(ctx context.Context) error {
    log.Printf("[%s] Stopping...", w.name)
    close(w.done)
    time.Sleep(100 * time.Millisecond) // Simulate cleanup
    log.Printf("[%s] Stopped successfully", w.name)
    return nil
}

// Output during boot:
// [Worker1] Starting...
// [Worker2] Starting...
// [Worker1] Started successfully
// [Worker2] Started successfully

// Output during shutdown:
// [Worker1] Stopping...
// [Worker2] Stopping...
// [Worker1] Worker goroutine stopped
// [Worker2] Worker goroutine stopped
// [Worker1] Stopped successfully
// [Worker2] Stopped successfully
```

---

## Patterns

### 1. Periodic Task Runner

```go
type PeriodicTask struct {
    interval time.Duration
    task     func() error
    done     chan struct{}
}

func NewPeriodicTask(interval time.Duration, task func() error) *PeriodicTask {
    return &PeriodicTask{
        interval: interval,
        task:     task,
        done:     make(chan struct{}),
    }
}

func (p *PeriodicTask) Start(ctx context.Context) error {
    go p.run()
    return nil
}

func (p *PeriodicTask) Stop(ctx context.Context) error {
    close(p.done)
    return nil
}

func (p *PeriodicTask) run() {
    ticker := time.NewTicker(p.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := p.task(); err != nil {
                log.Printf("Task error: %v", err)
            }
        case <-p.done:
            return
        }
    }
}

// Usage
cleanupTask := NewPeriodicTask(1*time.Hour, func() error {
    log.Println("Running cleanup...")
    return nil
})

app := core.New(
    core.WithRunner(cleanupTask),
)
```

### 2. Message Queue Consumer

```go
type QueueConsumer struct {
    queueURL string
    handler  func(message []byte) error
    done     chan struct{}
}

func (q *QueueConsumer) Start(ctx context.Context) error {
    q.done = make(chan struct{})
    go q.consume()
    return nil
}

func (q *QueueConsumer) Stop(ctx context.Context) error {
    close(q.done)
    return nil
}

func (q *QueueConsumer) consume() {
    for {
        select {
        case <-q.done:
            return
        default:
            message := q.receiveMessage()
            if message != nil {
                if err := q.handler(message); err != nil {
                    log.Printf("Handler error: %v", err)
                }
            }
        }
    }
}

func (q *QueueConsumer) receiveMessage() []byte {
    // Poll queue for messages
    return nil
}
```

### 3. Health Monitor

```go
type HealthMonitor struct {
    services []Checkable
    interval time.Duration
    done     chan struct{}
}

type Checkable interface {
    Check() error
}

func (h *HealthMonitor) Start(ctx context.Context) error {
    h.done = make(chan struct{})
    go h.monitor()
    return nil
}

func (h *HealthMonitor) Stop(ctx context.Context) error {
    close(h.done)
    return nil
}

func (h *HealthMonitor) monitor() {
    ticker := time.NewTicker(h.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            for _, service := range h.services {
                if err := service.Check(); err != nil {
                    log.Printf("Health check failed: %v", err)
                    // Could send alert
                }
            }
        case <-h.done:
            return
        }
    }
}
```

### 4. Cache Warmer

```go
type CacheWarmer struct {
    cache    Cache
    loader   DataLoader
    interval time.Duration
    done     chan struct{}
}

type Cache interface {
    Set(key string, value interface{}) error
}

type DataLoader interface {
    Load() (map[string]interface{}, error)
}

func (c *CacheWarmer) Start(ctx context.Context) error {
    c.done = make(chan struct{})

    // Warm cache immediately
    if err := c.warm(); err != nil {
        log.Printf("Initial cache warm failed: %v", err)
    }

    // Continue warming periodically
    go c.run()

    return nil
}

func (c *CacheWarmer) Stop(ctx context.Context) error {
    close(c.done)
    return nil
}

func (c *CacheWarmer) run() {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := c.warm(); err != nil {
                log.Printf("Cache warm failed: %v", err)
            }
        case <-c.done:
            return
        }
    }
}

func (c *CacheWarmer) warm() error {
    data, err := c.loader.Load()
    if err != nil {
        return err
    }

    for key, value := range data {
        if err := c.cache.Set(key, value); err != nil {
            log.Printf("Failed to cache %s: %v", key, err)
        }
    }

    return nil
}
```

### 5. Metrics Collector

```go
type MetricsCollector struct {
    interval time.Duration
    metrics  MetricsStore
    done     chan struct{}
}

type MetricsStore interface {
    Record(name string, value float64)
}

func (m *MetricsCollector) Start(ctx context.Context) error {
    m.done = make(chan struct{})
    go m.collect()
    return nil
}

func (m *MetricsCollector) Stop(ctx context.Context) error {
    close(m.done)
    return nil
}

func (m *MetricsCollector) collect() {
    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Collect system metrics
            m.metrics.Record("goroutines", float64(runtime.NumGoroutine()))
            m.metrics.Record("memory_alloc", float64(m.memStats().Alloc))
            m.metrics.Record("memory_sys", float64(m.memStats().Sys))
        case <-m.done:
            return
        }
    }
}

func (m *MetricsCollector) memStats() runtime.MemStats {
    var ms runtime.MemStats
    runtime.ReadMemStats(&ms)
    return ms
}
```

---

## Best Practices

### ✅ Do

1. **Launch Goroutines in Start()**
   ```go
   func (r *Runner) Start(ctx context.Context) error {
       go r.worker() // ✅ Don't block
       return nil
   }
   ```

2. **Use Channels for Shutdown**
   ```go
   type Runner struct {
       done chan struct{}
   }

   func (r *Runner) Start(ctx context.Context) error {
       r.done = make(chan struct{})
       go r.worker()
       return nil
   }

   func (r *Runner) Stop(ctx context.Context) error {
       close(r.done)
       return nil
   }
   ```

3. **Respect Context Timeout**
   ```go
   func (r *Runner) Stop(ctx context.Context) error {
       close(r.done)

       select {
       case <-r.workerDone:
           return nil
       case <-ctx.Done():
           return ctx.Err()
       }
   }
   ```

4. **Handle Errors Gracefully**
   ```go
   func (r *Runner) worker() {
       for {
           if err := r.doWork(); err != nil {
               log.Printf("Work error: %v", err)
               // Continue or break based on error type
           }
       }
   }
   ```

### ❌ Don't

1. **Don't Block in Start()**
   ```go
   // ❌ Blocks forever
   func (r *Runner) Start(ctx context.Context) error {
       for {
           r.doWork()
       }
       return nil
   }

   // ✅ Launch goroutine
   func (r *Runner) Start(ctx context.Context) error {
       go r.worker()
       return nil
   }
   ```

2. **Don't Leak Goroutines**
   ```go
   // ❌ No way to stop
   func (r *Runner) Start(ctx context.Context) error {
       go func() {
           for {
               r.doWork()
           }
       }()
       return nil
   }

   // ✅ With done channel
   func (r *Runner) worker() {
       for {
           select {
           case <-r.done:
               return
           default:
               r.doWork()
           }
       }
   }
   ```

3. **Don't Panic**
   ```go
   // ❌ Panics
   func (r *Runner) worker() {
       panic("something went wrong")
   }

   // ✅ Handle errors
   func (r *Runner) worker() {
       defer func() {
           if r := recover(); r != nil {
               log.Printf("Panic recovered: %v", r)
           }
       }()
       // work...
   }
   ```

---

## Examples

See [Runner Patterns](#patterns) section for complete examples including:
- Periodic Task Runner
- Message Queue Consumer
- Health Monitor
- Cache Warmer
- Metrics Collector

---

## Comparison: Runner vs Component

| Feature | Runner | Component |
|---------|--------|-----------|
| **Complexity** | Simple | Complex |
| **Routes** | ❌ No | ✅ Yes |
| **Middleware** | ❌ No | ✅ Yes |
| **Dependencies** | ❌ No | ✅ Yes |
| **Health Checks** | ❌ No | ✅ Yes |
| **Start Order** | Concurrent | Topological |
| **Use Case** | Background tasks | HTTP features |

---

## Next Steps

- **[Components](components.md)** - For HTTP-aware services
- **[Lifecycle](lifecycle.md)** - Understanding application lifecycle
- **[Scheduler](../scheduler/)** - For cron and delayed tasks

---

**Related**:
- [Component System](components.md)
- [Scheduler Module](../scheduler/)
- [Lifecycle Management](lifecycle.md)
