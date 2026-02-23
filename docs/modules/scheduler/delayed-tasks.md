# Delayed Tasks

> **Package**: `github.com/spcent/plumego/scheduler` | **Feature**: Future task execution

## Overview

Delayed tasks allow scheduling work for execution after a specified delay. This is useful for time-sensitive operations that should not block the current request/response cycle.

**Use Cases**:
- Send welcome emails after registration
- Process uploads asynchronously
- Retry failed operations
- Rate limit enforcement
- Scheduled notifications
- Cleanup temporary resources

## Quick Start

### Basic Delayed Task

```go
import "github.com/spcent/plumego/scheduler"

sch := scheduler.New(scheduler.WithWorkers(4))
sch.Start()

// Execute after 10 seconds
sch.Delay("send-email", 10*time.Second, func(ctx context.Context) error {
    return sendEmail("user@example.com", "Welcome!")
})
```

### Multiple Delays

```go
// Immediate (1 second)
sch.Delay("quick-task", time.Second, quickTask)

// Short delay (30 seconds)
sch.Delay("medium-task", 30*time.Second, mediumTask)

// Long delay (5 minutes)
sch.Delay("long-task", 5*time.Minute, longTask)

// Very long delay (1 hour)
sch.Delay("batch-task", time.Hour, batchTask)
```

## Common Patterns

### 1. Welcome Email After Registration

```go
func handleRegister(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Create user
    user, err := createUser(req.Email, req.Password)
    if err != nil {
        http.Error(w, "Registration failed", http.StatusInternalServerError)
        return
    }

    // Schedule welcome email (don't block response)
    sch.Delay("welcome-email", 5*time.Second, func(ctx context.Context) error {
        return emailService.SendWelcome(user.Email, user.Name)
    })

    // Return immediate response
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{
        "user_id": user.ID,
        "message": "Registration successful",
    })
}
```

### 2. Async File Processing

```go
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
    // Save uploaded file
    file, _, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Save to storage
    uploadID, err := storage.Save(file)
    if err != nil {
        http.Error(w, "Upload failed", http.StatusInternalServerError)
        return
    }

    // Process asynchronously
    sch.Delay("process-upload", time.Second, func(ctx context.Context) error {
        // Extract metadata
        metadata, err := extractMetadata(uploadID)
        if err != nil {
            return err
        }

        // Generate thumbnails
        if err := generateThumbnails(uploadID); err != nil {
            return err
        }

        // Update database
        return db.UpdateUploadStatus(uploadID, "processed", metadata)
    })

    // Return accepted response
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "upload_id": uploadID,
        "status":    "processing",
    })
}
```

### 3. Rate-Limited Notifications

```go
type NotificationQueue struct {
    sch     *scheduler.Scheduler
    delays  map[string]time.Duration
    mu      sync.Mutex
}

func (nq *NotificationQueue) SendDelayed(userID, message string) {
    nq.mu.Lock()
    defer nq.mu.Unlock()

    // Get or create delay for user
    delay, exists := nq.delays[userID]
    if !exists {
        delay = time.Second // First notification: 1 second
    } else {
        delay = delay * 2 // Exponential backoff
        if delay > 5*time.Minute {
            delay = 5 * time.Minute // Cap at 5 minutes
        }
    }
    nq.delays[userID] = delay

    // Schedule notification
    nq.sch.Delay(
        fmt.Sprintf("notify-%s", userID),
        delay,
        func(ctx context.Context) error {
            return sendNotification(userID, message)
        },
    )
}
```

### 4. Cleanup Temporary Files

```go
func createTemporaryFile(data []byte) (string, error) {
    // Create temp file
    tmpFile, err := os.CreateTemp("", "upload-*.tmp")
    if err != nil {
        return "", err
    }
    defer tmpFile.Close()

    // Write data
    if _, err := tmpFile.Write(data); err != nil {
        return "", err
    }

    filepath := tmpFile.Name()

    // Schedule cleanup after 1 hour
    sch.Delay("cleanup-temp", time.Hour, func(ctx context.Context) error {
        log.Printf("Cleaning up temporary file: %s", filepath)
        return os.Remove(filepath)
    })

    return filepath, nil
}
```

### 5. Retry Failed Operations

```go
func handleWebhookDelivery(webhookURL string, payload []byte) error {
    // Attempt immediate delivery
    if err := deliverWebhook(webhookURL, payload); err != nil {
        log.Printf("Webhook delivery failed, scheduling retry: %v", err)

        // Retry after 30 seconds
        sch.Delay("webhook-retry", 30*time.Second, func(ctx context.Context) error {
            return deliverWebhook(webhookURL, payload)
        },
            scheduler.WithRetry(scheduler.RetryExponential(time.Minute, 5)),
        )

        return err
    }

    return nil
}
```

