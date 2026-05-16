# 1324 - x/cache redis option model and stable surface

Status: done
Priority: P1
State: done
Primary module: `x/cache`

## Problem

Redis adapter options are `func(*Adapter)`, so external options can mutate
compatibility fields without setting internal validation flags. The adapter also
retains compatibility constructors and public mutable fields, which should not
be treated as the stable surface.

## Scope

- Add tests documenting that only package-provided options participate in
  validated constructor semantics.
- Clarify docs/comments that external `Option` functions are compatibility
  hooks and not part of the future stable constructor contract.
- Keep `NewValidatedAdapterWithOptions` as the canonical production path.
- Document capability-gated behavior as partial adapter behavior until a real
  client matrix is recorded.
- Sync x/cache evidence.

## Out of Scope

- Removing public compatibility fields.
- Adding a concrete Redis driver dependency.
- Running an external Redis server.

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

Redis adapter docs and tests clearly separate compatibility hooks from the
future stable constructor contract, without introducing new dependencies.

## Outcome

- Documented package-provided `With*` options as the future stable frozen
  constructor contract.
- Validated effective fields written by custom compatibility options in
  `NewValidatedAdapterWithOptions`.
- Added tests for invalid custom option values and for custom options remaining
  mutable compatibility-field hooks rather than frozen internal options.
- Validation passed:
  - `go test -race -timeout 60s ./x/cache/redis`
  - `go test -timeout 20s ./x/cache/...`
  - `go vet ./x/cache/...`
