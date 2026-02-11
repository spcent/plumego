# Retry Policies

> **Package**: `github.com/spcent/plumego/scheduler` | **Feature**: Automatic retry with backoff

## Overview

Retry policies enable automatic retry of failed tasks with configurable backoff strategies. This improves system resilience by handling transient failures without manual intervention.

**Use Cases**:
- External API calls (rate limits, timeouts)
- Database operations (connection failures)
- Network requests (temporary outages)
- File operations (lock conflicts)
- Message delivery (webhook failures)

## Quick Start

### Basic Retry

```go
import "github.com/spcent/plumego/scheduler"

// Retry with exponential backoff: 1s, 2s, 4s, 8s, 16s (max 5 attempts)
sch.Delay(
    "api-call",
    time.Second,
    func(ctx context.Context) error {
        return callExternalAPI()
    },
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 5)),
)
```

### Linear Retry

```go
// Retry with fixed delay: 1s, 1s, 1s, 1s, 1s (max 5 attempts)
sch.Queue("task", taskFunc,
    scheduler.WithRetry(scheduler.RetryLinear(time.Second, 5)),
)
```

## Retry Strategies

### 1. Exponential Backoff

Doubles the wait time after each failure.

```go
func RetryExponential(initialDelay time.Duration, maxAttempts int) RetryPolicy
```

**Delays**: `initial`, `2×initial`, `4×initial`, `8×initial`, ...

**Example**:
```go
// 1s, 2s, 4s, 8s, 16s
policy := scheduler.RetryExponential(time.Second, 5)

sch.Queue("webhook", deliverWebhook,
    scheduler.WithRetry(policy),
)
```

**Best for**: External APIs, network requests (handles rate limits and temporary outages)

### 2. Linear Backoff

Uses fixed delay between retries.

```go
func RetryLinear(delay time.Duration, maxAttempts int) RetryPolicy
```

**Delays**: `delay`, `delay`, `delay`, ...

**Example**:
```go
// 5s, 5s, 5s, 5s, 5s
policy := scheduler.RetryLinear(5*time.Second, 5)

sch.Queue("db-operation", dbOperation,
    scheduler.WithRetry(policy),
)
```

**Best for**: Operations with predictable failure duration

### 3. No Retry

Default behavior when no retry policy specified.

```go
// No retry - fail immediately on error
sch.Queue("task", taskFunc)
// Equivalent to: scheduler.WithRetry(scheduler.RetryNever())
```

## Retry Policy Configuration

### MaxAttempts

Total number of attempts (including initial attempt):

```go
// 1 initial + 4 retries = 5 total attempts
policy := scheduler.RetryExponential(time.Second, 5)
```

### Initial Delay

Starting delay for exponential/linear backoff:

```go
// Start with 2 second delay
policy := scheduler.RetryExponential(2*time.Second, 5)
// Delays: 2s, 4s, 8s, 16s, 32s
```

### Maximum Backoff

Limit exponential growth:

```go
// Cap backoff at 1 minute
policy := scheduler.RetryExponential(time.Second, 10).
    WithMaxDelay(time.Minute)
// Delays: 1s, 2s, 4s, 8s, 16s, 32s, 60s, 60s, 60s, 60s
```

## Practical Examples

### 1. Webhook Delivery

```go
func deliverWebhook(webhookURL string, payload []byte) error {
    sch.Queue(
        "webhook-delivery",
        func(ctx context.Context) error {
            resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(payload))
            if err != nil {
                return fmt.Errorf("webhook delivery failed: %w", err)
            }
            defer resp.Body.Close()

            if resp.StatusCode >= 500 {
                // Server error - retry
                return fmt.Errorf("server error: %d", resp.StatusCode)
            }

            if resp.StatusCode >= 400 {
                // Client error - don't retry
                return nil
            }

            return nil
        },
        scheduler.WithRetry(scheduler.RetryExponential(time.Second, 5)),
        scheduler.WithTimeout(30*time.Second),
    )

    return nil
}
```

### 2. Database Operations with Deadlock

```go
func updateInventory(productID string, quantity int) error {
    return sch.Queue(
        fmt.Sprintf("update-inventory-%s", productID),
        func(ctx context.Context) error {
            _, err := db.ExecContext(ctx,
                "UPDATE products SET quantity = quantity + ? WHERE id = ?",
                quantity, productID,
            )

            // Retry on deadlock
            if isDeadlockError(err) {
                return err // Triggers retry
            }

            return nil
        },
        scheduler.WithRetry(scheduler.RetryLinear(100*time.Millisecond, 3)),
    )
}
```

### 3. External API with Rate Limiting

