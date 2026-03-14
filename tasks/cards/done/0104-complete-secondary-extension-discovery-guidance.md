# Card 0104

Priority: P2

Goal:
- Finish the documentation-level convergence for `x/frontend`, `x/devtools`, and `x/discovery` so each has a stable discovery role without leaking into bootstrap or stable roots.

Scope:
- secondary extension primers
- architecture and repo metadata wording

Non-goals:
- Do not add runtime features.
- Do not change quality gates.

Files:
- `specs/agent-entrypoints.yaml`
- `docs/modules/x-frontend/README.md`
- `docs/modules/x-devtools/README.md`
- `docs/modules/x-discovery/README.md`
- `README.md`

Tests:
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Keep top-level workflow guidance aligned with secondary extension primers.

Done Definition:
- frontend, devtools, and discovery each have a clear first-read path.
- None of these extensions are described as bootstrap surfaces.
- The top-level docs and primers are consistent.
