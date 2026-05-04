# 0736 - core PreparationState Reader

State: active
Priority: P1
Primary Module: core

## Goal

Make the exported `PreparationState` stable surface practically usable through a
read-only app state accessor.

## Scope

- Add a thread-safe `(*App).PreparationState() PreparationState` method.
- Document and snapshot the new stable method.
- Add focused tests for nil, zero-value, mutable, handler-prepared, and
  server-prepared states.

## Non-goals

- Do not add new preparation states.
- Do not expose mutable internals.
- Do not change lifecycle transitions.

## Files

- `core/app.go`
- `core/app_test.go`
- `core/module.yaml`
- `docs/modules/core/README.md`
- `docs/stable-api/snapshots/core-head.snapshot`

## Tests

- `go test -timeout 20s ./core/...`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/extension-api-snapshot -compare docs/stable-api/snapshots/core-head.snapshot /private/tmp/plumego-core-current.snapshot`

## Docs Sync

Required in `docs/modules/core/README.md` and stable API snapshot.

## Done Definition

- Callers can read app preparation state without touching internals.
- Manifest, docs, and stable API snapshot include the new method.

