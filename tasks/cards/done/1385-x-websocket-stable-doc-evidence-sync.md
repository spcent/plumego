# Card 1385

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: done
Primary Module: x/websocket
Owned Files:
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`
- `x/websocket/module.yaml`
- `tasks/cards/done/1374-x-websocket-stable-evidence-readiness.md`
- `tasks/cards/done/1381-x-websocket-stable-governance-closure.md`
Depends On:
- 1374
- 1381

Goal:
- Synchronize WebSocket docs and evidence after cards 0773-0784 and record the
  resolved beta governance state.

Scope:
- Update public API inventory and runtime gate narrative for the newest cards.
- Align README public entrypoints with `module.yaml`.
- Keep beta status and future GA obligations explicit.

Non-goals:
- Do not fabricate release evidence.
- Do not promote beyond checked release evidence.
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
- Beta governance evidence is explicit.

Outcome:
- README, manifest, and evidence describe `x/websocket` as beta with
  release-backed API snapshots and `realtime` owner sign-off.
- Remaining GA obligations are left to future promotion work.

Validation:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/extension-beta-evidence
- go run ./internal/checks/extension-maturity
