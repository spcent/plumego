# Card 2285

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: docs
Owned Files:
- docs/EXTENSION_MATURITY.md
- docs/README.md
- docs/ROADMAP.md
- specs/repo.yaml
Depends On: 2280

Goal:
Create a human-readable `x/*` family maturity dashboard for agent and maintainer triage.

Scope:
- Add a dashboard listing each extension family, status, risk, recommended entrypoint, test command, docs primer, and promotion blocker.
- Distinguish app-facing families from subordinate primitives.
- Link candidate beta evidence where available.
- Link the dashboard from the docs entrypoint and roadmap.

Non-goals:
- Do not change module status values.
- Do not replace module manifests as the machine-readable source of truth.
- Do not include stale or aspirational capabilities.

Files:
- `docs/EXTENSION_MATURITY.md`
- `docs/README.md`
- `docs/ROADMAP.md`
- `specs/repo.yaml`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `scripts/check-spec tasks/cards/done/2285-extension-maturity-dashboard.md`

Docs Sync:
- Required because the docs surface gains a new navigation artifact.

Done Definition:
- A maintainer can see the maturity state of every `x/*` family from one page.
- The dashboard clearly identifies beta blockers and recommended entrypoints.
- The dashboard does not supersede manifests or evidence files.

Outcome:
