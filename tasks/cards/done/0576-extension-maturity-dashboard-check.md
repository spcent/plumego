# Card 0576

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: internal/checks
Owned Files:
- internal/checks/extension-maturity/main.go
- docs/EXTENSION_MATURITY.md
- specs/checks.yaml
- docs/ROADMAP.md
Depends On: 2285

Goal:
Add a lightweight check that catches maturity dashboard drift against module manifests.

Scope:
- Implement a check that reads `x/*/module.yaml` files and verifies every extension root appears in `docs/EXTENSION_MATURITY.md`.
- Verify dashboard rows include current manifest `status` and `risk` values.
- Register the check in `specs/checks.yaml` and roadmap/tooling guidance.

Non-goals:
- Do not parse every prose field in the dashboard.
- Do not make the check decide beta eligibility.
- Do not change module manifest schema.

Files:
- `internal/checks/extension-maturity/main.go`
- `docs/EXTENSION_MATURITY.md`
- `specs/checks.yaml`
- `docs/ROADMAP.md`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-maturity`
- `scripts/check-spec tasks/cards/done/0576-extension-maturity-dashboard-check.md`

Docs Sync:
- Required because a new repo check is added.

Done Definition:
- Dashboard drift is detected locally before review.
- The check uses module manifests as the source of truth.
- The check is documented with the rest of the repository control-plane checks.

Outcome:
- Added `internal/checks/extension-maturity` to compare dashboard rows against
  extension module manifests.
- The check verifies each declared extension root appears in
  `docs/EXTENSION_MATURITY.md` with the current manifest `status` and `risk`.
- Registered the check in `specs/checks.yaml` and documented it in the
  dashboard and roadmap tooling guidance.

Validations:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-maturity`
