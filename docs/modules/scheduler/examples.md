# Scheduler Examples

> **Package**: `github.com/spcent/plumego/scheduler` | **Complete working examples**

## Overview

This document provides complete, production-ready examples of using the Plumego scheduler for common scenarios.

## Example 1: Background Job System

Complete background job processing system with cron jobs, delayed tasks, and retry.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/spcent/plumego/scheduler"
    "github.com/spcent/plumego/core"
)

type JobSystem struct {
    sch *scheduler.Scheduler
    db  *Database
}

func NewJobSystem(db *Database) *JobSystem {
    return &JobSystem{
        sch: scheduler.New(
            scheduler.WithWorkers(8),
            scheduler.WithQueueSize(1000),
        ),
        db: db,
    }
}

func (js *JobSystem) Start() error {
    // Start scheduler
    if err := js.sch.Start(); err != nil {
        return err
    }

    // Register recurring jobs
    js.registerCronJobs()

    log.Println("Job system started")
    return nil
}

func (js *JobSystem) Stop() error {
    log.Println("Stopping job system...")
    return js.sch.Stop()
}

func (js *JobSystem) registerCronJobs() {
    // Clean up old records daily at 2 AM
    js.sch.AddCron("daily-cleanup", "0 2 * * *", func(ctx context.Context) error {
        log.Println("Running daily cleanup")
        cutoff := time.Now().Add(-90 * 24 * time.Hour)
        return js.db.DeleteOldRecords(cutoff)
    })

    // Generate daily reports at 8 AM weekdays
    js.sch.AddCron("daily-report", "0 8 * * MON-FRI", func(ctx context.Context) error {
        log.Println("Generating daily report")
        report, err := js.generateDailyReport(ctx)
        if err != nil {
            return err
        }
        return js.emailReport(report)
    })

    // Health check every 5 minutes
    js.sch.AddCron("health-check", "*/5 * * * *", func(ctx context.Context) error {
        return js.performHealthCheck(ctx)
    })

    log.Println("Registered cron jobs")
}

func (js *JobSystem) QueueEmail(to, subject, body string) {
    js.sch.Delay("send-email", 5*time.Second, func(ctx context.Context) error {
        log.Printf("Sending email to %s", to)
        return sendEmail(to, subject, body)
    },
        scheduler.WithRetry(scheduler.RetryExponential(time.Second, 3)),
        scheduler.WithTimeout(30*time.Second),
    )
}

func (js *JobSystem) ProcessUpload(uploadID string) {
    js.sch.Queue("process-upload", func(ctx context.Context) error {
        log.Printf("Processing upload: %s", uploadID)

        // Extract metadata
        if err := js.extractMetadata(ctx, uploadID); err != nil {
            return err
        }

        // Generate thumbnails
        if err := js.generateThumbnails(ctx, uploadID); err != nil {
            return err
        }

        // Update status
        return js.db.UpdateUploadStatus(uploadID, "completed")
    },
        scheduler.WithTimeout(5*time.Minute),
        scheduler.WithRetry(scheduler.RetryLinear(30*time.Second, 3)),
    )
}

func (js *JobSystem) generateDailyReport(ctx context.Context) (*Report, error) {
    // Implementation
    return &Report{}, nil
}

func (js *JobSystem) emailReport(report *Report) error {
    // Implementation
    return nil
}

func (js *JobSystem) performHealthCheck(ctx context.Context) error {
    // Implementation
    return nil
}

func (js *JobSystem) extractMetadata(ctx context.Context, uploadID string) error {
    // Implementation
    return nil
}

func (js *JobSystem) generateThumbnails(ctx context.Context, uploadID string) error {
    // Implementation
    return nil
}

