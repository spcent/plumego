# Card 0302

Priority: P1

Goal:
- Ensure transport health endpoint ownership remains outside the stable `health` root and is documented consistently.

Scope:
- `health` docs and metadata
- reference and extension guidance for HTTP health exposure

Non-goals:
- Do not add or move health HTTP handlers in this card.
- Do not change readiness logic.

Files:
- `health/module.yaml`
- `docs/modules/health/README.md`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `README.md`

Tests:
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Keep `health` ownership guidance consistent across top-level docs and primers.

Done Definition:
- The repository no longer leaves room to misread `health` as an HTTP endpoint package.
- Reference and extension guidance stay aligned with the stable boundary.
