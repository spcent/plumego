# Scheduler Module

> **Package Path**: `github.com/spcent/plumego/scheduler` | **Stability**: High | **Priority**: P1

## Overview

The `scheduler/` package provides a complete task scheduling system for background jobs, cron tasks, delayed execution, and retry management. It enables applications to run recurring tasks, schedule work for future execution, and handle failures with configurable retry policies.

**Key Features**:
- **Cron Jobs**: Schedule recurring tasks with cron expressions
- **Delayed Tasks**: Execute tasks after a delay
- **Immediate Tasks**: Queue tasks for immediate execution
- **Retry Policies**: Automatic retry with exponential backoff
- **Worker Pools**: Concurrent task execution
- **Graceful Shutdown**: Complete in-flight tasks on shutdown
- **Persistence**: Optional task persistence (coming soon)

## Quick Start

### Basic Usage

```go
import "github.com/spcent/plumego/scheduler"

// Create scheduler with 4 workers
sch := scheduler.New(scheduler.WithWorkers(4))

// Start scheduler
sch.Start()
defer sch.Stop()

// Add cron job (runs every hour)
sch.AddCron("cleanup", "0 * * * *", func(ctx context.Context) error {
    log.Println("Running cleanup job")
    return cleanupOldRecords()
})

// Schedule delayed task (runs after 10 seconds)
sch.Delay("send-email", 10*time.Second, func(ctx context.Context) error {
    return sendEmail("user@example.com", "Welcome!")
})

// Queue immediate task
sch.Queue("process-upload", func(ctx context.Context) error {
    return processUpload(uploadID)
})
```

### With Retry Policy

```go
// Schedule task with exponential backoff retry
sch.Delay(
    "api-call",
    time.Second,
    func(ctx context.Context) error {
        return callExternalAPI()
    },
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 5)),
)
```

### Integration with Plumego App

```go
import "github.com/spcent/plumego/core"

// Create scheduler as a runner
sch := scheduler.New(scheduler.WithWorkers(4))

app := core.New(
    core.WithRunner(sch), // Scheduler auto-starts/stops with app
)

// Add jobs
sch.AddCron("daily-report", "0 0 * * *", generateDailyReport)
```

## Core Types

### Scheduler

The main scheduler instance:

```go
type Scheduler struct {
    // ... internal fields
}

func New(opts ...Option) *Scheduler
func (s *Scheduler) Start() error
func (s *Scheduler) Stop() error
func (s *Scheduler) AddCron(name, spec string, fn TaskFunc, opts ...TaskOption) error
func (s *Scheduler) Delay(name string, delay time.Duration, fn TaskFunc, opts ...TaskOption)
func (s *Scheduler) Queue(name string, fn TaskFunc, opts ...TaskOption)
func (s *Scheduler) Remove(name string) bool
```

### TaskFunc

Function signature for tasks:

```go
type TaskFunc func(ctx context.Context) error
```

**Context Usage**:
- Check `ctx.Done()` for cancellation
- Use for timeout management
- Pass to downstream operations

### Options

Configuration options:

