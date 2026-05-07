# 0746 - x/cache distributed async operational guardrails

Status: done
Priority: P0
Primary module: `x/cache`

## Problem

Async replication is bounded by timeout but still starts one goroutine per
secondary write. Stable readiness needs an explicit concurrency/backpressure
guard instead of unbounded background work.

## Scope

- Add a configurable async replication concurrency limit.
- Fail closed at schedule time when the async replication limit is exhausted by
  recording replication failure metrics and dropping that secondary write.
- Keep current best-effort caller semantics for async primary writes.
- Document the limit and remaining repair/callback blocker.
- Add regression tests for exhausted async scheduling.

## Out of Scope

- Durable repair queues.
- Changing async `Set`/`Incr`/`Append` return values.
- Promoting distributed cache stability status.

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

Async replication has a bounded scheduling path, exposed configuration, tests,
and documentation.

## Outcome

- Added `Config.AsyncReplicationMaxConcurrency` with a default bounded async
  replication limiter.
- Made exhausted async scheduling record `ReplicationFailures` and drop the
  secondary write without changing primary-write caller semantics.
- Reused the scheduling path for async `Set`, `Incr`, and `Append` secondary
  replication.
- Documented async timeout/concurrency bounds and the remaining repair/callback
  blocker.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
