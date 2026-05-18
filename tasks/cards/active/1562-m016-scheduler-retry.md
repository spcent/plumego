# Card 1562

Milestone: M-016
Recipe: specs/change-recipes/add-package.yaml
Priority: P2
State: active
Primary Module: reference/with-events
Owned Files:
- `reference/with-events/internal/scheduler/retry_job.go`
- `reference/with-events/internal/scheduler/retry_job_test.go`

Goal:
- Implement a delayed-retry job in reference/with-events using
  x/messaging/scheduler that re-enqueues failed order events for reprocessing
  after a configurable backoff delay.

Scope:
- Create internal/scheduler/retry_job.go:
  - RetryJob struct holding a scheduler.Scheduler and an OrderPublisher.
  - Register(ctx context.Context) — registers a named job "order.retry" with
    the scheduler; job payload is a serialised OrderCreated event.
  - Schedule(ctx context.Context, event OrderCreated, delay time.Duration) error
    — enqueues a delayed retry via scheduler.Enqueue.
  - The job handler republishes the event via OrderPublisher.Publish; on failure
    increments a retry counter and schedules another retry up to MaxRetries (3).
- Create internal/scheduler/retry_job_test.go covering:
  - Schedule enqueues a job with correct delay.
  - Job handler republishes the event on first attempt.
  - Job handler stops retrying after MaxRetries exceeded.
  - Job handler does not retry on idempotency-store hit (event already processed).

Non-goals:
- Do not use a real distributed scheduler; use the in-process x/messaging/scheduler.
- Do not persist retry state beyond an in-memory counter.
- Do not add cron scheduling; only delayed one-off retries.

Files:
- `reference/with-events/internal/scheduler/retry_job.go`
- `reference/with-events/internal/scheduler/retry_job_test.go`

Tests:
- `go test -timeout 30s ./reference/with-events/internal/scheduler/...`
- `go vet ./reference/with-events/...`
- `go build ./reference/with-events/...`

Docs Sync:
- none at this card; covered in M-016 README.

Done Definition:
- RetryJob.Schedule enqueues a delayed job.
- Job handler republishes; stops at MaxRetries.
- All four retry_job_test.go cases pass.
- `go build ./reference/with-events/...` exits 0.

Outcome:
-
