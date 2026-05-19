# 1250 - x/websocket API freeze governance

Status: done
Priority: P2
State: done
Primary module: `x/websocket`

## Goal

Refresh websocket stable-readiness evidence after this cleanup while preserving
real governance blockers.

## Scope

- Refresh current-head API snapshot.
- Update public API inventory for any new/changed exported errors or semantics.
- Update evidence with fixed issues and remaining blockers.
- Keep module status `experimental`.
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

Evidence must distinguish implemented code quality from missing release
history, release snapshots, and owner approval.

## Done Definition

- Evidence and API inventory match current code.
- Governance blockers remain explicit.
- Active queue is empty.

## Outcome

- Refreshed the current-head API snapshot; exported symbol count remains 150.
- Updated beta evidence with route setup validation, outbound protocol guards,
  payload ownership, finite write deadlines, broadcast input validation, stream
  contract wording, and security event shutdown semantics.
- Updated public API inventory current-decision notes for the latest runtime
  cleanup without claiming stable promotion.
- Kept `x/websocket/module.yaml` status as `experimental` and preserved release
  refs, release snapshot, and owner sign-off blockers.
- Emptied `tasks/cards/active`.

## Validations

- `go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
