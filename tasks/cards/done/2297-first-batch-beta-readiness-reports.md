# Card 2297

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P1
State: done
Primary Module: docs
Owned Files:
- docs/extension-evidence/x-rest.md
- docs/extension-evidence/x-websocket.md
- docs/extension-evidence/x-tenant.md
- docs/extension-evidence/x-gateway.md
- docs/extension-evidence/x-observability.md
Depends On: 2296

Goal:
Use the release evidence workflow to update first-batch beta readiness reports
without promoting modules.

Scope:
- Add release-evidence command examples to first-batch evidence docs.
- Record current blocker state consistently for all first-batch candidates.
- Keep all module statuses experimental.

Non-goals:
- Do not update `module.yaml` status values.
- Do not claim release-history evidence without real refs.
- Do not change code behavior.

Files:
- `docs/extension-evidence/x-rest.md`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/x-tenant.md`
- `docs/extension-evidence/x-gateway.md`
- `docs/extension-evidence/x-observability.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`
- `scripts/check-spec tasks/cards/done/2297-first-batch-beta-readiness-reports.md`

Docs Sync:
- Required because beta readiness guidance changes.

Done Definition:
- First-batch evidence records all point to the same release-aware workflow.
- Remaining blockers are explicit and not overstated as completed.

Outcome:
- Added release-aware comparison workflow commands to first-batch beta evidence
  docs for `x/rest`, `x/websocket`, `x/tenant`, `x/gateway`, and
  `x/observability`.
- Kept all first-batch blockers intact until real release refs, snapshot files,
  and owner sign-off exist.

Validations:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`
- `scripts/check-spec tasks/cards/done/2297-first-batch-beta-readiness-reports.md`
