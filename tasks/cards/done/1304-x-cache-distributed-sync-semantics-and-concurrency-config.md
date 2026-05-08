# 1304 - x/cache distributed sync semantics and concurrency config

Status: done
Priority: P1
Primary module: `x/cache`

## Problem

`ReplicationSync` mutation ordering is split: `Set` writes replicas concurrently
while `Incr` and `Append` mutate the primary first and then secondaries. The
bounded sync `Set` fanout also reuses `AsyncReplicationMaxConcurrency`, which
makes the exported config surface unclear.

## Scope

- Make synchronous `Set` use primary-first ordering, then bounded secondary
  writes.
- Add a sync replication concurrency config field with clear naming.
- Preserve existing `AsyncReplicationMaxConcurrency` behavior for async workers.
- Keep compatibility fallback for callers that have not set the new sync field.
- Add tests for primary-first `Set` behavior and sync fanout config selection.
- Sync x/cache docs and evidence.

## Out of Scope

- Strong consistency or rollback.
- Changing `Incr` and `Append` partial-write behavior.
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

Synchronous mutation ordering is documented and uniform for primary-first
operations, and sync fanout no longer depends on an async-named config field.

## Outcome

- Added `Config.SyncReplicationMaxConcurrency` with zero-value fallback to the
  normalized async concurrency for compatibility.
- Made synchronous `Set` write primary first before bounded secondary writes.
- Added tests for explicit sync fanout selection, compatibility fallback,
  primary-first `Set`, and negative sync concurrency validation.
- Validation passed:
  - `go test -race -timeout 60s ./x/cache/distributed`
  - `go test -timeout 20s ./x/cache/...`
  - `go vet ./x/cache/...`
