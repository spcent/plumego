# Card 2271

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: done
Primary Module: x/observability
Owned Files:
- docs/modules/x-observability/README.md
- docs/modules/x-ops/README.md
- docs/modules/x-devtools/README.md
- reference/standard-service/README.md
Depends On: 2270

Goal:
Document a production observability and operations profile that separates runtime telemetry, protected ops routes, and local devtools.

Scope:
- Describe when to use stable metrics middleware, `x/observability`, `x/ops`, and `x/devtools`.
- Clarify that devtools routes are local or protected-environment tools.
- Define a recommended production wiring checklist using implemented packages only.
- Keep health HTTP ownership outside stable `health`.

Non-goals:
- Do not add new ops endpoints.
- Do not expose devtools by default.
- Do not move exporter wiring into stable roots.

Files:
- `docs/modules/x-observability/README.md`
- `docs/modules/x-ops/README.md`
- `docs/modules/x-devtools/README.md`
- `reference/standard-service/README.md`

Tests:
- `go test -timeout 20s ./x/observability/...`
- `go test -timeout 20s ./x/ops/... ./x/devtools/...`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- Required; this is an operational guidance card.

Done Definition:
- Production telemetry, protected ops, and local debug surfaces have distinct documented ownership.
- No doc implies devtools are part of the default kernel path.

Outcome:
Completed. Documented the production ownership split across stable middleware,
`x/observability`, `x/ops`, and `x/devtools`, including guidance that devtools
routes are local or explicitly protected surfaces and not part of canonical
production bootstrap.
