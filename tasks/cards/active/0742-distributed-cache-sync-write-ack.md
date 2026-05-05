# Card 0742

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P1
State: active
Primary Module: x/cache/distributed
Owned Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go
Depends On:

Goal:
Ensure synchronous distributed cache writes fail when no healthy replica accepts the write.

Scope:
- Count attempted and successful synchronous replica writes.
- Return an existing no-node/unhealthy error when every candidate is skipped or fails.
- Add regression coverage for all-unhealthy sync replication.

Non-goals:
- Do not redesign consistent hashing or node health checks.
- Do not add quorum configuration in this card.

Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go

Tests:
- go test -timeout 20s ./x/cache/distributed

Docs Sync:
- None unless distributed cache behavior docs exist and mention write guarantees.

Done Definition:
- Set returns an error when synchronous replication writes to zero healthy nodes.
- Existing success paths still pass.
- Targeted tests pass.

Outcome:

