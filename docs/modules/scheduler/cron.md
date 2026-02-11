# Cron Jobs

> **Package**: `github.com/spcent/plumego/scheduler` | **Feature**: Recurring tasks

## Overview

Cron jobs enable recurring task execution using cron expressions. The scheduler uses a cron parser that supports standard Unix cron syntax with second-level precision.

**Use Cases**:
- Daily/weekly/monthly reports
- Database cleanup
- Cache warming
- Health checks
- Data synchronization
- Backup tasks

## Quick Start

### Basic Cron Job

```go
import "github.com/spcent/plumego/scheduler"

sch := scheduler.New(scheduler.WithWorkers(4))
sch.Start()

// Run every hour at minute 0
sch.AddCron("hourly-cleanup", "0 * * * *", func(ctx context.Context) error {
    log.Println("Running hourly cleanup")
    return cleanupTempFiles()
})
```

### Multiple Cron Jobs

```go
// Daily report at 8:00 AM
sch.AddCron("daily-report", "0 8 * * *", generateDailyReport)

// Weekly backup every Sunday at 2:00 AM
sch.AddCron("weekly-backup", "0 2 * * 0", performBackup)

// Clean cache every 5 minutes
sch.AddCron("cache-cleanup", "*/5 * * * *", cleanCache)

// End-of-month processing
sch.AddCron("monthly-close", "0 0 1 * *", monthlyClose)
```

## Cron Expression Syntax

### Standard Format (5 fields)

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

### Field Values

| Field | Values | Special Characters |
|-------|--------|-------------------|
| Minute | 0-59 | `*` `,` `-` `/` |
| Hour | 0-23 | `*` `,` `-` `/` |
| Day of Month | 1-31 | `*` `,` `-` `/` `?` |
| Month | 1-12 or JAN-DEC | `*` `,` `-` `/` |
| Day of Week | 0-6 or SUN-SAT | `*` `,` `-` `/` `?` |

### Special Characters

- **`*`** (asterisk): All values (every minute, every hour, etc.)
- **`,`** (comma): List of values (`1,15,30` = 1st, 15th, and 30th)
- **`-`** (hyphen): Range of values (`1-5` = 1, 2, 3, 4, 5)
- **`/`** (slash): Step values (`*/15` = every 15 units)
- **`?`** (question mark): No specific value (day of month or day of week)

## Common Cron Expressions

### Every N Minutes

```go
// Every minute
"* * * * *"

// Every 5 minutes
"*/5 * * * *"

// Every 15 minutes
"*/15 * * * *"

// Every 30 minutes
"*/30 * * * *"
```

### Every N Hours

```go
// Every hour
"0 * * * *"

// Every 2 hours
"0 */2 * * *"

// Every 6 hours
"0 */6 * * *"

// Every 12 hours
"0 */12 * * *"
```

### Specific Times

```go
// Every day at 8:00 AM
"0 8 * * *"

// Every day at 2:30 PM
"30 14 * * *"

// Every day at midnight
"0 0 * * *"

// Every day at noon
"0 12 * * *"
```

### Weekdays vs Weekends

```go
// Monday through Friday at 9:00 AM
"0 9 * * 1-5"

// Weekdays only (Mon-Fri)
"0 9 * * MON-FRI"

// Saturday and Sunday
"0 9 * * 0,6"
"0 9 * * SAT,SUN"

// Only on Monday
"0 9 * * 1"
"0 9 * * MON"
```

### Monthly

```go
// First day of every month at midnight
"0 0 1 * *"

// Last day of month (use day 28-31 with logic)
"0 0 28-31 * *"

// 15th of every month at 3:00 PM
"0 15 15 * *"

// First Monday of every month
"0 9 1-7 * 1"
```

### Quarterly/Yearly

```go
// First day of quarter (Jan, Apr, Jul, Oct)
"0 0 1 1,4,7,10 *"

// First day of year
"0 0 1 1 *"
```

## Practical Examples

### 1. Database Cleanup

