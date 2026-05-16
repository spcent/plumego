# Card 1408

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/pubsub
Owned Files:
- x/pubsub/replay.go
- x/pubsub/replay_store.go
- x/pubsub/replay_filters.go
- x/pubsub/replay_test.go
Depends On:
- 1407

Goal:
- Reduce `x/pubsub` replay edit radius by separating store and filtering helpers.

Scope:
- Move replay persistence/store helpers into `replay_store.go`.
- Move replay filtering and selection helpers into `replay_filters.go`.
- Preserve replay ordering, history semantics, and error values.

Non-goals:
- Do not change broker publish/subscribe behavior.
- Do not change DLQ, ordering, or consumer group behavior in this card.
- Do not promote `x/pubsub` maturity.

Files:
- x/pubsub/replay.go
- x/pubsub/replay_store.go
- x/pubsub/replay_filters.go
- x/pubsub/replay_test.go

Tests:
- go test -timeout 30s ./x/pubsub
- go vet ./x/pubsub
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless replay public comments change.

Done Definition:
- Replay storage and filter code are split from the main replay flow.
- Existing replay tests pass.
- No exported symbol or behavior change is introduced.

Outcome:
- Completed on May 15, 2026.
- Split replay in-memory storage, oldest-message eviction, archive worker, archive write, and archive load helpers into `x/pubsub/replay_store.go`.
- Split replay query, filter, sorting, and ID selection helpers into `x/pubsub/replay_filters.go`.
- Kept `x/pubsub/replay.go` focused on replay config/types, construction, capture loop, replay flow, stats, and close lifecycle.
- Preserved replay ordering, history semantics, and error values.
- Validation passed:
  - `go test -timeout 30s ./x/pubsub`
  - `go vet ./x/pubsub`
  - `go run ./internal/checks/dependency-rules`