```go
// Scheduler options
func WithWorkers(n int) Option          // Number of concurrent workers
func WithQueueSize(n int) Option        // Task queue size

// Task options
func WithTimeout(d time.Duration) TaskOption    // Task timeout
func WithRetry(policy RetryPolicy) TaskOption   // Retry policy
```

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────┐
│                  Scheduler                       │
├─────────────────────────────────────────────────┤
│  Cron Manager     │ Task Queue  │ Worker Pool   │
│  ┌──────────┐    │ ┌────────┐  │ ┌──────────┐ │
│  │ Job 1    │────┼→│ Task 1 │──┼→│ Worker 1 │ │
│  │ Job 2    │    │ │ Task 2 │  │ │ Worker 2 │ │
│  │ Job 3    │    │ │ Task 3 │  │ │ Worker 3 │ │
│  └──────────┘    │ └────────┘  │ │ Worker 4 │ │
│                  │             │ └──────────┘ │
│                  │             │              │
│  Delayed Queue   │             │ Retry Manager│
│  ┌──────────┐    │             │ ┌──────────┐ │
│  │ Task +5s │────┼─────────────┼→│ Backoff  │ │
│  │ Task +1m │    │             │ │ Logic    │ │
│  └──────────┘    │             │ └──────────┘ │
└─────────────────────────────────────────────────┘
```

### Lifecycle

```
1. New()       → Create scheduler
2. Start()     → Start cron manager + workers
3. AddCron()   → Register recurring jobs
4. Delay()     → Schedule delayed tasks
5. Queue()     → Queue immediate tasks
6. Stop()      → Graceful shutdown (finish in-flight tasks)
```

## Common Use Cases

### 1. Cleanup Old Records

```go
sch.AddCron("cleanup-old-records", "0 2 * * *", func(ctx context.Context) error {
    // Runs daily at 2:00 AM
    cutoff := time.Now().Add(-30 * 24 * time.Hour)
    return db.DeleteRecordsBefore(cutoff)
})
```

### 2. Send Welcome Email

```go
func handleRegister(w http.ResponseWriter, r *http.Request) {
    // ... registration logic ...

    // Send welcome email after 5 seconds (avoid blocking request)
    sch.Delay("welcome-email", 5*time.Second, func(ctx context.Context) error {
        return emailService.SendWelcome(user.Email)
    })

    w.WriteHeader(http.StatusCreated)
}
```

### 3. Generate Daily Reports

```go
sch.AddCron("daily-report", "0 8 * * MON-FRI", func(ctx context.Context) error {
    // Runs weekdays at 8:00 AM
    report := generateDailyReport()
    return emailService.SendReport("admin@example.com", report)
})
```

### 4. Process Uploads

```go
func handleUpload(w http.ResponseWriter, r *http.Request) {
    // Save upload
    uploadID, err := saveUpload(r)
    if err != nil {
        http.Error(w, "Upload failed", http.StatusInternalServerError)
        return
    }

    // Process asynchronously
    sch.Queue("process-upload", func(ctx context.Context) error {
        return processUpload(uploadID)
    })

    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "upload_id": uploadID,
        "status":    "processing",
    })
}
```

### 5. Retry API Calls

```go
sch.Delay(
    "webhook-delivery",
    time.Second,
    func(ctx context.Context) error {
        return deliverWebhook(webhookURL, payload)
    },
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 5)),
    scheduler.WithTimeout(10 * time.Second),
)
```

## Configuration

### Worker Pool Size

```go
// Low traffic: 2-4 workers
sch := scheduler.New(scheduler.WithWorkers(2))

// Medium traffic: 4-8 workers
sch := scheduler.New(scheduler.WithWorkers(4))

// High traffic: 8-16 workers
sch := scheduler.New(scheduler.WithWorkers(16))
```

### Queue Size

```go
// Small queue: 100 tasks
sch := scheduler.New(
    scheduler.WithWorkers(4),
    scheduler.WithQueueSize(100),
)

// Large queue: 1000 tasks
sch := scheduler.New(
    scheduler.WithWorkers(8),
    scheduler.WithQueueSize(1000),
)
```

### Task Timeout

```go
// Short timeout: 5 seconds
sch.Queue("quick-task", taskFunc,
    scheduler.WithTimeout(5*time.Second),
)

// Long timeout: 5 minutes
sch.Queue("long-task", taskFunc,
    scheduler.WithTimeout(5*time.Minute),
)
```

## Error Handling

### Task Errors

```go
sch.Queue("my-task", func(ctx context.Context) error {
    // Task execution
    if err := doWork(); err != nil {
        // Log error
        log.Printf("Task failed: %v", err)

        // Return error (triggers retry if configured)
        return err
    }

    return nil
})
```

### Error Monitoring

```go
type MonitoredScheduler struct {
    *scheduler.Scheduler
    errorCount int
    mu         sync.Mutex
}

