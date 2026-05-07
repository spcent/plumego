# 0762 - x/cache distributed config and health bounds

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

Hash ring and health checker constructors mutate caller config objects while
leaderboard uses constructor-local normalization. Hash ring virtual-node
placement also multiplies `VirtualNodes * Weight()` without a practical upper
bound, and the default health probe still needs a stable production contract.

## Scope

- Normalize hash ring and health checker configs without mutating caller-owned
  config values.
- Add practical validation/bounds for virtual-node weight expansion.
- Add focused tests for non-mutating constructors and bounded weight expansion.
- Document default health probe semantics and production guidance.
- Sync x/cache evidence.

## Out of Scope

- Replacing consistent hashing.
- Adding tenant-aware health behavior.
- Changing distributed cache replication semantics.

## Files

- `x/cache/distributed/hashring.go`
- `x/cache/distributed/node.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Distributed construction avoids caller config mutation, hash ring placement has
a defensible bound, and default health probing is documented as a lightweight
fallback rather than a universal production probe.

## Outcome

- Made `NewConsistentHashRing` normalize config on an internal copy without
  mutating caller-owned config.
- Made `NewHealthChecker` normalize config on an internal copy without mutating
  caller-owned config.
- Added `ErrHashRingNodeTooLarge` for pathological weighted virtual-node
  expansion before allocation.
- Added focused tests for config immutability and excessive weighted placement.
- Documented the default health probe as a lightweight fallback and kept
  production guidance explicit.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
