# 0821 - x/websocket release governance blockers

Status: done
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

## Outcome

- Updated `docs/extension-evidence/x-websocket.md` with the current cleanup
  state from the stable-readiness pass.
- Preserved `x/websocket/module.yaml` status as `experimental`.
- Recorded exact external evidence still required for promotion: consecutive
  release refs, release API snapshots/comparison output, and named realtime
  owner sign-off.
- Validation passed:
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/agent-workflow`