```go
// Delete old records every day at 2:00 AM
sch.AddCron("db-cleanup", "0 2 * * *", func(ctx context.Context) error {
    cutoff := time.Now().Add(-90 * 24 * time.Hour) // 90 days ago

    result, err := db.ExecContext(ctx,
        "DELETE FROM logs WHERE created_at < ?",
        cutoff,
    )
    if err != nil {
        return fmt.Errorf("cleanup failed: %w", err)
    }

    rows, _ := result.RowsAffected()
    log.Printf("Deleted %d old log entries", rows)
    return nil
})
```

### 2. Cache Warming

```go
// Warm cache every 5 minutes
sch.AddCron("cache-warm", "*/5 * * * *", func(ctx context.Context) error {
    // Fetch frequently accessed data
    popularItems, err := db.GetPopularItems(ctx, 100)
    if err != nil {
        return err
    }

    // Populate cache
    for _, item := range popularItems {
        cache.Set(fmt.Sprintf("item:%s", item.ID), item, 10*time.Minute)
    }

    log.Printf("Warmed cache with %d items", len(popularItems))
    return nil
})
```

### 3. Daily Reports

```go
// Generate and email daily report at 8:00 AM weekdays
sch.AddCron("daily-report", "0 8 * * MON-FRI", func(ctx context.Context) error {
    // Generate report for yesterday
    yesterday := time.Now().Add(-24 * time.Hour)
    report, err := generateReport(ctx, yesterday)
    if err != nil {
        return fmt.Errorf("report generation failed: %w", err)
    }

    // Email to recipients
    recipients := []string{"admin@example.com", "team@example.com"}
    for _, email := range recipients {
        if err := sendEmail(email, "Daily Report", report); err != nil {
            log.Printf("Failed to send report to %s: %v", email, err)
        }
    }

    return nil
})
```

### 4. Session Cleanup

```go
// Clean expired sessions every hour
sch.AddCron("session-cleanup", "0 * * * *", func(ctx context.Context) error {
    result, err := db.ExecContext(ctx,
        "DELETE FROM sessions WHERE expires_at < NOW()",
    )
    if err != nil {
        return err
    }

    rows, _ := result.RowsAffected()
    log.Printf("Cleaned up %d expired sessions", rows)
    return nil
})
```

### 5. Health Check Pings

```go
// Ping external health check service every 5 minutes
sch.AddCron("health-ping", "*/5 * * * *", func(ctx context.Context) error {
    resp, err := http.Get("https://healthchecks.io/ping/your-uuid")
    if err != nil {
        return fmt.Errorf("health ping failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("health ping returned %d", resp.StatusCode)
    }

    return nil
})
```

### 6. Weekly Backup

```go
// Full database backup every Sunday at 2:00 AM
sch.AddCron("weekly-backup", "0 2 * * 0", func(ctx context.Context) error {
    timestamp := time.Now().Format("20060102-150405")
    filename := fmt.Sprintf("backup-%s.sql.gz", timestamp)

    // Create backup
    if err := createDatabaseBackup(ctx, filename); err != nil {
        return fmt.Errorf("backup failed: %w", err)
    }

    // Upload to S3
    if err := uploadToS3(filename); err != nil {
        return fmt.Errorf("upload failed: %w", err)
    }

    log.Printf("Backup completed: %s", filename)
    return nil
},
    scheduler.WithTimeout(30*time.Minute), // Allow 30 minutes for backup
)
```

## Task Options

### Timeout

```go
// Set timeout for long-running cron job
sch.AddCron("long-task", "0 2 * * *", longRunningTask,
    scheduler.WithTimeout(10*time.Minute),
)
```

### Retry Policy

```go
// Retry on failure with exponential backoff
sch.AddCron("api-sync", "*/15 * * * *", syncExternalAPI,
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 3)),
)
```

## Managing Cron Jobs

### Remove Cron Job

```go
// Remove cron job by name
removed := sch.Remove("hourly-cleanup")
if removed {
    log.Println("Cron job removed")
}
```

### List Active Jobs

```go
// Get list of active cron jobs
jobs := sch.ListCronJobs()
for _, job := range jobs {
    log.Printf("Job: %s, Next: %s", job.Name, job.Next)
}
```

### Pause and Resume

```go
// Pause cron job (stop scheduling, but keep definition)
sch.Pause("daily-report")

// Resume cron job
sch.Resume("daily-report")
```

