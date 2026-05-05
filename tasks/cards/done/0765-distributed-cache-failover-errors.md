# Card 0765

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/cache
Owned Files: x/cache/distributed/distributed.go, x/cache/distributed/distributed_test.go
Depends On:

Goal:

Make distributed cache failover preserve backend errors instead of converting every replica failure into cache.ErrNotFound.

Scope:

- Track the first non-ErrNotFound replica Get error during failover.
- Return that backend error when no healthy replica returns a value.
- Keep ErrNotFound only for real misses across attempted replicas.
- Add focused regression coverage.

Non-goals:

- Changing the hash ring or replica selection policy.
- Adding metrics for failover errors.
- Changing Set/Delete replication behavior.

Files:

- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go

Tests:

- go test -race -timeout 60s ./x/cache/distributed
- go test -timeout 20s ./x/cache/distributed
- go vet ./x/cache/distributed

Docs Sync:

- Not required; this aligns implementation with cache error semantics.

Done Definition:

- Non-miss replica Get errors survive failover.
- True replica misses still return cache.ErrNotFound.
- Package tests and vet pass.

Outcome:

- Distributed failover now returns the first non-miss replica backend error when no replica can serve the value.
- Added regression coverage for preserving replica Get errors.
- Validation passed:
  - go test -race -timeout 60s ./x/cache/distributed
  - go test -timeout 20s ./x/cache/distributed
  - go vet ./x/cache/distributed
