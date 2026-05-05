# 0754 - x/cache redis production adapter contract

Status: done
Priority: P1
Primary module: `x/cache`

## Problem

Redis adapter stable readiness still depends on clear production guidance:
canonical construction, cache-miss mapping, optional capabilities, and
destructive `Clear` behavior need an explicit adapter contract and compatibility
matrix without adding a concrete Redis dependency.

## Scope

- Add a documented compatibility matrix for client miss mapping, atomic/append
  support, prefix clear, and `FlushDB`.
- Strengthen constructor documentation around
  `NewValidatedAdapterWithOptions` as the canonical constructor.
- Add tests for cache-miss mapping and optional capability failure behavior.
- Document production guidance for namespaced clear and why `FlushDB` is opt-in.
- Keep legacy constructors as compatibility entry points.

## Out of Scope

- Adding go-redis or another concrete Redis dependency.
- Removing exported compatibility fields.
- Promoting Redis adapter to stable.

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

Redis adapter production semantics and compatibility boundaries are explicit in
code tests and docs, while preserving the existing dependency-free adapter.

## Outcome

- Documented the dependency-free Redis client adapter contract in package and
  constructor comments.
- Added a cache-miss mapping regression test showing raw client miss errors
  remain raw unless `WithNotFound` is configured.
- Documented the compatibility matrix for miss mapping, basic operations,
  optional atomic/append support, namespaced clear, and opt-in `FlushDB`.
- Added production guidance for `NewValidatedAdapterWithOptions`,
  `WithClearPrefix`, and limited `FlushDB` use.

## Validation Run

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
