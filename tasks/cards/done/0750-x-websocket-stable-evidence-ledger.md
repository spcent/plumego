# 0750 - x/websocket stable evidence ledger

Status: done
Priority: P2
Primary module: `x/websocket`

## Goal

Record the current stable-readiness state after the selected audit items while
leaving unavailable release governance evidence explicit.

## Scope

- Update evidence with the final state of the selected runtime/security/config
  fixes.
- Keep `release_history_missing`, `api_snapshot_missing`, and
  `owner_signoff_missing` as blockers.
- Leave `x/websocket/module.yaml` as `experimental`.
- Empty `tasks/cards/active` after this card is complete.

## Non-goals

- Fabricating release refs.
- Promoting the module.
- Broad repo gates beyond manifest/workflow checks.

## Files

- `docs/extension-evidence/x-websocket.md`
- `x/websocket/module.yaml`
- `tasks/cards/active/README.md`

## Tests

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`

## Docs Sync

Evidence must remain precise: code quality improvements are not release
history, release snapshots, or owner sign-off.

## Done Definition

- Evidence reflects current code and remaining blockers.
- Module status remains `experimental`.
- Active task queue is empty.

## Outcome

- Updated x/websocket evidence with the final state for async security event
  handler shutdown semantics, `SetReadLimit(0)`, and API inventory.
- Kept `release_history_missing`, `api_snapshot_missing`, and
  `owner_signoff_missing` explicit.
- Confirmed `x/websocket/module.yaml` remains `experimental`.
- Emptied the active task queue.

## Validations

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