```go
func fetchUserData(userID string) error {
    return sch.Queue(
        "fetch-user-data",
        func(ctx context.Context) error {
            resp, err := externalAPI.GetUser(userID)
            if err != nil {
                return err
            }

            // Check for rate limit
            if resp.StatusCode == http.StatusTooManyRequests {
                // Parse Retry-After header
                retryAfter := resp.Header.Get("Retry-After")
                log.Printf("Rate limited, retry after: %s", retryAfter)
                return errors.New("rate limited")
            }

            return processUserData(resp.Body)
        },
        scheduler.WithRetry(scheduler.RetryExponential(5*time.Second, 3)),
        scheduler.WithTimeout(30*time.Second),
    )
}
```

### 4. File Operations

```go
func processFile(filepath string) error {
    return sch.Queue(
        "process-file",
        func(ctx context.Context) error {
            // Try to acquire file lock
            file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
            if err != nil {
                if errors.Is(err, os.ErrPermission) {
                    // Permission error - don't retry
                    return nil
                }
                // Other errors - retry
                return err
            }
            defer file.Close()

            // Process file
            return processFileContent(file)
        },
        scheduler.WithRetry(scheduler.RetryLinear(500*time.Millisecond, 5)),
    )
}
```

### 5. Message Queue Consumer

```go
func consumeMessage(msg Message) error {
    return sch.Queue(
        fmt.Sprintf("consume-%s", msg.ID),
        func(ctx context.Context) error {
            if err := processMessage(msg); err != nil {
                log.Printf("Message processing failed: %v", err)
                return err // Triggers retry
            }

            // Acknowledge message
            return acknowledgeMessage(msg.ID)
        },
        scheduler.WithRetry(scheduler.RetryExponential(time.Second, 3)),
    )
}
```

## Advanced Patterns

### 1. Conditional Retry

Only retry specific errors:

```go
type RetryableError struct {
    error
}

func (e RetryableError) Retryable() bool { return true }

sch.Queue("task", func(ctx context.Context) error {
    err := performOperation()
    if err != nil {
        if isTransientError(err) {
            // Wrap to indicate retryable
            return RetryableError{err}
        }
        // Permanent error - don't retry
        return nil
    }
    return nil
},
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 5)),
)
```

### 2. Custom Backoff Strategy

```go
type CustomRetryPolicy struct {
    delays []time.Duration
}

func (p *CustomRetryPolicy) ShouldRetry(attempt int) (bool, time.Duration) {
    if attempt >= len(p.delays) {
        return false, 0
    }
    return true, p.delays[attempt]
}

// Usage: 1s, 5s, 10s, 30s, 1m
customPolicy := &CustomRetryPolicy{
    delays: []time.Duration{
        1 * time.Second,
        5 * time.Second,
        10 * time.Second,
        30 * time.Second,
        1 * time.Minute,
    },
}
```

### 3. Retry with Jitter

Add randomness to prevent thundering herd:

```go
func exponentialBackoffWithJitter(base time.Duration, attempt int) time.Duration {
    // Calculate exponential delay
    delay := base * time.Duration(1<<uint(attempt))

    // Add jitter (±20%)
    jitter := time.Duration(rand.Float64() * 0.4 * float64(delay))
    if rand.Float64() < 0.5 {
        return delay - jitter
    }
    return delay + jitter
}
```

### 4. Circuit Breaker Pattern

Stop retrying after consecutive failures:

```go
type CircuitBreaker struct {
    failures    int
    maxFailures int
    resetTime   time.Time
    mu          sync.Mutex
}

func (cb *CircuitBreaker) Allow() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    // Reset if timeout elapsed
    if time.Now().After(cb.resetTime) {
        cb.failures = 0
    }

    // Check if circuit open (too many failures)
    return cb.failures < cb.maxFailures
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failures++
    cb.resetTime = time.Now().Add(time.Minute)
}

// Usage
var breaker = &CircuitBreaker{maxFailures: 5}

sch.Queue("api-call", func(ctx context.Context) error {
    if !breaker.Allow() {
        return errors.New("circuit breaker open")
    }

    err := callAPI()
    if err != nil {
        breaker.RecordFailure()
        return err
    }

    return nil
},
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 3)),
)
```

## Retry Metrics

### Track Retry Attempts

```go
var (
    retryCounter = metrics.NewCounter("task_retries", "Total retry attempts")
    retryGauge   = metrics.NewGauge("task_retry_attempts", "Current retry attempt")
)

func wrapWithMetrics(name string, fn scheduler.TaskFunc, maxAttempts int) scheduler.TaskFunc {
    attempts := 0

    return func(ctx context.Context) error {
        attempts++
        retryGauge.WithLabelValues(name).Set(float64(attempts))

        err := fn(ctx)

        if err != nil && attempts < maxAttempts {
            retryCounter.WithLabelValues(name).Inc()
        }

        return err
    }
}
```

