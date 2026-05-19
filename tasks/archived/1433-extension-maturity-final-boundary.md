# Card 1433

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: extension-evidence
Owned Files:
- `docs/EXTENSION_MATURITY.md`
- `specs/extension-beta-evidence.yaml`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`
Depends On:
- 1432

Goal:
- Record final extension maturity boundaries for the v1 decision.

Scope:
- Confirm beta modules remain backed by complete evidence.
- Confirm experimental modules keep blockers where evidence is incomplete.
- Update release evidence only with actual status, not planned promotions.

Non-goals:
- Do not promote experimental modules without complete evidence.
- Do not use `HEAD` as release-history evidence.
- Do not change extension runtime behavior.

Files:
- `docs/EXTENSION_MATURITY.md`
- `specs/extension-beta-evidence.yaml`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`

Tests:
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-beta-evidence`
- `git status --short --branch`

Docs Sync:
- Required for any maturity dashboard or release evidence wording change.

Done Definition:
- Beta and experimental boundaries are explicit.
- Evidence checks pass.
- Missing evidence remains listed as blockers.

Outcome:
- Recorded extension maturity boundary evidence in
  `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`.
- Updated `docs/release/v1.0.0-rc.1.md` with final extension maturity
  boundary evidence for the rc.
- Validation passed:
  - `go run ./internal/checks/extension-maturity`
  - `go run ./internal/checks/extension-beta-evidence`
  - `git status --short --branch`
- Beta remains limited to `x/gateway`, `x/observability`, `x/rest`, and
  `x/websocket`.
- Remaining candidates keep explicit blockers; no extension status was promoted.
