# 0747 - core Use Empty Contract

State: done
Priority: P2
Primary Module: core

## Goal

Make `App.Use()` with no middleware an explicit no-op contract.

## Scope

- Add a focused test proving `Use()` with zero arguments returns nil, does not
  mutate the middleware chain, and leaves route/server preparation usable.
- Document the no-op behavior in the core frozen behavior matrix.

## Non-goals

- Do not change middleware ordering.
- Do not change nil middleware rejection.
- Do not add new middleware APIs.

## Files

- `core/app_test.go`
- `docs/modules/core/README.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

## Docs Sync

Required in the core module primer only.

## Done Definition

- `Use()` empty input behavior has a regression test.
- Core docs list empty `Use()` as a stable no-op.
- Core tests and vet pass.

## Outcome

- Added a regression test proving empty `Use()` returns nil, leaves the
  middleware chain and mutable state unchanged, and does not block preparation.
- Documented empty `Use()` as a stable no-op in the core frozen behavior matrix.
- Verified with `go test -timeout 20s ./core/...` and `go vet ./core/...`.
