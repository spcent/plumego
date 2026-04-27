# Card 0028

Priority: P2

Goal:
- Clarify reusable response/error conventions and layering guidance for `x/rest` without turning it into a bootstrap surface.

Scope:
- `x/rest` docs and metadata
- resource-controller layering guidance

Non-goals:
- Do not introduce a new response helper family in this card.
- Do not add hidden controller binding.

Files:
- `x/rest/README.md`
- `x/rest/module.yaml`
- `docs/ROADMAP.md`
- `README.md`

Tests:
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Keep `x/rest` guidance aligned across the module README, manifest, roadmap, and top-level README.

Done Definition:
- `x/rest` response and layering guidance is explicit and non-overlapping with bootstrap or gateway concerns.
- Top-level docs do not leave room for conflicting usage patterns.
