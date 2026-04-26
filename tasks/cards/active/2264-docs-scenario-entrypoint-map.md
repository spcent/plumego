# Card 2264

Milestone:
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: active
Primary Module: core
Owned Files:
- README.md
- docs/README.md
- docs/getting-started.md
- docs/ROADMAP.md
Depends On: 2263

Goal:
Add a user-facing scenario entrypoint map that turns the current framework positioning into clear adoption paths.

Scope:
- Define concise paths for REST API service, multi-tenant API, edge gateway, realtime service, and AI service.
- Point each path to the correct stable roots, `x/*` family, and reference/example surface.
- Keep `reference/standard-service` as the canonical bootstrap path.
- Make the map descriptive of implemented behavior only.

Non-goals:
- Do not add new runtime features.
- Do not create a new canonical app layout.
- Do not document planned extensions as already stable.

Files:
- `README.md`
- `docs/README.md`
- `docs/getting-started.md`
- `docs/ROADMAP.md`

Tests:
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/reference-layout`
- `go test -timeout 20s ./reference/standard-service/...`

Docs Sync:
- Required; this is a docs and onboarding card.

Done Definition:
- A new user can identify the right first module for each major scenario without reading `specs/*`.
- The docs still describe `x/*` as optional capability families, not bootstrap defaults.

Outcome:
