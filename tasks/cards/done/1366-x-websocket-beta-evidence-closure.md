# Card 1366

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: done
Primary Module: x/websocket
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/release-artifacts.md`
Depends On: release refs and owner sign-off

Goal:
- Complete beta evidence closure for `x/websocket` when real release refs and owner sign-off are available.

Problem:
The evidence ledger originally had only a current-head snapshot for
`x/websocket`. The card closed after release refs, matching snapshots, and
owner sign-off were recorded.

Scope:
- Add two real release refs after tags or release commits exist.
- Generate release-to-release API snapshots with `extension-release-evidence`.
- Record owner sign-off from `realtime`.
- Clear blockers only after all evidence is present.

Non-goals:
- Do not promote `x/websocket` without complete evidence.
- Do not use `HEAD` as release-history evidence.
- Do not change WebSocket runtime behavior.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/release-artifacts.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Required when evidence is added.

Done Definition:
- `x/websocket` has two release refs, matching release snapshots, and owner sign-off.

Outcome:
- Release refs `d2c25c3` and `ec70358`, release-backed API snapshots, and
  `realtime` owner sign-off are recorded.
- `x/websocket/module.yaml` is `status: beta`.
- `docs/modules/x-websocket/README.md`,
  `docs/extension-evidence/x-websocket.md`, and
  `specs/extension-beta-evidence.yaml` agree that blockers are cleared.

Validation:
- go run ./internal/checks/extension-beta-evidence
- go run ./internal/checks/extension-maturity
