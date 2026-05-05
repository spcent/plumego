# 0753 - Core Manifest Import Boundary Tightening

State: done
Priority: P1
Primary module: core

## Goal

Align `core/module.yaml` with the current core ownership boundary by removing stale `health` and `metrics` import allowances.

## Scope

- Remove `health` and `metrics` from `core/module.yaml` `allowed_imports`.
- Validate that current core code does not rely on those imports.
- Keep core focused on app kernel, constructor dependencies, route entrypoints, middleware entrypoints, and HTTP lifecycle.

## Non-goals

- Do not move health or metrics code.
- Do not change public APIs.
- Do not add new dependency rules outside the core manifest.

## Files

- `core/module.yaml`

## Tests

- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`
- `go test -timeout 20s ./core/...`

## Docs Sync

No docs update required; `docs/modules/core/README.md` already states that metrics are attached through middleware/app-local wiring.

## Done Definition

- `core/module.yaml` no longer allows direct `health` or `metrics` imports.
- Boundary and manifest checks pass.
- Core tests pass.

## Validation

- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`
- `go test -timeout 20s ./core/...`
