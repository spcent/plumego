# 0758 - x/cache redis compatibility field concurrency contract

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

`redis.Adapter` keeps exported mutable fields for compatibility, while the
option path freezes behavior internally. Stable readiness needs the concurrency
contract for those compatibility fields to be explicit and tested where the
canonical option path is expected to be immutable.

## Scope

- Document exported adapter fields as compatibility-only configuration fields
  that must not be mutated concurrently with cache operations.
- Strengthen canonical constructor documentation.
- Add tests that option-owned behavior remains stable after field mutation for
  cache miss, max key length, clear prefix, and FlushDB policy.
- Update module docs and evidence.

## Out of Scope

- Removing exported fields.
- Adding a concrete Redis dependency.
- Implementing locking around compatibility field mutation.

## Files

- `x/cache/redis/redis.go`
- `x/cache/redis/redis_test.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`

## Validation

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Redis adapter compatibility-field concurrency boundaries and canonical option
immutability are explicit, tested, and documented.
