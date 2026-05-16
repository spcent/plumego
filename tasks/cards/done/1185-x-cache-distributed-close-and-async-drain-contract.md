# 1185 - x/cache distributed close and async drain contract

Status: done
Priority: P0
State: done
Primary module: `x/cache`

## Problem

`DistributedCache.Close` stops background workers, but foreground operations do
not have a closed-cache contract and queued async replication work can be
discarded without an explicit drop callback. Stable readiness needs lifecycle
semantics that are deterministic and testable.

## Scope

- Add a package-level closed-cache error for distributed cache operations.
- Make foreground cache operations and topology mutations fail after `Close`.
- Keep snapshot-style inspection methods safe after close where appropriate.
- Drain queued async replication jobs on close through the existing drop path.
- Add tests for post-close operation errors and close-time async drop callbacks.
- Sync module docs and evidence.

## Out of Scope

- Durable async repair.
- Changing best-effort replication semantics before close.
- Promoting `x/cache` to stable.

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

Distributed cache lifecycle behavior after `Close` and close-time async queue
drop behavior are explicit, tested, and documented.

## Outcome

- Added `distributed.ErrClosed`.
- Made foreground cache operations, topology mutation, and `NodeHealth` return
  `ErrClosed` after `Close`.
- Kept `Nodes` and `GetMetrics` available as snapshot-style inspection methods.
- Made `Close` drain queued async replication jobs through the closed-drop
  callback path.
- Documented post-close behavior and close-time queued async drops.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
