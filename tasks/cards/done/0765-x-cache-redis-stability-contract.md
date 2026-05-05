# 0765 - x/cache redis stability contract

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

The Redis adapter is dependency-free and well-covered with stubs, but it is not
yet a proven production Redis adapter. Stable readiness needs a sharper
constructor/capability contract and an explicit real-driver evidence gap.

## Scope

- Keep `NewValidatedAdapterWithOptions` as the canonical constructor in docs.
- Add tests that compatibility field mutation after option construction cannot
  override frozen option behavior for miss mapping and clear policy.
- Clarify that `Incr`, `Append`, and `Clear` are capability-gated behavior.
- Record the real-driver integration matrix still required before promotion.
- Sync x/cache evidence.

## Out of Scope

- Adding a concrete Redis client dependency to the main module.
- Running an external Redis server.
- Removing compatibility fields in this pass.

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

Redis adapter docs and tests make option-owned behavior, optional capabilities,
and the remaining real-driver evidence gap explicit.

## Outcome

- Added validated-constructor coverage proving later compatibility-field
  mutation cannot override the option-owned clear policy.
- Kept cache-miss mapper freezing covered through the validated constructor.
- Clarified in docs that option values win over later compatibility-field
  mutation.
- Recorded the real-driver matrix that remains required before Redis adapter
  promotion.
- Kept the adapter dependency-free.

## Validation Run

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
