# 0725 - x/websocket release governance blockers

Status: active
Priority: P3
Primary module: `x/websocket`

## Problem

`x/websocket` remains `experimental`; evidence still lacks release refs,
release API snapshots, and owner sign-off. Those cannot be manufactured by code
cleanup.

## Scope

- Update evidence docs with the current cleanup state and remaining blockers.
- Keep `module.yaml` status as `experimental`.
- Record exact external inputs required for beta/stable promotion.

## Out of Scope

- Inventing release evidence.
- Owner sign-off without owner approval.
- Promotion status changes.

## Validation

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`

