# Card 0103

Priority: P1

Goal:
- Make `x/ops` and `x/observability` clearly non-overlapping in discovery guidance and module ownership language.

Scope:
- protected admin surfaces
- broader observability adapter guidance

Non-goals:
- Do not move metrics or tracing code in this card.
- Do not add new admin endpoints.

Files:
- `specs/extension-taxonomy.yaml`
- `x/ops/module.yaml`
- `x/observability/module.yaml`
- `docs/modules/x-ops/README.md`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`

Tests:
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Keep observability and ops guidance aligned across specs and architecture docs.

Done Definition:
- `x/ops` is explicitly reserved for protected admin endpoints.
- `x/observability` remains the broader adapter and export discovery root.
- Agent entrypoint guidance no longer leaves room for overlap.
