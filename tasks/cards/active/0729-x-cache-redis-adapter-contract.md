# 0729 - x/cache redis adapter contract

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

`x/cache/redis` claims atomic increment behavior and TTL preservation but its
minimal client interface only supports get/set/delete/exists. The adapter also
exposes mutable fields and `Clear` can flush an entire DB.

## Scope

- Make increment/decrement and append semantics honest and test-covered.
- Add optional interfaces for Redis-native atomic operations where available.
- Define TTL preservation behavior explicitly instead of relying on `Set(..., 0)`.
- Harden or clearly constrain `Clear` behavior.
- Add tests for atomic-interface dispatch, unsupported fallbacks, and key
  validation consistency.

## Out of Scope

- Adding a concrete third-party Redis dependency.
- Tenant-aware cache scoping.
- Stable root interface changes.

## Files

- `x/cache/redis/redis.go`
- `x/cache/redis/redis_test.go`
- `docs/modules/x-cache/README.md` if public adapter semantics change

## Validation

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Redis adapter documentation, comments, tests, and code agree on atomicity, TTL,
and clear semantics without introducing non-stdlib dependencies.
