# Card 1409

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
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

