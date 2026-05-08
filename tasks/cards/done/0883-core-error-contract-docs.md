# 0883 - core Error Contract Docs

State: done
Priority: P2
Primary Module: core

## Goal

Clarify the stable core lifecycle error contract without expanding the public
API surface with new sentinel errors.

## Scope

- Document that core lifecycle errors are returned as Go errors with canonical
  operation-prefixed messages and wrapped causes where available.
- Document that callers should handle errors directly and must not depend on new
  exported core sentinel errors.
- Keep runtime behavior unchanged.

## Non-goals

- Do not add exported error sentinel variables.
- Do not rewrite existing error helpers.
- Do not change error strings.

## Files

- `docs/modules/core/README.md`

## Tests

- `go run ./internal/checks/module-manifests`

## Docs Sync

Required in `docs/modules/core/README.md`.

## Done Definition

- Core module docs describe the lifecycle error contract and its intentional
  small-surface boundary.

## Outcome

- Documented core operation-prefixed lifecycle/wiring errors and wrapped-cause
  behavior.
- Documented that core does not export lifecycle sentinel errors, keeping the
  stable API surface small.
- Verified with `go run ./internal/checks/module-manifests`.