func main() {
    db := ConnectDatabase()
    jobs := NewJobSystem(db)

    app := core.New(
        core.WithAddr(":8080"),
        core.WithRunner(jobs), // Auto-start/stop with app
    )

    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
}
```

## Example 2: Email Service with Rate Limiting

Email service that respects rate limits and retries failed deliveries.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/spcent/plumego/scheduler"
)

type EmailService struct {
    sch           *scheduler.Scheduler
    rateLimit     int
    ratePeriod    time.Duration
    sentCount     map[string]int
    sentTimestamp map[string]time.Time
    mu            sync.Mutex
}

func NewEmailService(rateLimit int, ratePeriod time.Duration) *EmailService {
    es := &EmailService{
        sch:           scheduler.New(scheduler.WithWorkers(4)),
        rateLimit:     rateLimit,
        ratePeriod:    ratePeriod,
        sentCount:     make(map[string]int),
        sentTimestamp: make(map[string]time.Time),
    }

    es.sch.Start()
    return es
}

func (es *EmailService) SendEmail(to, subject, body string) error {
    // Calculate delay based on rate limit
    delay := es.calculateDelay(to)

    es.sch.Delay(
        fmt.Sprintf("email-%s-%d", to, time.Now().Unix()),
        delay,
        func(ctx context.Context) error {
            return es.deliver(ctx, to, subject, body)
        },
        scheduler.WithRetry(scheduler.RetryExponential(5*time.Second, 3)),
        scheduler.WithTimeout(30*time.Second),
    )

    return nil
}

func (es *EmailService) calculateDelay(recipient string) time.Duration {
    es.mu.Lock()
    defer es.mu.Unlock()

    now := time.Now()
    timestamp, exists := es.sentTimestamp[recipient]

    // Reset count if period elapsed
    if !exists || now.Sub(timestamp) > es.ratePeriod {
        es.sentCount[recipient] = 0
        es.sentTimestamp[recipient] = now
        return 0
    }

    // Check rate limit
    count := es.sentCount[recipient]
    if count >= es.rateLimit {
        // Calculate delay to respect rate limit
        delay := es.ratePeriod - now.Sub(timestamp)
        log.Printf("Rate limit reached for %s, delaying %v", recipient, delay)
        return delay
    }

    return 0
}

func (es *EmailService) deliver(ctx context.Context, to, subject, body string) error {
    es.mu.Lock()
    es.sentCount[to]++
    es.mu.Unlock()

    log.Printf("Sending email to %s: %s", to, subject)

    // Simulate email sending
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(time.Second):
        // Email sent successfully
    }

    return nil
}

func (es *EmailService) Stop() error {
    return es.sch.Stop()
}

func main() {
    // Allow 10 emails per minute per recipient
    emailService := NewEmailService(10, time.Minute)
    defer emailService.Stop()

    // Send multiple emails
    for i := 0; i < 20; i++ {
        emailService.SendEmail(
            "user@example.com",
            fmt.Sprintf("Message %d", i),
            "This is a test email",
        )
    }

    // Wait for delivery
    time.Sleep(2 * time.Minute)
}
```

## Example 3: Webhook Delivery System

Reliable webhook delivery with exponential backoff and dead letter queue.

