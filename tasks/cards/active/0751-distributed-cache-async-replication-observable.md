# Card 0751

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P2
State: active
Primary Module: x/cache/distributed
Owned Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go
Depends On:
- tasks/cards/active/0750-distributed-cache-delete-ack.md

Goal:
Make async replication honor caller cancellation and expose failed replica writes through metrics.

Scope:
- Avoid replacing caller context with context.Background for async replica writes.
- Track async replication failures in DistributedMetrics.
- Add regression coverage for failed async replica writes.

Non-goals:
- Do not make async replication synchronous.
- Do not add logging or external observability dependencies.

Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go

Tests:
- go test -timeout 20s ./x/cache/distributed

Docs Sync:
- None unless distributed metrics docs exist.

Done Definition:
- Async replica writes use caller context.
- Failed async replica writes increment an observable metric.
- Targeted tests pass.

Outcome:

