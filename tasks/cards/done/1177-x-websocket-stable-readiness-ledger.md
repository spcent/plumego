# 1177 - x/websocket stable readiness ledger

Status: done
Priority: P2
Primary module: `x/websocket`

## Goal

Refresh x/websocket stable-readiness evidence after this pass without claiming
release governance that does not exist.

## Scope

- Refresh current-head API snapshot.
- Update public API inventory for the renamed/removed config surface.
- Update evidence with fixed items and remaining blockers.
- Keep `module.yaml` status as `experimental`.
- Empty `tasks/cards/active`.

## Non-goals

- Fabricating release refs.
- Claiming owner sign-off.
- Promoting to beta or stable.

## Files

- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/x-websocket-public-api-inventory.md`
- `docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `x/websocket/module.yaml`
- `tasks/cards/active/README.md`

## Tests

- `go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`

## Docs Sync

Evidence must separate implemented code quality from release refs, release
snapshots, and owner approval.

## Done Definition

- Evidence and API inventory match current code.
- Governance blockers remain explicit.
- Active queue is empty.

## Outcome

- Refreshed the current-head `x/websocket` API snapshot.
- Updated the public API inventory for the exported `RouteRegistrar`, current
  symbol counts, removed `SecurityConfig` Hub fields, and
  `EnableSecurityEvents` naming.
- Updated beta evidence with the completed hub guards, bounded security event
  dispatch, public route registrar, config surface cleanup, and message read
  contract work.
- Kept `x/websocket/module.yaml` status as `experimental` and preserved release
  refs, release snapshot, and owner sign-off blockers.
- Emptied `tasks/cards/active`.

## Validations

- `go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