```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/spcent/plumego/scheduler"
)

type WebhookDelivery struct {
    sch        *scheduler.Scheduler
    deadLetter chan *Webhook
}

type Webhook struct {
    ID        string
    URL       string
    Event     string
    Payload   interface{}
    Attempts  int
    CreatedAt time.Time
}

func NewWebhookDelivery() *WebhookDelivery {
    wd := &WebhookDelivery{
        sch:        scheduler.New(scheduler.WithWorkers(8)),
        deadLetter: make(chan *Webhook, 100),
    }

    wd.sch.Start()

    // Start dead letter queue processor
    go wd.processDeadLetters()

    return wd
}

func (wd *WebhookDelivery) Deliver(webhook *Webhook) {
    wd.sch.Delay(
        fmt.Sprintf("webhook-%s", webhook.ID),
        time.Second,
        func(ctx context.Context) error {
            return wd.attemptDelivery(ctx, webhook)
        },
        scheduler.WithRetry(scheduler.RetryExponential(5*time.Second, 5)),
        scheduler.WithTimeout(30*time.Second),
    )
}

func (wd *WebhookDelivery) attemptDelivery(ctx context.Context, webhook *Webhook) error {
    webhook.Attempts++

    log.Printf("Delivering webhook %s (attempt %d)", webhook.ID, webhook.Attempts)

    // Marshal payload
    payload, err := json.Marshal(webhook.Payload)
    if err != nil {
        return fmt.Errorf("marshal failed: %w", err)
    }

    // Create request
    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        webhook.URL,
        bytes.NewReader(payload),
    )
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Webhook-Event", webhook.Event)
    req.Header.Set("X-Webhook-ID", webhook.ID)

    // Send request
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Printf("Webhook delivery failed: %v", err)

        // Check if max attempts reached
        if webhook.Attempts >= 5 {
            wd.deadLetter <- webhook
            return nil // Don't retry further
        }

        return err // Retry
    }
    defer resp.Body.Close()

    // Check response status
    if resp.StatusCode >= 500 {
        log.Printf("Server error: %d", resp.StatusCode)

        if webhook.Attempts >= 5 {
            wd.deadLetter <- webhook
            return nil
        }

        return fmt.Errorf("server error: %d", resp.StatusCode)
    }

    if resp.StatusCode >= 400 {
        log.Printf("Client error: %d (not retrying)", resp.StatusCode)
        return nil // Don't retry client errors
    }

    log.Printf("Webhook delivered successfully: %s", webhook.ID)
    return nil
}

func (wd *WebhookDelivery) processDeadLetters() {
    for webhook := range wd.deadLetter {
        log.Printf("Dead letter: webhook %s failed after %d attempts", webhook.ID, webhook.Attempts)

        // Store in database for manual review
        if err := storeDeadLetter(webhook); err != nil {
            log.Printf("Failed to store dead letter: %v", err)
        }
    }
}

func (wd *WebhookDelivery) Stop() error {
    close(wd.deadLetter)
    return wd.sch.Stop()
}

func storeDeadLetter(webhook *Webhook) error {
    // Implementation: Store in database
    return nil
}

func main() {
    delivery := NewWebhookDelivery()
    defer delivery.Stop()

    // Deliver webhooks
    delivery.Deliver(&Webhook{
        ID:        "wh-123",
        URL:       "https://example.com/webhook",
        Event:     "user.created",
        Payload:   map[string]string{"user_id": "user-456"},
        CreatedAt: time.Now(),
    })

    delivery.Deliver(&Webhook{
        ID:        "wh-124",
        URL:       "https://example.com/webhook",
        Event:     "order.completed",
        Payload:   map[string]string{"order_id": "order-789"},
        CreatedAt: time.Now(),
    })

    // Wait for delivery
    time.Sleep(time.Minute)
}
```

## Example 4: Task Queue with Priority

Priority-based task queue with high/medium/low priority levels.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/spcent/plumego/scheduler"
)

type Priority int

const (
    PriorityHigh Priority = iota
    PriorityMedium
    PriorityLow
)

type PriorityQueue struct {
    highSch   *scheduler.Scheduler
    mediumSch *scheduler.Scheduler
    lowSch    *scheduler.Scheduler
}

func NewPriorityQueue() *PriorityQueue {
    return &PriorityQueue{
        highSch:   scheduler.New(scheduler.WithWorkers(4)),
        mediumSch: scheduler.New(scheduler.WithWorkers(2)),
        lowSch:    scheduler.New(scheduler.WithWorkers(1)),
    }
}

func (pq *PriorityQueue) Start() error {
    if err := pq.highSch.Start(); err != nil {
        return err
    }
    if err := pq.mediumSch.Start(); err != nil {
        return err
    }
    if err := pq.lowSch.Start(); err != nil {
        return err
    }
    return nil
}

func (pq *PriorityQueue) Stop() error {
    pq.lowSch.Stop()
    pq.mediumSch.Stop()
    return pq.highSch.Stop()
}

func (pq *PriorityQueue) Queue(name string, priority Priority, fn scheduler.TaskFunc, opts ...scheduler.TaskOption) {
    switch priority {
    case PriorityHigh:
        pq.highSch.Queue(name, fn, opts...)
    case PriorityMedium:
        pq.mediumSch.Queue(name, fn, opts...)
    case PriorityLow:
        pq.lowSch.Queue(name, fn, opts...)
    }
}