## Best Practices

### 1. Choose Appropriate Strategy

```go
// ✅ Exponential: External APIs, rate limits
scheduler.WithRetry(scheduler.RetryExponential(time.Second, 5))

// ✅ Linear: Database deadlocks, file locks
scheduler.WithRetry(scheduler.RetryLinear(100*time.Millisecond, 3))

// ✅ No retry: Idempotent operations, non-critical tasks
// (no retry policy)
```

### 2. Set Reasonable Max Attempts

```go
// ✅ Good: 3-5 attempts for most cases
scheduler.WithRetry(scheduler.RetryExponential(time.Second, 3))

// ⚠️ Too many: Wastes resources
scheduler.WithRetry(scheduler.RetryExponential(time.Second, 20))

// ⚠️ Too few: May not recover from transient issues
scheduler.WithRetry(scheduler.RetryExponential(time.Second, 1))
```

### 3. Log Retry Attempts

```go
sch.Queue("task", func(ctx context.Context) error {
    attempt := getAttemptFromContext(ctx) // Hypothetical
    log.Printf("Attempt %d: executing task", attempt)

    err := performTask()
    if err != nil {
        log.Printf("Attempt %d failed: %v", attempt, err)
    }

    return err
},
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 5)),
)
```

### 4. Combine with Timeout

```go
// Prevent individual attempts from hanging
sch.Queue("task", taskFunc,
    scheduler.WithTimeout(10*time.Second),  // Per-attempt timeout
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 3)),
)
```

### 5. Handle Permanent Failures

```go
sch.Queue("task", func(ctx context.Context) error {
    err := performOperation()
    if err != nil {
        if isPermanentError(err) {
            log.Printf("Permanent failure, not retrying: %v", err)
            return nil // Don't retry
        }
        return err // Retry
    }
    return nil
})
```

## Testing

### Test Retry Behavior

```go
func TestRetryPolicy(t *testing.T) {
    attempts := 0
    maxAttempts := 3

    sch := scheduler.New(scheduler.WithWorkers(1))
    sch.Start()
    defer sch.Stop()

    sch.Queue("test-retry", func(ctx context.Context) error {
        attempts++
        if attempts < maxAttempts {
            return errors.New("transient error")
        }
        return nil // Success on 3rd attempt
    },
        scheduler.WithRetry(scheduler.RetryLinear(10*time.Millisecond, maxAttempts)),
    )

    // Wait for retries
    time.Sleep(100 * time.Millisecond)

    if attempts != maxAttempts {
        t.Errorf("Expected %d attempts, got %d", maxAttempts, attempts)
    }
}
```

### Test Exponential Backoff

```go
func TestExponentialBackoff(t *testing.T) {
    policy := scheduler.RetryExponential(time.Second, 5)

    expectedDelays := []time.Duration{
        time.Second,
        2 * time.Second,
        4 * time.Second,
        8 * time.Second,
        16 * time.Second,
    }

    for i, expected := range expectedDelays {
        shouldRetry, delay := policy.ShouldRetry(i)
        if !shouldRetry {
            t.Errorf("Attempt %d: expected retry", i)
        }
        if delay != expected {
            t.Errorf("Attempt %d: delay = %v, want %v", i, delay, expected)
        }
    }

    // Should not retry after max attempts
    shouldRetry, _ := policy.ShouldRetry(5)
    if shouldRetry {
        t.Error("Should not retry after max attempts")
    }
}
```

## Troubleshooting

### Too Many Retries

**Problem**: Task retrying indefinitely

**Check**:
- Max attempts configured correctly
- Task eventually succeeds or returns nil
- Not catching and re-throwing errors incorrectly

### Retry Not Happening

**Problem**: Task fails without retry

**Check**:
- Retry policy actually set
- Task returning error (not panicking)
- Scheduler has available workers

### Long Delays

**Problem**: Exponential backoff causing very long delays

**Solution**: Add max delay cap

```go
policy := scheduler.RetryExponential(time.Second, 10).
    WithMaxDelay(time.Minute)
```

## Related Documentation

- [Scheduler Overview](README.md) — Scheduler module overview
- [Cron Jobs](cron.md) — Recurring task scheduling
- [Delayed Tasks](delayed-tasks.md) — Future task execution
- [Examples](examples.md) — Complete examples

## Reference Implementation

See examples:
- `examples/scheduler/` — Retry policy examples
- `examples/reference/` — Production retry patterns
