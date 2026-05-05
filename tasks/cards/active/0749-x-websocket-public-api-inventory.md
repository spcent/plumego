# 0749 - x/websocket public API inventory

Status: active
Priority: P1
Primary module: `x/websocket`

## Goal

Make the exported websocket surface explicit enough for freeze review without
removing symbols outside a dedicated breaking-change pass.

## Scope

- Compare current exported symbols against `x/websocket/module.yaml`.
- Add or refresh an inventory document that classifies symbols as keep,
  review-before-stable, or release-governance blocker.
- Update evidence to point at the inventory.
- Keep `module.yaml` status as `experimental`.

## Non-goals

- Removing exported symbols.
- Claiming stable API freeze.
- Creating release refs or owner sign-off.

## Files

- `x/websocket/module.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/x-websocket-public-api-inventory.md`
- `docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`

## Tests

- `go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`

## Docs Sync

Document that API inventory exists but does not replace release snapshot
evidence or owner approval.

## Done Definition

- Exported API inventory is current.
- Evidence distinguishes inventory from promotion evidence.
- Validation passes.
