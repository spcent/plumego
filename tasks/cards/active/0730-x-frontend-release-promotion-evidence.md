# Card 0730

Milestone:
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P3
State: blocked
Primary Module: x/frontend
Owned Files: docs/extension-evidence/x-frontend.md, x/frontend/module.yaml
Depends On: 0729
Blocked By: two concrete consecutive minor release refs and frontend owner sign-off

Goal:
Complete the external release-backed evidence required before changing x/frontend from experimental to stable.

Scope:
- Identify two concrete consecutive minor release refs that include `x/frontend`.
- Run release-backed API snapshot comparison for the exported x/frontend surface.
- Confirm no exported API churn across those refs.
- Re-run the full repository release gate from the final candidate ref.
- Record frontend owner sign-off.
- Only after all evidence exists, update `x/frontend/module.yaml` status.

Non-goals:
- Do not generate fake release refs.
- Do not treat current-head snapshots as release evidence.
- Do not promote without owner approval.

Files:
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`

Tests:
- `go run ./internal/checks/extension-release-evidence -module ./x/frontend -base <older-minor-release-ref> -head <newer-minor-release-ref> -out-dir <release-evidence-dir>`
- `GOCACHE=/private/tmp/plumego-gocache make gates`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
- Record exact release refs, snapshot output path, final gate command, and owner sign-off.

Done Definition:
- Release-backed API snapshot evidence exists.
- Two consecutive minor release refs are recorded.
- Final candidate release gate passes.
- Frontend owner sign-off is recorded.
- Only then is `x/frontend/module.yaml` eligible for status promotion.

Outcome:
- Blocked until concrete release refs and owner sign-off are available.