### 6. Scheduled Notifications

```go
func scheduleNotification(userID string, message string, sendAt time.Time) {
    delay := time.Until(sendAt)
    if delay < 0 {
        delay = 0 // Send immediately if in past
    }

    sch.Delay(
        fmt.Sprintf("notify-%s-%d", userID, sendAt.Unix()),
        delay,
        func(ctx context.Context) error {
            return pushNotification(userID, message)
        },
    )
}

// Usage
sendAt := time.Now().Add(24 * time.Hour) // Tomorrow
scheduleNotification("user-123", "Your subscription expires soon", sendAt)
```

## Task Options

### With Timeout

```go
// Set timeout for delayed task
sch.Delay("api-call", 10*time.Second, callExternalAPI,
    scheduler.WithTimeout(5*time.Second),
)
```

### With Retry

```go
// Retry on failure
sch.Delay("webhook", time.Second, deliverWebhook,
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 3)),
)
```

### Combined Options

```go
// Both timeout and retry
sch.Delay("critical-task", 5*time.Second, criticalTask,
    scheduler.WithTimeout(30*time.Second),
    scheduler.WithRetry(scheduler.RetryLinear(time.Second, 5)),
)
```

## Immediate Execution

Use `Queue()` for immediate execution (0 delay):

```go
// Execute as soon as worker available
sch.Queue("immediate-task", func(ctx context.Context) error {
    return processTask()
})

// Equivalent to:
sch.Delay("immediate-task", 0, processTask)
```

## Advanced Patterns

### 1. Cascading Tasks

```go
func handleComplexWorkflow(orderID string) {
    // Step 1: Process order (immediate)
    sch.Queue("process-order", func(ctx context.Context) error {
        if err := processOrder(orderID); err != nil {
            return err
        }

        // Step 2: Send confirmation (after 5 seconds)
        sch.Delay("send-confirmation", 5*time.Second, func(ctx context.Context) error {
            if err := sendOrderConfirmation(orderID); err != nil {
                return err
            }

            // Step 3: Schedule follow-up (after 7 days)
            sch.Delay("follow-up", 7*24*time.Hour, func(ctx context.Context) error {
                return sendFollowUp(orderID)
            })

            return nil
        })

        return nil
    })
}
```

### 2. Debounced Tasks

```go
type DebouncedScheduler struct {
    sch    *scheduler.Scheduler
    timers map[string]*time.Timer
    mu     sync.Mutex
}

func (ds *DebouncedScheduler) Debounce(key string, delay time.Duration, fn scheduler.TaskFunc) {
    ds.mu.Lock()
    defer ds.mu.Unlock()

    // Cancel existing timer
    if timer, exists := ds.timers[key]; exists {
        timer.Stop()
    }

    // Create new timer
    ds.timers[key] = time.AfterFunc(delay, func() {
        ds.sch.Queue(key, fn)
        ds.mu.Lock()
        delete(ds.timers, key)
        ds.mu.Unlock()
    })
}

// Usage: Multiple rapid calls will only execute once after delay
ds.Debounce("save-draft", 5*time.Second, func(ctx context.Context) error {
    return saveDraft(userID, content)
})
```

### 3. Batch Processing with Delay

```go
type BatchScheduler struct {
    sch       *scheduler.Scheduler
    items     map[string][]interface{}
    batchSize int
    mu        sync.Mutex
}

func (bs *BatchScheduler) Add(batchKey string, item interface{}) {
    bs.mu.Lock()
    defer bs.mu.Unlock()

    // Add item to batch
    bs.items[batchKey] = append(bs.items[batchKey], item)

    // Process if batch full
    if len(bs.items[batchKey]) >= bs.batchSize {
        batch := bs.items[batchKey]
        bs.items[batchKey] = nil

        bs.sch.Queue("process-batch", func(ctx context.Context) error {
            return processBatch(batch)
        })
    } else {
        // Schedule delayed processing for partial batch
        bs.sch.Delay(
            fmt.Sprintf("batch-%s", batchKey),
            30*time.Second,
            func(ctx context.Context) error {
                bs.mu.Lock()
                batch := bs.items[batchKey]
                bs.items[batchKey] = nil
                bs.mu.Unlock()

                if len(batch) > 0 {
                    return processBatch(batch)
                }
                return nil
            },
        )
    }
}
```

### 4. Scheduled Reminders

