# 1006 - x/cache distributed replication mode semantics

Status: done
Priority: P0
Primary module: `x/cache`

## Problem

`ReplicationNone` still asks the hash ring for the full configured
replication factor before writing, so a single-copy write can fail with
`ErrInsufficientReplicas`. `Exists` and `Delete` also do not clearly align with
the replication mode and failover contract.

## Scope

- Make `ReplicationNone` select only the primary node for writes.
- Keep replicated modes fail-closed when the configured replica count cannot be
  satisfied.
- Align and document `Exists` and `Delete` semantics with the selected
  replication/failover model.
- Add focused regression tests for none, sync, async, delete, and exists paths.

## Out of Scope

- Durable repair queues.
- Async worker pools.
- Stable `store/cache` API changes.

## Files

- `x/cache/distributed/distributed.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Distributed replication mode selection is explicit, `ReplicationNone` no
longer depends on secondary replicas, and the behavior is documented and tested.

## Outcome

- Added shared replication-node selection so `ReplicationNone` uses only the
  primary node while replicated modes continue to require the configured
  replica count.
- Applied that selection to `Set`, `Delete`, `Incr`, `Decr`, `Append`, and
  `FailoverNextNode`.
- Made `Exists` use the configured failover strategy when the primary is
  unhealthy or returns an infrastructure error.
- Documented primary-only non-replication behavior and `Exists` failover
  semantics.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
