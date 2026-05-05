# 0757 - x/cache distributed callback and config clarity

Status: done
Priority: P1
Primary module: `x/cache`

## Problem

The async replication drop callback runs on the caller scheduling path and zero
values in several config fields mean “use default”. Stable readiness needs these
operational constraints to be clear, and callback panics should not break cache
operations.

## Scope

- Recover from panics in `AsyncReplicationDropHandler`.
- Record replication failure metrics when the drop handler panics.
- Document that the callback is synchronous and must not block.
- Document config zero-value default semantics for retry, async worker, queue,
  and timeout fields.
- Add focused regression coverage for callback panic isolation.

## Out of Scope

- Asynchronous callback delivery.
- Introducing a separate disable knob for retry or queueing.
- Durable repair.

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

Drop callback failure cannot panic through cache operations, and config default
semantics are explicit in tests/docs.

## Outcome

- Recovered panics from `Config.AsyncReplicationDropHandler`.
- Counted drop handler panics as additional replication failures.
- Documented synchronous drop handler execution and zero-value default semantics
  for retry and async replication config fields.
- Added regression coverage for drop handler panic isolation.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
