# 1139 - x/cache distributed async queue contract

Status: done
Priority: P0
Primary module: `x/cache`

## Problem

Async replication is bounded by concurrent work but still has no explicit queue,
drop callback, or repair handoff. Stable readiness needs an operational contract
for queued secondary writes and dropped async replication attempts.

## Scope

- Add an internal async replication worker queue with configurable queue limit.
- Keep the caller contract unchanged: async primary writes succeed or fail based
  on the primary write only.
- Drop queued work fail-closed when the queue is full, record metrics, and invoke
  an optional drop callback.
- Add clean shutdown behavior for the async worker path.
- Add tests for queue exhaustion, drop callback, and close behavior.
- Document that dropped async work is observable but not repaired by this
  package.

## Out of Scope

- Durable retry/repair storage.
- Changing `store/cache.Cache`.
- Promoting distributed cache to stable.

## Files

- `x/cache/distributed/distributed.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Async replication has a bounded queue, explicit drop behavior, callback
observability, close semantics, tests, and documentation.

## Outcome

- Replaced per-write async goroutine scheduling with a bounded worker queue.
- Added `Config.AsyncReplicationQueueSize` and
  `Config.AsyncReplicationDropHandler`.
- Added `AsyncReplicationDrop`, `AsyncReplicationDropReason`,
  `AsyncReplicationDropQueueFull`, and `AsyncReplicationDropClosed` for
  observable dropped async work.
- Made `Close` stop async workers and report post-close async schedule attempts
  through the drop path.
- Documented queue limits, drop callback behavior, and the remaining no-durable
  repair contract.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
