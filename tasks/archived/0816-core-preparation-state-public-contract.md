# 0816 - core PreparationState Public Contract

State: done
Priority: P1
Primary Module: core

## Goal

Make `PreparationState` an explicit, documented stable core entrypoint because it
is exported and already appears in stable API output.

## Scope

- Add `PreparationState` to `core/module.yaml` public entrypoints.
- Align `docs/modules/core/README.md` public-entrypoint text with exported core config and lifecycle types.
- Keep behavior unchanged.

## Non-goals

- Do not add new lifecycle states.
- Do not rename or remove exported symbols.
- Do not widen core responsibilities beyond lifecycle state description.

## Files

- `core/module.yaml`
- `docs/modules/core/README.md`

## Tests

- `go run ./internal/checks/module-manifests`
- `go test -timeout 20s ./core/...`

## Docs Sync

Required in `docs/modules/core/README.md`.

## Done Definition

- The module manifest and module README list the same stable exported core
  lifecycle/config surfaces.
- Module manifest validation passes.

## Outcome

- Added `PreparationState` to the core module public-entrypoint manifest.
- Updated the core module README to list exported config/lifecycle types and
  describe the preparation-state surface.
- Verified with `go run ./internal/checks/module-manifests` and
  `go test -timeout 20s ./core/...`.
