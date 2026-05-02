# Card 0761

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: todo
Primary Module: x/websocket
Owned Files:
- `x/websocket/module.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/snapshots/**`
- `docs/modules/x-websocket/README.md`
- `tasks/cards/active/0738-x-websocket-stable-evidence-readiness.md`
Depends On: none

Goal:
- Define the governance evidence required before `x/websocket` can be promoted toward stable.

Problem:
`x/websocket` is still `experimental`, and the evidence file explicitly records missing release refs, release API snapshots, and owner sign-off. Stable promotion must not proceed from current-head confidence alone.

Scope:
- Keep `x/websocket/module.yaml` at `experimental` until release evidence exists.
- Record the exact release refs required for stable consideration.
- Record the snapshot workflow and expected artifact locations.
- Record the `realtime` owner sign-off requirement.
- Align card 0738 with the updated evidence checklist.

Non-goals:
- Do not fabricate release refs or owner sign-off.
- Do not promote maturity status in this card.
- Do not change runtime behavior.

Files:
- `x/websocket/module.yaml`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/snapshots/**`
- `docs/modules/x-websocket/README.md`
- `tasks/cards/active/0738-x-websocket-stable-evidence-readiness.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Required for maturity state, evidence blockers, and promotion workflow.

Done Definition:
- Evidence blockers are explicit and cannot be accidentally cleared by HEAD-only snapshots.
- The stable promotion path requires real release refs, release snapshots, and owner sign-off.
- Maturity docs and active evidence cards agree.

Outcome:
-
