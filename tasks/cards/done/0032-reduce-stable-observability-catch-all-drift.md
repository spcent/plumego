# Card 0032

Priority: P2

Goal:
- Reduce ambiguity around observability catch-all behavior in stable roots and keep adapter-heavy observability work in the right extension layer.

Scope:
- stable middleware observability wording
- `x/observability` relationship to stable roots

Non-goals:
- Do not move runtime observability code in this card.
- Do not add new exporters.

Files:
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`
- `x/observability/module.yaml`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `README.md`

Tests:
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Keep stable transport observability guidance distinct from extension adapter guidance.

Done Definition:
- Stable roots are not described as broad observability catalogs.
- `x/observability` remains the adapter/export layer without pulling transport primitives out of `middleware`.
