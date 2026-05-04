# 0739 - x/cache redis stable adapter contract

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

`x/cache/redis.Adapter` still exposes mutable fields as the primary behavior
surface and returns inconsistent key validation errors. Stable use needs a
clearer adapter contract without breaking compatibility.

## Scope

- Add immutable internal option state used by adapter operations.
- Keep exported fields for compatibility while preferring constructor options.
- Normalize Redis key validation errors to wrap stable cache errors.
- Add tests for post-construction field mutation isolation and error wrapping.

## Out of Scope

- Removing exported compatibility fields.
- Adding a concrete Redis dependency.
- Tenant-specific key scoping.

## Files

- `x/cache/redis/redis.go`
- `x/cache/redis/redis_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

New Redis adapter call sites get stable constructor-owned behavior, key errors
wrap stable cache errors, and old mutable fields remain compatibility-only.
