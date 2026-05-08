# Card 1381

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: blocked
Primary Module: x/websocket
Owned Files:
- `x/websocket/module.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/modules/x-websocket/README.md`
- `tasks/cards/active/1374-x-websocket-stable-evidence-readiness.md`

Goal:
- Close the remaining governance evidence needed before any stable promotion decision.

Scope:
- Record real release refs when they exist.
- Generate release-to-release API snapshots rather than current-head-only snapshots.
- Record `realtime` owner sign-off.
- Keep `x/websocket` experimental until evidence is complete.

Non-goals:
- Do not fabricate release refs.
- Do not promote based on runtime gates alone.
- Do not change runtime behavior in this card.

Files:
- `x/websocket/module.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/modules/x-websocket/README.md`
- `tasks/cards/active/1374-x-websocket-stable-evidence-readiness.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required when evidence or maturity state changes.

Done Definition:
- Real release refs, API snapshots, and owner sign-off exist.
- Maturity/evidence checks pass.
- Module status is promoted only if the evidence gate is satisfied.
