# Plan for M-014: OpenAPI 3.1 Generation

Milestone: `M-014`
Objective: Ship x/openapi with a route-driven OpenAPI 3.1 document generator
and integrate it into the plumego CLI as `plumego generate spec`, turning the
framework's explicit route registration into a first-class API documentation
capability without any external OpenAPI library dependency.
Constraints: x/openapi depends only on router, contract, and stdlib; no external
OpenAPI library in any go.mod; Op hints are optional; generator reads router.Routes()
directly; CLI writes to stdout or --output path; YAML via --format yaml flag.
Affected Modules: x/openapi, cmd/plumego, reference/with-rest.

## Phase Map

- Phase 1: Orient — read router.go to understand RouteInfo shape; read existing
  generate.go command to understand CLI wiring pattern.
- Phase 2: Implement (parallel) — write openapi core generator, marshal module,
  and CLI subcommand concurrently.
- Phase 3: Test — write openapi_test.go and CLI integration test covering empty,
  static, parameterised, hinted, and unhinted routes.
- Phase 4: Reference and Validate — add reference/with-rest example, run acceptance
  criteria, commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1540 | Create x/openapi core generator with Op hint type and Generator struct | x/openapi | `x/openapi/openapi.go`, `x/openapi/module.yaml` | M-009 | `go test ./x/openapi/...`, `go vet ./x/openapi/...` |
| 1541 | Create x/openapi/marshal.go with JSON and YAML serialisation | x/openapi | `x/openapi/marshal.go` | 1540 | `go test ./x/openapi/...` |
| 1542 | Add plumego generate spec subcommand to cmd/plumego | cmd/plumego | `cmd/plumego/commands/generate.go` | 1540, 1541 | `go test ./cmd/plumego/...`, command exits 0 |

## Dependency Edges

- `1540 -> 1541`
- `1540 -> 1542`
- `1541 -> 1542`

## Parallel Groups

- Group A: card 1540 — must complete first; defines the Generator and Op types.
- Group B (parallel after A): cards 1541 and 1542 are independent consumers of 1540.
  In practice 1542 needs 1541's marshalling, so run 1541 before 1542.
- Group C (sequential): tests and reference example after all three cards complete.

## Risk Register

- Risk: router.RouteInfo does not expose enough information (method, path, name) for
  a meaningful spec.
  Mitigation: card 1540 reads router/router.go and router/module.yaml before writing;
  stop and flag if RouteInfo is missing required fields.
- Risk: YAML serialisation tempts a third-party dependency.
  Mitigation: keep x/openapi in the main module and use dependency-free
  serialization; no `go.mod` may exist under `x/**`.

## Verification Strategy

- Card-level checks: `go test ./x/openapi/...` after each openapi card; `go test ./cmd/plumego/...`
  and `plumego generate spec --output /dev/null` after 1542.
- Dependency audit: `go run ./internal/checks/dependency-rules` confirms no external
  OpenAPI library in the main module go.mod.
- Reference check: reference/with-rest Makefile `make spec` target runs without error.

## Exit Condition

- all three implementation cards completed and tests written
- x/openapi/openapi.go with Generator and Op type exists
- `plumego generate spec` exits 0 on a minimal app
- x/openapi/module.yaml exists with status = experimental
- reference/with-rest shows annotated routes and spec generation
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
