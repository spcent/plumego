# Card 0003

Priority: P1

Goal:
- Expand `specs/agent-entrypoints.yaml` to cover remaining common task families at finer granularity.

Scope:
- frontend
- devtools
- discovery
- ops/admin surfaces

Non-goals:
- Do not change extension code in this card.
- Do not redefine already-stable entrypoints for `security`, `store`, `tenant`, or `websocket`.

Files:
- `specs/agent-entrypoints.yaml`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `README.md`
- `README_CN.md`

Tests:
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Keep README and blueprint entrypoint guidance aligned with the expanded task map.

Done Definition:
- Agents have a documented first-read path for the remaining common extension workflows.
- Discovery defaults are consistent across specs and top-level docs.
