# 0734 - x/websocket stable readiness evidence refresh

Status: active
Priority: P2
Primary module: `x/websocket`

## Problem

After the implementation pass, the evidence record must reflect what has been
fixed and what still blocks promotion. Governance evidence cannot be invented.

## Scope

- Refresh `docs/extension-evidence/x-websocket.md` with this pass's outcomes.
- Keep `module.yaml` status as `experimental` unless real release evidence and
  owner sign-off exist.
- Record remaining external inputs required for beta/stable promotion.
- Leave `tasks/cards/active` empty after the card is done.

## Out of Scope

- Creating fake release snapshots.
- Owner sign-off without owner approval.
- Promotion status changes.

## Validation

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`

