# 0736 - x/cache distributed fail closed contract

Status: active
Priority: P0
Primary module: `x/cache`

## Problem

`x/cache/distributed` can still accept nodes with nil cache instances, silently
write fewer replicas than configured, and report successful `Clear` when no
healthy node was cleared.

## Scope

- Reject nil cache nodes during construction and `AddNode`.
- Make configured replication underfill observable as an error.
- Make `Clear` fail when all selected nodes are unhealthy or when any selected
  node fails to clear.
- Add regression tests for nil cache, underfilled replication, and clear
  failure semantics.

## Out of Scope

- Async worker queues.
- Durable repair queues.
- Changing stable `store/cache`.

## Files

- `x/cache/distributed/node.go`
- `x/cache/distributed/hashring.go`
- `x/cache/distributed/distributed.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Distributed cache construction and destructive operations fail closed for nil
nodes, missing replicas, and no-clear situations.
