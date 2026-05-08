# Card 0722

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: done
Primary Module: specs
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `specs/extension-maturity.yaml`
- `docs/extension-evidence/release-artifacts.md`
- `tasks/cards/active/README.md`
Depends On: —

Goal:
- Convert extension beta evidence gaps into explicit module-owned follow-up work.

Problem:
`go run ./internal/checks/extension-beta-evidence` reports missing API snapshots, owner signoff, and release history for multiple experimental or stable-tier extension surfaces including `x/gateway`, `x/rest`, `x/tenant`, `x/websocket`, `x/messaging`, `x/data`, and `x/ai`.

Scope:
- Re-run the evidence checker and record current blockers.
- Decide which gaps are documentation-only, API snapshot work, owner signoff work, or release-history work.
- Update evidence docs/specs only where the evidence already exists.
- Create separate follow-up cards for each module where evidence is missing.

Non-goals:
- Do not promote extension maturity without evidence.
- Do not fabricate release references or owner signoff.
- Do not change extension runtime code.

Files:
- `specs/extension-beta-evidence.yaml`
- `specs/extension-maturity.yaml`
- `docs/extension-evidence/release-artifacts.md`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required. This card is evidence/control-plane cleanup.

Done Definition:
- Evidence checker output is mapped to concrete module-owned next steps.
- Existing evidence is linked in specs/docs.
- Missing evidence remains visible and is not hidden by broad wording.

Outcome:
- Re-ran `go run ./internal/checks/extension-beta-evidence` and recorded the current blocker map in `docs/extension-evidence/release-artifacts.md`.
- Confirmed existing evidence:
  - first-batch candidates (`x/rest`, `x/websocket`, `x/tenant`, `x/observability`, `x/gateway`) have evidence docs and current-head snapshots.
  - `x/ai` stable-tier subpackages have evidence docs and current-head snapshots.
  - `x/data`, `x/discovery`, and `x/messaging` have evidence docs but still need snapshots.
- Preserved blockers for missing release history, release snapshots, and owner sign-off; no maturity promotion was made.
- Created blocked module-owned follow-up cards:
  - `0723` `x/rest`
  - `0724` `x/websocket`
  - `0725` `x/tenant`
  - `0726` `x/observability`
  - `0727` `x/gateway`
  - `0728` `x/ai` stable-tier subpackages
  - `0729` `x/data` selected surfaces
  - `0730` `x/discovery` core/static surface
  - `0731` `x/messaging` app-facing service
- Validation:
  - `go run ./internal/checks/extension-maturity`
  - `go run ./internal/checks/extension-beta-evidence`
  - `go run ./internal/checks/module-manifests`
