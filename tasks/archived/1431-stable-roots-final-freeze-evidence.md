# Card 1431

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: stable-roots
Owned Files:
- `docs/stable-api/README.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`
Depends On:
- 1430

Goal:
- Record final stable-root freeze evidence for the v1 decision.

Scope:
- Verify stable-root API freeze, deprecation strictness, and focused tests.
- Record stable-root command evidence in the M-005 verify artifact.
- Open bounded blocker cards if a stable-root gate fails.

Non-goals:
- Do not change stable-root public APIs.
- Do not add dependencies.
- Do not move behavior between stable roots and `x/*`.

Files:
- `docs/stable-api/README.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`

Tests:
- `go run ./internal/checks/deprecation-inventory -strict`
- `go test -race -timeout 60s ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics`
- `go vet ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics`

Docs Sync:
- Required only for recorded freeze evidence or blocker status.

Done Definition:
- Stable-root freeze evidence is recorded.
- Any failure is converted into a bounded blocker card.
- No unapproved public API change is introduced.

Outcome:
- Recorded stable-root final freeze evidence in `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`.
- Updated `docs/release/v1.0.0-rc.1.md` with the stable-root final freeze command evidence.
- Validation passed:
  - `go run ./internal/checks/deprecation-inventory -strict`
  - `go test -race -timeout 60s ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics`
  - `go vet ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics`
- No stable-root public API changes were made.
