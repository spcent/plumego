# Card 1385

Milestone: M-003
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: active
Primary Module: x/websocket
Owned Files:
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`
- `x/websocket/module.yaml`
- `tasks/cards/active/1374-x-websocket-stable-evidence-readiness.md`
- `tasks/cards/active/1381-x-websocket-stable-governance-closure.md`

Goal:
- Synchronize WebSocket docs and evidence after cards 0773-0784 while keeping governance blockers explicit.

Scope:
- Update public API inventory and runtime gate narrative for the newest cards.
- Keep `x/websocket` marked experimental.
- Keep missing release refs, release snapshots, and owner sign-off explicit.
- Align README public entrypoints with `module.yaml`.

Non-goals:
- Do not fabricate release evidence.
- Do not promote maturity status.
- Do not change runtime behavior.

Files:
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`
- `x/websocket/module.yaml`
- `tasks/cards/active/1374-x-websocket-stable-evidence-readiness.md`
- `tasks/cards/active/1381-x-websocket-stable-governance-closure.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- This is the docs/evidence sync card.

Done Definition:
- README, manifest, and evidence no longer describe stale 0761-0772-only runtime state.
- Governance blockers remain explicit.
