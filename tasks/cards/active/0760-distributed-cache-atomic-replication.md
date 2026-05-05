# Card 0760

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P2
State: active
Primary Module: x/cache/distributed
Owned Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go
Depends On:

Goal:
Align distributed Incr/Decr/Append with the configured replication semantics.

Scope:
- Ensure atomic mutation capabilities do not bypass replicationFactor and replicationMode silently.
- Prefer fail-closed behavior when a replicated atomic mutation cannot be safely applied.
- Add tests for replicated counter/appender behavior or explicit unsupported behavior.

Non-goals:
- Do not implement cross-node distributed transactions.
- Do not change the stable cache capability interfaces.

Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go

Tests:
- go test ./x/cache/distributed

Docs Sync:
- Update docs only if the chosen behavior is an explicit unsupported mode.

Done Definition:
- Atomic capability behavior is consistent with distributed replication configuration and covered by tests.
- Distributed cache tests pass.

Outcome:

