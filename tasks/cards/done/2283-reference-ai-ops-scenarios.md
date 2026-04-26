# Card 2283

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: active
Primary Module: reference
Owned Files:
- reference/with-ai/README.md
- reference/with-ai/main.go
- reference/with-ops/README.md
- reference/with-ops/main.go
- docs/README.md
Depends On: 2282

Goal:
Add runnable AI and operations scenario references that demonstrate safe first-use paths without broadening stable roots.

Scope:
- Add `reference/with-ai` using only `x/ai/provider`, `x/ai/session`, `x/ai/streaming`, and `x/ai/tool` with mock/offline behavior.
- Add `reference/with-ops` showing protected operations surfaces and explicit observability wiring.
- Keep debug/devtools routes off by default and document the production boundary.
- Link the scenarios from the docs entrypoint.

Non-goals:
- Do not use live AI provider credentials or network calls.
- Do not mount `x/devtools` as a production admin surface.
- Do not promote `x/ai`, `x/ops`, or `x/observability`.

Files:
- `reference/with-ai/README.md`
- `reference/with-ai/main.go`
- `reference/with-ops/README.md`
- `reference/with-ops/main.go`
- `docs/README.md`

Tests:
- `go test -timeout 20s ./reference/with-ai ./reference/with-ops`
- `go run ./internal/checks/reference-layout`
- `scripts/check-spec tasks/cards/done/2283-reference-ai-ops-scenarios.md`

Docs Sync:
- Required because new scenario references are added.

Done Definition:
- AI and ops examples are runnable without external services.
- AI usage stays on the stable-tier subpackage path.
- Ops/debug/observability production boundaries are explicit in the example docs.

Outcome:
- Added `reference/with-ai` using only `x/ai/provider`, `x/ai/session`,
  `x/ai/streaming`, and `x/ai/tool` with offline mock behavior.
- Added `reference/with-ops` with protected `x/ops` routes and stable request
  observability middleware, without mounting `x/devtools`.
- Linked both scenarios from `docs/README.md`.
- Documented that the involved `x/*` modules remain experimental until
  promotion evidence is complete.

Validations:
- `go test -timeout 20s ./reference/with-ai ./reference/with-ops`
- `go run ./internal/checks/reference-layout`