## Best Practices

### 1. Use Descriptive Names

```go
// ❌ Bad: Generic names
sch.AddCron("job1", "0 * * * *", task1)
sch.AddCron("cleanup", "0 2 * * *", task2)

// ✅ Good: Descriptive names
sch.AddCron("hourly-cache-refresh", "0 * * * *", refreshCache)
sch.AddCron("daily-log-cleanup", "0 2 * * *", cleanupLogs)
```

### 2. Set Appropriate Timeouts

```go
// Quick tasks: short timeout
sch.AddCron("health-check", "*/5 * * * *", pingHealth,
    scheduler.WithTimeout(10*time.Second),
)

// Long tasks: longer timeout
sch.AddCron("database-backup", "0 2 * * 0", backupDB,
    scheduler.WithTimeout(1*time.Hour),
)
```

### 3. Handle Context Cancellation

```go
sch.AddCron("data-sync", "*/10 * * * *", func(ctx context.Context) error {
    for _, item := range items {
        // Check for cancellation
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // Process item
        if err := processItem(ctx, item); err != nil {
            return err
        }
    }
    return nil
})
```

### 4. Log Execution

```go
sch.AddCron("daily-report", "0 8 * * *", func(ctx context.Context) error {
    start := time.Now()
    log.Printf("Starting daily report generation")

    err := generateReport(ctx)

    log.Printf("Daily report completed in %v (error: %v)", time.Since(start), err)
    return err
})
```

### 5. Avoid Overlapping Runs

```go
var mu sync.Mutex

sch.AddCron("heavy-task", "*/5 * * * *", func(ctx context.Context) error {
    // Prevent concurrent execution
    if !mu.TryLock() {
        log.Println("Previous run still in progress, skipping")
        return nil
    }
    defer mu.Unlock()

    // Execute task
    return heavyTask(ctx)
})
```

## Testing Cron Jobs

### Unit Test

```go
func TestCronJob(t *testing.T) {
    sch := scheduler.New(scheduler.WithWorkers(1))
    sch.Start()
    defer sch.Stop()

    executed := false
    err := sch.AddCron("test-job", "* * * * *", func(ctx context.Context) error {
        executed = true
        return nil
    })
    if err != nil {
        t.Fatalf("AddCron failed: %v", err)
    }

    // Wait for next minute boundary (plus buffer)
    time.Sleep(65 * time.Second)

    if !executed {
        t.Error("Cron job not executed")
    }
}
```

### Test Cron Expression

```go
func TestCronExpression(t *testing.T) {
    tests := []struct {
        expr  string
        valid bool
    }{
        {"* * * * *", true},
        {"0 8 * * *", true},
        {"invalid", false},
        {"60 * * * *", false}, // Invalid minute
    }

    for _, tt := range tests {
        t.Run(tt.expr, func(t *testing.T) {
            _, err := cron.Parse(tt.expr)
            valid := err == nil
            if valid != tt.valid {
                t.Errorf("Parse(%q) valid=%v, want %v", tt.expr, valid, tt.valid)
            }
        })
    }
}
```

## Troubleshooting

### Cron Job Not Running

**Check**:
1. Scheduler started: `sch.Start()` called
2. Cron expression valid
3. Job not removed
4. No errors in logs
5. System time correct

### Job Running Multiple Times

**Cause**: Overlapping runs if job takes longer than interval

**Solution**: Add mutex or check for running instance

```go
var running atomic.Bool

sch.AddCron("job", "* * * * *", func(ctx context.Context) error {
    if !running.CompareAndSwap(false, true) {
        return nil // Already running
    }
    defer running.Store(false)

    return doWork(ctx)
})
```

### Missed Executions

**Cause**: System downtime or scheduler stopped

**Solution**: On startup, check for missed runs and handle accordingly

## Related Documentation

- [Scheduler Overview](README.md) — Scheduler module overview
- [Delayed Tasks](delayed-tasks.md) — Future task execution
- [Retry Policies](retry-policies.md) — Retry strategies
- [Examples](examples.md) — Complete examples

## Reference Implementation

See examples:
- `examples/scheduler/` — Cron job examples
- `examples/reference/` — Production cron jobs
