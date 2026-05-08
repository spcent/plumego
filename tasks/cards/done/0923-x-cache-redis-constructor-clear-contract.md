# 0923 - x/cache redis constructor clear contract

Status: done
Priority: P1
Primary module: `x/cache`

## Problem

`x/cache/redis.Adapter` still exposes mutable fields as the primary
configuration surface, and `Clear` remains a DB-wide operation when enabled.
Stable use needs an option-based construction path and a safer clear contract.

## Scope

- Add option-based Redis adapter construction.
- Keep the existing constructor compatible.
- Add a prefix-clear capability for clients that can delete namespaced keys.
- Prefer prefix clear over `FlushDB` when configured.
- Add tests for constructor options and clear selection.

## Out of Scope

- Adding a concrete Redis dependency.
- Tenant-specific key scoping.
- Removing compatibility fields in this pass.

## Files

- `x/cache/redis/redis.go`
- `x/cache/redis/redis_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Redis adapter has a stable option-based construction path, clear behavior can be
namespaced, and DB-wide flush remains explicitly opt-in.

## Outcome

- Added `NewAdapterWithOptions` and option helpers while keeping
  `NewAdapter` compatible.
- Added `PrefixFlusher` and `ClearPrefix` support for namespaced clear.
- Kept DB-wide `FlushDB` disabled by default and avoided fallback flushing when
  prefix clear is configured but unsupported.
- Documented Redis construction and clear behavior.

## Validation Run

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