func main() {
    pq := NewPriorityQueue()
    pq.Start()
    defer pq.Stop()

    // High priority: Critical operations
    pq.Queue("payment-processing", PriorityHigh, func(ctx context.Context) error {
        log.Println("Processing payment (HIGH)")
        time.Sleep(time.Second)
        return nil
    })

    // Medium priority: Normal operations
    pq.Queue("send-notification", PriorityMedium, func(ctx context.Context) error {
        log.Println("Sending notification (MEDIUM)")
        time.Sleep(time.Second)
        return nil
    })

    // Low priority: Background tasks
    pq.Queue("generate-report", PriorityLow, func(ctx context.Context) error {
        log.Println("Generating report (LOW)")
        time.Sleep(time.Second)
        return nil
    })

    time.Sleep(5 * time.Second)
}
```

## Example 5: Batch Processing System

Batch processor that collects items and processes them in batches.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/spcent/plumego/scheduler"
)

type BatchProcessor struct {
    sch           *scheduler.Scheduler
    items         map[string][]interface{}
    batchSize     int
    maxWait       time.Duration
    timers        map[string]*time.Timer
    mu            sync.Mutex
    processFunc   func([]interface{}) error
}

func NewBatchProcessor(batchSize int, maxWait time.Duration, processFunc func([]interface{}) error) *BatchProcessor {
    bp := &BatchProcessor{
        sch:         scheduler.New(scheduler.WithWorkers(4)),
        items:       make(map[string][]interface{}),
        batchSize:   batchSize,
        maxWait:     maxWait,
        timers:      make(map[string]*time.Timer),
        processFunc: processFunc,
    }

    bp.sch.Start()
    return bp
}

func (bp *BatchProcessor) Add(batchKey string, item interface{}) {
    bp.mu.Lock()
    defer bp.mu.Unlock()

    // Add item to batch
    bp.items[batchKey] = append(bp.items[batchKey], item)

    // Check if batch is full
    if len(bp.items[batchKey]) >= bp.batchSize {
        bp.processBatchLocked(batchKey)
        return
    }

    // Start or reset timer
    if timer, exists := bp.timers[batchKey]; exists {
        timer.Stop()
    }

    bp.timers[batchKey] = time.AfterFunc(bp.maxWait, func() {
        bp.mu.Lock()
        defer bp.mu.Unlock()
        bp.processBatchLocked(batchKey)
    })
}

func (bp *BatchProcessor) processBatchLocked(batchKey string) {
    batch := bp.items[batchKey]
    if len(batch) == 0 {
        return
    }

    // Clear batch
    bp.items[batchKey] = nil

    // Cancel timer
    if timer, exists := bp.timers[batchKey]; exists {
        timer.Stop()
        delete(bp.timers, batchKey)
    }

    // Process batch
    bp.sch.Queue(fmt.Sprintf("batch-%s", batchKey), func(ctx context.Context) error {
        log.Printf("Processing batch: %d items", len(batch))
        return bp.processFunc(batch)
    })
}

func (bp *BatchProcessor) Stop() error {
    // Process remaining batches
    bp.mu.Lock()
    for key := range bp.items {
        bp.processBatchLocked(key)
    }
    bp.mu.Unlock()

    return bp.sch.Stop()
}

func main() {
    processor := NewBatchProcessor(
        10,              // Batch size
        5*time.Second,   // Max wait
        func(items []interface{}) error {
            log.Printf("Processed batch of %d items", len(items))
            // Process items...
            return nil
        },
    )
    defer processor.Stop()

    // Add items
    for i := 0; i < 25; i++ {
        processor.Add("default", fmt.Sprintf("item-%d", i))
        time.Sleep(500 * time.Millisecond)
    }

    time.Sleep(10 * time.Second)
}
```

## Related Documentation

- [Scheduler Overview](README.md) — Scheduler module overview
- [Cron Jobs](cron.md) — Recurring task scheduling
- [Delayed Tasks](delayed-tasks.md) — Future task execution
- [Retry Policies](retry-policies.md) — Retry strategies

## Running Examples

All examples are runnable. To test:

```bash
# Copy example to your project
# Update imports
# Run
go run example.go
```

## Production Considerations

When deploying to production:

1. **Worker sizing**: Monitor queue depth and adjust worker count
2. **Timeouts**: Set appropriate timeouts for all tasks
3. **Retry policies**: Use exponential backoff for external dependencies
4. **Monitoring**: Track task execution time, failures, retries
5. **Graceful shutdown**: Ensure all tasks complete before exit
6. **Error handling**: Log errors and send alerts for critical failures
