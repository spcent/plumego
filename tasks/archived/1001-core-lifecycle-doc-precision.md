# 1001 - core Lifecycle Doc Precision

State: done
Priority: P2
Primary Module: core

## Goal

Make core module lifecycle docs precisely match the current stable behavior for
prepare failures and `PreparationState`.

## Scope

- Clarify that non-destructive `Prepare` failure applies while the app is still
  mutable.
- Document `PreparationState()` empty return value for nil or zero-value apps.
- Keep runtime behavior unchanged.

## Non-goals

- Do not add public API.
- Do not change lifecycle transitions.
- Do not rewrite unrelated docs.

## Files

- `docs/modules/core/README.md`

## Tests

- `go run ./internal/checks/module-manifests`

## Docs Sync

Required in `docs/modules/core/README.md`.

## Done Definition

- Core docs no longer overstate prepare failure mutability.
- Core docs describe all public `PreparationState()` return categories,
  including empty state for nil/zero-value apps.

## Outcome

- Clarified that non-destructive prepare failure applies while the app is still
  mutable.
- Documented empty `PreparationState()` for nil and zero-value apps.
- Verified with `go run ./internal/checks/module-manifests`.
