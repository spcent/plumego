# Card 1093: x/frontend Stable Blocker Sync

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: done
Primary Module: x/frontend
Owned Files:
- `docs/extension-evidence/x-frontend.md`
- `docs/EXTENSION_MATURITY.md`
- `docs/modules/x-frontend/README.md`
- `x/frontend/module.yaml`
Depends On: 0747

Goal:
Sync stable readiness evidence and remaining blockers after the latest
hardening pass without promoting `x/frontend`.

Scope:
- Update evidence docs with the latest behavior and validation status.
- Keep `x/frontend` status `experimental`.
- Keep blockers for release history, release-backed API snapshots, and owner
  sign-off.
- Mention that current-head snapshots are not release evidence.

Non-goals:
- Do not invent release refs.
- Do not clear blockers requiring human sign-off or real releases.
- Do not change runtime behavior.

Files:
- `docs/extension-evidence/x-frontend.md`
- `docs/EXTENSION_MATURITY.md`
- `docs/modules/x-frontend/README.md`
- `x/frontend/module.yaml`

Tests:
- `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot`
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
This is a docs/evidence sync card.

Done Definition:
- Evidence docs match current `x/frontend` behavior and known blockers.
- No promotion status changes are made.
- The listed validation commands pass.

Outcome:
- Updated `x/frontend` evidence with the latest hardening coverage: `http.Dir`
  safety convergence, directory precompressed scan fail-fast behavior,
  negotiation parser convergence, and behavior-focused test organization.
- Kept `x/frontend` status `experimental` and preserved blockers for release
  history, release-backed API snapshots, and owner sign-off.
- Clarified that current-head snapshots are development evidence only.
- Validation passed:
  - `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot`
  - `go run ./internal/checks/extension-beta-evidence`
  - `go run ./internal/checks/extension-maturity`
