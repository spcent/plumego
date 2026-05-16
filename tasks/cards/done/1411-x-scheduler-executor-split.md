# Card 1411

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/scheduler
Owned Files:
- x/scheduler/scheduler_executor.go
- x/scheduler/executor_runner.go
- x/scheduler/executor_retry.go
- x/scheduler/scheduler_test.go
Depends On:
- 1410

Goal:
- Split `x/scheduler` executor runtime and retry/callback compatibility code.

Scope:
- Move executor run-loop helpers into `executor_runner.go`.
- Move retry and legacy callback compatibility helpers into `executor_retry.go`.
- Preserve job execution, delayed schedule, retry, DLQ, and legacy callback behavior.

Non-goals:
- Do not change cron parsing.
- Do not alter admin HTTP behavior.
- Do not remove legacy callback compatibility in this card.

Files:
- x/scheduler/scheduler_executor.go
- x/scheduler/executor_runner.go
- x/scheduler/executor_retry.go
- x/scheduler/scheduler_test.go

Tests:
- go test -timeout 30s ./x/scheduler
- go vet ./x/scheduler
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless executor public comments change.

Done Definition:
- Executor runtime and retry helpers have separate file ownership.
- Existing scheduler tests pass.
- No public API or scheduling behavior change is introduced.

Outcome:
- Completed on May 15, 2026.
- Split scheduler run loop, due-item dispatch, backpressure handling, worker loop, and execution flow into `x/scheduler/executor_runner.go`.
- Split retry handling, retry-exhaustion behavior, legacy dead-letter callback path, and DLQ entry construction into `x/scheduler/executor_retry.go`.
- Kept `x/scheduler/scheduler_executor.go` focused on persistence, dependency, state, cancellation, and resume helpers.
- Preserved job execution, delayed scheduling, retry, DLQ, and legacy callback behavior.
- Validation passed:
  - `go test -timeout 30s ./x/scheduler`
  - `go vet ./x/scheduler`
  - `go run ./internal/checks/dependency-rules`
