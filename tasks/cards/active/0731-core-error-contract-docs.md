# 0731 - core Error Contract Docs

State: active
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

