# Card 2305

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
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

Docs Sync:
- Required because dashboard source-of-truth behavior changes.

Done Definition:
- Dashboard recommendation/docs/coverage signals are checked from a spec rather
  than maintained only as prose.

Outcome:
