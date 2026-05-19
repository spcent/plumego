# Card 1239: x/frontend Stable Release Evidence Governance

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`
- `tasks/cards/done/1239-x-frontend-stable-release-evidence-governance.md`
Depends On: 0760

Goal:
Record the remaining stable promotion evidence work without falsely clearing
external release blockers.

Scope:
- Keep `x/frontend` experimental until release-backed evidence exists.
- Record that release-backed API snapshots require concrete release refs, not a
  current-head snapshot.
- Record that frontend owner sign-off and candidate release gate are required
  external governance steps.
- Run a local candidate gate as current-head evidence only.

Non-goals:
- Do not fabricate release refs.
- Do not clear `api_snapshot_missing`, `release_history_missing`, or
  `owner_signoff_missing`.
- Do not change module status.

Files:
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`
- `tasks/cards/done/1239-x-frontend-stable-release-evidence-governance.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`
- `GOCACHE=/private/tmp/plumego-gocache make gates`

Docs Sync:
Update evidence and manifest comments only.

Done Definition:
- Remaining stable release evidence steps are explicit and unambiguous.
- Current-head validation is recorded without clearing release blockers.
- The listed validation commands pass.

Outcome:
- Recorded current-head `make gates` success as candidate-state evidence while
  preserving the release-backed evidence distinction.
- Kept `api_snapshot_missing`, `release_history_missing`, and
  `owner_signoff_missing` uncleared because they require real release refs and
  human owner sign-off.
- Added the shortest path to stable in the x/frontend evidence ledger.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
  - `GOCACHE=/private/tmp/plumego-gocache make gates`