```go
type ReminderService struct {
    sch *scheduler.Scheduler
}

func (rs *ReminderService) SetReminder(userID, message string, remindAt time.Time) error {
    delay := time.Until(remindAt)
    if delay < 0 {
        return errors.New("reminder time in past")
    }

    rs.sch.Delay(
        fmt.Sprintf("reminder-%s-%d", userID, remindAt.Unix()),
        delay,
        func(ctx context.Context) error {
            return sendReminder(userID, message)
        },
    )

    return nil
}

// Usage
reminderService.SetReminder(
    "user-123",
    "Meeting in 15 minutes",
    time.Now().Add(45*time.Minute), // Remind 15 min before 1hr meeting
)
```

## Best Practices

### 1. Choose Appropriate Delays

```go
// ✅ Good: Reasonable delays
sch.Delay("email", 5*time.Second, sendEmail)        // Short delay for emails
sch.Delay("cleanup", time.Hour, cleanup)            // Long delay for cleanup
sch.Delay("notification", 10*time.Minute, notify)   // Medium delay for notifications

// ❌ Bad: Excessive delays
sch.Delay("email", 10*time.Minute, sendEmail)       // Too long for user-facing feature
```

### 2. Handle Context Cancellation

```go
sch.Delay("long-task", time.Minute, func(ctx context.Context) error {
    // Check for cancellation before long operations
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Perform work
    return doWork()
})
```

### 3. Add Retry for Critical Tasks

```go
// ✅ Good: Retry critical operations
sch.Delay("payment-webhook", time.Second, deliverPaymentWebhook,
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 5)),
)

// ⚠️ Consider: Non-critical tasks may not need retry
sch.Delay("analytics-event", time.Second, trackEvent)
```

### 4. Log Delayed Task Execution

```go
sch.Delay("task", 10*time.Second, func(ctx context.Context) error {
    log.Printf("Executing delayed task (scheduled 10s ago)")

    start := time.Now()
    err := performTask(ctx)

    log.Printf("Task completed in %v (error: %v)", time.Since(start), err)
    return err
})
```

### 5. Unique Task Names

```go
// ✅ Good: Unique names with context
sch.Delay(fmt.Sprintf("email-%s-%d", userID, time.Now().Unix()), time.Second, sendEmail)

// ❌ Bad: Duplicate names may cause conflicts
sch.Delay("email", time.Second, sendEmail)
```

## Testing

### Unit Test

```go
func TestDelayedTask(t *testing.T) {
    sch := scheduler.New(scheduler.WithWorkers(1))
    sch.Start()
    defer sch.Stop()

    executed := false
    sch.Delay("test-delay", 100*time.Millisecond, func(ctx context.Context) error {
        executed = true
        return nil
    })

    // Should not execute immediately
    if executed {
        t.Error("Task executed before delay")
    }

    // Wait for execution
    time.Sleep(200 * time.Millisecond)

    if !executed {
        t.Error("Task not executed after delay")
    }
}
```

### Test with Mock Time

```go
func TestDelayedTaskWithMockTime(t *testing.T) {
    // Use channels for synchronization
    done := make(chan bool, 1)

    sch := scheduler.New(scheduler.WithWorkers(1))
    sch.Start()
    defer sch.Stop()

    sch.Delay("test", 50*time.Millisecond, func(ctx context.Context) error {
        done <- true
        return nil
    })

    select {
    case <-done:
        // Task executed
    case <-time.After(200 * time.Millisecond):
        t.Error("Task not executed within timeout")
    }
}
```

## Performance Considerations

### Memory Usage

Delayed tasks consume memory until execution. For many delayed tasks:

```go
// ✅ Good: Reasonable number of delayed tasks
for i := 0; i < 100; i++ {
    sch.Delay(fmt.Sprintf("task-%d", i), time.Minute, processTask)
}

// ⚠️ Consider: Very large number may consume significant memory
for i := 0; i < 10000; i++ {
    sch.Delay(fmt.Sprintf("task-%d", i), time.Hour, processTask)
}
// Consider using external queue (Redis, RabbitMQ) for scale
```

### Scheduling Overhead

Very short delays add minimal overhead:

```go
// Fine for occasional use
sch.Delay("task", 100*time.Millisecond, quickTask)

// For high-frequency, consider direct worker pool
```

## Related Documentation

- [Scheduler Overview](README.md) — Scheduler module overview
- [Cron Jobs](cron.md) — Recurring task scheduling
- [Retry Policies](retry-policies.md) — Retry strategies
- [Examples](examples.md) — Complete examples

## Reference Implementation

See examples:
- `examples/scheduler/` — Delayed task examples
- `examples/reference/` — Production delayed tasks