func (m *MonitoredScheduler) Queue(name string, fn scheduler.TaskFunc, opts ...scheduler.TaskOption) {
    // Wrap task function with error monitoring
    wrappedFn := func(ctx context.Context) error {
        err := fn(ctx)
        if err != nil {
            m.mu.Lock()
            m.errorCount++
            m.mu.Unlock()

            log.Printf("Task %s failed: %v", name, err)
            metrics.TaskErrors.WithLabelValues(name).Inc()
        }
        return err
    }

    m.Scheduler.Queue(name, wrappedFn, opts...)
}
```

## Graceful Shutdown

The scheduler implements graceful shutdown:

```go
// When Stop() is called:
// 1. Stop accepting new tasks
// 2. Wait for in-flight tasks to complete
// 3. Cancel long-running tasks after timeout

sch := scheduler.New(scheduler.WithWorkers(4))
sch.Start()

// ... use scheduler ...

// Graceful shutdown (waits for tasks)
sch.Stop()
```

## Performance Considerations

### Task Execution Time

- **Quick tasks** (<1s): Use immediate queue
- **Medium tasks** (1-30s): Use worker pool
- **Long tasks** (>30s): Consider separate worker pool or external job queue

### Concurrency

```go
// CPU-bound tasks: workers = num CPU
sch := scheduler.New(scheduler.WithWorkers(runtime.NumCPU()))

// I/O-bound tasks: workers = 2-4x num CPU
sch := scheduler.New(scheduler.WithWorkers(runtime.NumCPU() * 4))
```

### Queue Pressure

Monitor queue depth:

```go
func (s *Scheduler) QueueDepth() int {
    return len(s.taskQueue)
}

// Alert if queue is backing up
if sch.QueueDepth() > 800 {
    log.Warn("Task queue is backing up", "depth", sch.QueueDepth())
}
```

## Module Documentation

Detailed documentation for scheduler features:

- **[Cron Jobs](cron.md)** — Recurring tasks with cron expressions
- **[Delayed Tasks](delayed-tasks.md)** — Schedule tasks for future execution
- **[Retry Policies](retry-policies.md)** — Automatic retry with backoff strategies
- **[Examples](examples.md)** — Complete working examples

## Testing

### Unit Tests

```go
func TestSchedulerLifecycle(t *testing.T) {
    sch := scheduler.New(scheduler.WithWorkers(2))

    // Start scheduler
    if err := sch.Start(); err != nil {
        t.Fatalf("Start failed: %v", err)
    }

    // Queue task
    executed := false
    sch.Queue("test-task", func(ctx context.Context) error {
        executed = true
        return nil
    })

    // Wait for execution
    time.Sleep(100 * time.Millisecond)

    if !executed {
        t.Error("Task not executed")
    }

    // Stop scheduler
    if err := sch.Stop(); err != nil {
        t.Fatalf("Stop failed: %v", err)
    }
}
```

### Mock Time for Cron Tests

```go
// Use time.AfterFunc for testing delayed tasks
func TestDelayedTask(t *testing.T) {
    sch := scheduler.New(scheduler.WithWorkers(1))
    sch.Start()
    defer sch.Stop()

    executed := false
    sch.Delay("test", 10*time.Millisecond, func(ctx context.Context) error {
        executed = true
        return nil
    })

    time.Sleep(50 * time.Millisecond)

    if !executed {
        t.Error("Delayed task not executed")
    }
}
```

## Related Documentation

- [Core: Runners](../core/runners.md) — Background service patterns
- [Retry Policies](retry-policies.md) — Retry strategies
- [Examples](examples.md) — Complete examples

## Reference Implementation

See examples:
- `examples/reference/` — Scheduler with cron jobs
- `examples/scheduler/` — Complete scheduler examples
- `examples/api-gateway/` — Background task processing

---

**Next Steps**:
1. Read [Cron Jobs](cron.md) for recurring task scheduling
2. Read [Delayed Tasks](delayed-tasks.md) for future task execution
3. Read [Retry Policies](retry-policies.md) for failure handling
