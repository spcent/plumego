# Card 0581

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P1
State: done
Primary Module: internal/checks
Owned Files:
- internal/checks/extension-maturity/main.go
- internal/checks/extension-maturity/README.md
- docs/EXTENSION_MATURITY.md
- specs/checks.yaml
- tasks/cards/active/README.md
Depends On: 2285, 2286, 2290

Goal:
Strengthen maturity dashboard automation so dashboard rows are derived or
strictly checked against manifests and beta evidence.

Scope:
- Extend the existing maturity check to validate evidence links and candidate
  blocker text for beta candidates.
- Add a deterministic report mode if it stays small.
- Document the dashboard maintenance command.

Non-goals:
- Do not replace the human-readable dashboard with generated-only output.
- Do not promote modules.
- Do not change module status values.

Files:
- `internal/checks/extension-maturity/main.go`
- `internal/checks/extension-maturity/README.md`
- `docs/EXTENSION_MATURITY.md`
- `specs/checks.yaml`
- `tasks/cards/active/README.md`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-maturity`
- `scripts/check-spec tasks/cards/done/0581-extension-maturity-dashboard-generation-check.md`

Docs Sync:
- Required because dashboard maintenance behavior changes.

Done Definition:
- Dashboard drift catches status, risk, and beta evidence link mismatches.
- Maintainers have a documented command for dashboard validation/reporting.

Outcome:
- Extended `extension-maturity` to validate dashboard owner values against
  extension `module.yaml` files.
- Added beta-candidate dashboard checks for evidence links and blocker text
  against `specs/extension-beta-evidence.yaml`.
- Added `-report` mode and documented dashboard drift/report commands.

Validations:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-maturity -report`
- `scripts/check-spec tasks/cards/done/0581-extension-maturity-dashboard-generation-check.md`
