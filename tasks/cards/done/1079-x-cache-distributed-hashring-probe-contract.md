# 1079 - x/cache distributed hashring probe contract

Status: done
Priority: P0
Primary module: `x/cache`

## Problem

Virtual-node collision resolution probes by incrementing the uint32 hash with no
explicit upper bound. Stable readiness needs a deterministic failure path for
pathological hash functions or saturated probe windows.

## Scope

- Add a bounded virtual-node collision probe limit.
- Return a sentinel error when virtual-node placement cannot be resolved.
- Add regression coverage for a pathological hash ring placement failure.
- Document the collision placement failure behavior.

## Out of Scope

- Replacing the consistent-hash algorithm.
- Changing normal hash distribution behavior.

## Files

- `x/cache/distributed/hashring.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Hash-ring collision resolution has a deterministic error path and test coverage
for pathological placement failure.

## Outcome

- Added `ErrHashRingSaturated` for bounded virtual-node placement failure.
- Bounded collision probing with `maxVirtualNodeHashProbes`.
- Rolled back virtual-node entries when node placement fails mid-add.
- Added regression coverage for pathological constant-hash placement failure.
- Documented saturated hash-ring placement behavior.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
