# 0888 - x/cache distributed hashring collision weight

Status: done
Priority: P1
State: done
Primary module: `x/cache`

## Problem

`x/cache/distributed` exposes node weights and hash-collision metrics, but the
hash ring does not use weights and hash collisions overwrite ring entries.
Stable behavior needs deterministic collision handling and a meaningful weight
contract.

## Scope

- Preserve the existing `HashRing` entrypoint.
- Use `CacheNode.Weight()` to scale virtual-node placement.
- Avoid silent virtual-node overwrite on hash collision.
- Expose collision counts through distributed metrics.
- Add focused tests for custom colliding hash functions and weighted placement.

## Out of Scope

- Rebalancing existing cached data.
- Changing distributed cache replication semantics.
- Adding non-stdlib dependencies.

## Files

- `x/cache/distributed/hashring.go`
- `x/cache/distributed/distributed.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go vet ./x/cache/distributed`
- `go run ./internal/checks/dependency-rules`

## Done Definition

Hash collisions no longer overwrite ring entries, node weights affect virtual
node placement, metrics expose collision counts, and behavior is test-covered.

## Outcome

- Recorded actual virtual-node hashes per node so removals remain correct.
- Resolved virtual-node collisions with deterministic probing instead of
  overwriting ring entries.
- Applied `CacheNode.Weight()` to virtual-node placement.
- Exposed collision counts through `DistributedMetrics.HashCollisions`.
- Documented weight and collision behavior in the `x/cache` primer.

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go vet ./x/cache/distributed`
- `go run ./internal/checks/dependency-rules`
