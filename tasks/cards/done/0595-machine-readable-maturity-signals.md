# Card 0595

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: specs
Owned Files:
- specs/extension-maturity.yaml
- internal/checks/extension-maturity/main.go
- internal/checks/extension-maturity/README.md
- docs/EXTENSION_MATURITY.md
- tasks/cards/active/README.md
Depends On: 2302

Goal:
Move maturity dashboard signals for recommended entrypoint, docs state, and
coverage state into a machine-readable spec that the dashboard checker verifies.

Scope:
- Add a small `specs/extension-maturity.yaml` source for dashboard-only signals.
- Extend `extension-maturity` to validate dashboard rows against the spec.
- Keep module manifest as authority for status, risk, and owner.

Non-goals:
- Do not move all prose out of the dashboard.
- Do not change module status.
- Do not add coverage tooling beyond declared signal verification.

Files:
- `specs/extension-maturity.yaml`
- `internal/checks/extension-maturity/main.go`
- `internal/checks/extension-maturity/README.md`
- `docs/EXTENSION_MATURITY.md`
- `tasks/cards/active/README.md`

Tests:
- `go test ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/agent-workflow`
- `scripts/check-spec tasks/cards/done/0595-machine-readable-maturity-signals.md`
- `scripts/check-spec tasks/cards/done/0595-machine-readable-maturity-signals.md`

Docs Sync:
- Required because dashboard source-of-truth behavior changes.

Done Definition:
- Dashboard recommendation/docs/coverage signals are checked from a spec rather
  than maintained only as prose.

Outcome:
- Added `specs/extension-maturity.yaml` as the machine-readable source for
  dashboard-only recommended entrypoint, docs, and coverage signals.
- Extended `extension-maturity` to require every declared extension root to have
  a signal entry and to verify those signals appear in the dashboard row.
- Updated `docs/EXTENSION_MATURITY.md` to include a checked `Signals` column
  for app-facing families and subordinate primitives.

Validations:
- `go test ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/agent-workflow`
