# Card 0760

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P2
State: done
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
- Updated docs/modules/x-cache/README.md to document that replicated atomic
  mutations fail closed until a cross-node atomic strategy exists.

Done Definition:
- Atomic capability behavior is consistent with distributed replication configuration and covered by tests.
- Distributed cache tests pass.

Outcome:
- Added atomicMutationNode to centralize primary-node selection for counter and
  appender operations.
- Made Incr/Decr/Append return cache.ErrCapabilityUnsupported when a replicated
  configuration would otherwise bypass replicas.
- Preserved atomic mutations for ReplicationNone or single-replica operation.
- Added regression tests for replicated-mode rejection and ReplicationNone.
- Validated with:
  - go test -timeout 20s ./x/cache/...
  - go test -race -timeout 60s ./x/cache/...
  - go vet ./x/cache/...
