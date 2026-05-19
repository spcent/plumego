# Card 1409

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/pubsub
Owned Files:
- x/pubsub/dlq.go
- x/pubsub/dlq_policy.go
- x/pubsub/ordering.go
- x/pubsub/ordering_state.go
- x/pubsub/dlq_test.go
Depends On:
- 1408

Goal:
- Separate `x/pubsub` DLQ policy and ordering state helpers from larger implementation files.

Scope:
- Move DLQ retry/dead-letter policy helpers into `dlq_policy.go`.
- Move ordering state bookkeeping into `ordering_state.go`.
- Preserve DLQ enqueue/replay behavior and ordering guarantees.

Non-goals:
- Do not change broker APIs.
- Do not change replay or persistence behavior in this card.
- Do not alter metrics or Prometheus output.

Files:
- x/pubsub/dlq.go
- x/pubsub/dlq_policy.go
- x/pubsub/ordering.go
- x/pubsub/ordering_state.go
- x/pubsub/dlq_test.go

Tests:
- go test -timeout 30s ./x/pubsub
- go vet ./x/pubsub
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless public DLQ or ordering comments move.

Done Definition:
- DLQ policy and ordering state have clear file ownership.
- Existing DLQ and ordering tests pass.
- No public API or behavior change is introduced.

Outcome:
- Completed on May 15, 2026.
- Split DLQ retry, retry-delay, archive, eviction, and alert policy helpers into `x/pubsub/dlq_policy.go`.
- Split ordered queue creation, queue lookup, sequence verification, ordering stats, and missing-sequence state helpers into `x/pubsub/ordering_state.go`.
- Preserved DLQ enqueue/replay behavior, retry semantics, and ordering guarantees.
- Validation passed:
  - `go test -timeout 30s ./x/pubsub`
  - `go vet ./x/pubsub`
  - `go run ./internal/checks/dependency-rules`
