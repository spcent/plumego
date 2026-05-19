# Card 1386

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: done
Primary Module: x/websocket
Owned Files:
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/modules/x-websocket/README.md`
- `x/websocket/module.yaml`
Depends On:
- 1381

Goal:
- Close WebSocket release-governance evidence only when real release refs and owner sign-off exist.

Scope:
- Record two concrete release refs.
- Generate release-to-release API snapshots.
- Record `realtime` owner sign-off.
- Promote only if the evidence gate passes.

Non-goals:
- Do not fabricate release refs.
- Do not use current-head snapshots as release evidence.
- Do not change runtime behavior.

Files:
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/modules/x-websocket/README.md`
- `x/websocket/module.yaml`

Tests:
- `go run ./internal/checks/extension-release-evidence`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required when real release evidence is available.

Done Definition:
- Real release refs, matching snapshots, and owner sign-off are checked in.
- `x/websocket` is promoted only after release governance is satisfied.

Outcome:
- Real release refs `d2c25c3` and `ec70358`, matching release snapshots, and
  `realtime` owner sign-off are checked in.
- `x/websocket` is beta; no GA promotion is implied.
- No runtime behavior changed in this card.

Validation:
- go run ./internal/checks/extension-beta-evidence
- go run ./internal/checks/extension-maturity
- go run ./internal/checks/module-manifests
