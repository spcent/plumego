# Card 0766: x/frontend Release-backed Evidence Blockers

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P1
State: active
Primary Module: x/frontend
Owned Files:
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`
- `tasks/cards/done/0766-x-frontend-release-backed-evidence-blockers.md`
Depends On: 0765

Goal:
Keep stable promotion blockers accurate after the latest hardening pass.

Scope:
- Re-state that release-backed API snapshots require concrete release refs.
- Keep two-minor-release history and owner sign-off as uncleared blockers.
- Run current-head checks as supporting evidence only.

Non-goals:
- Do not fabricate release refs.
- Do not clear `api_snapshot_missing`, `release_history_missing`, or
  `owner_signoff_missing`.
- Do not change module status.

Files:
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`
- `tasks/cards/done/0766-x-frontend-release-backed-evidence-blockers.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`
- `GOCACHE=/private/tmp/plumego-gocache make gates`

Docs Sync:
Update evidence and manifest comments only.

Done Definition:
- Release-backed evidence blockers are accurate and explicit.
- Current-head checks are recorded without clearing external blockers.
- The listed validation commands pass.
