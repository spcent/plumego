# M-014: OpenAPI 3.1 Generation

- **Branch:** `milestone/M-014-openapi-generation`
- **Depends on:** M-009
- **Parallel OK:** yes

---

## Goal

Graduate `x/openapi` from Experimental to Beta by implementing a `RouteInfo`-driven OpenAPI 3.1 document generator and a `plumego generate spec` CLI command, so applications can produce machine-readable API contracts without struct tag scanning or reflection.

---

## Architecture Decisions

- Generation is **declaration-first**: callers annotate routes at registration time using option functions; there is no post-hoc reflection scan over handler signatures.
- The API must be additive at the route registration call site:
  ```go
  app.Get("/users/{id}", handlers.GetUser,
      openapi.Summary("Get user by ID"),
      openapi.Tag("users"),
      openapi.Response(200, openapi.SchemaRef("User")),
      openapi.Response(404, openapi.SchemaRef("APIError")),
  )
  ```
- If no OpenAPI options are provided for a route, the route is still included in the document with a generated `operationId` and no schema.
- Output format: OpenAPI 3.1.0 JSON (canonical) and YAML (optional flag). No custom envelope.
- `plumego generate spec` reads a compiled binary's embedded spec (via a new `x/openapi.Embed()` call) or a Go file that imports the app wiring. Implementation detail left to Phase 2 analysis.
- `x/openapi` imports `router` (for `RouteInfo`) and `contract` (for `APIError` schema shape). It must not import any other stable root or `x/*` package.
- No external OpenAPI library dependency. Stdlib JSON marshaling only.
- After promotion, update `specs/extension-maturity.yaml` and `docs/concepts/extension-maturity.md`.

---

## Context — Read Before Touching Code

1. `AGENTS.md`
2. `specs/dependency-rules.yaml`
3. `specs/task-routing.yaml`
4. `x/openapi/module.yaml`
5. `router/router.go` (RouteInfo type)
6. `contract/errors.go` (APIError shape)
7. `specs/change-recipes/new-extension-module.yaml`
8. `reference/standard-service/internal/app/routes.go`

## Affected Modules

- **Primary:** `x/openapi`
- **Secondary:** `cmd/plumego`, `specs/extension-maturity.yaml`, `docs/concepts/extension-maturity.md`

---

## Tasks

### Phase 1 — Orient (sequential)

1. Read every file in the **Context** section above.
2. Read all existing files in `x/openapi/` to understand current state.
3. Confirm `router.RouteInfo` exports sufficient fields for path, method, and name. If not, document the gap as an Open Question before touching code.

### Phase 2 — Implement (parallel)

- [ ] `x/openapi/spec.go`: `Builder` type with `AddRoute(info router.RouteInfo, opts ...Option) *Builder` and `Build() (*Document, error)`. `Document` marshals to valid OpenAPI 3.1.0 JSON.
- [ ] `x/openapi/options.go`: option functions — `Summary`, `Description`, `Tag`, `Response`, `RequestBody`, `SchemaRef`, `SchemaInline`.
- [ ] `x/openapi/schema.go`: `Schema` and `SchemaRef` types; `APIErrorSchema()` convenience that returns the canonical `contract.APIError` shape.
- [ ] `x/openapi/spec_test.go`: table-driven tests — empty builder produces valid document, single GET route with no options produces operationId, route with `Response(200, SchemaRef("User"))` appears in components.schemas, two routes with same tag are grouped.
- [ ] `x/openapi/example_test.go`: runnable godoc example wiring `Builder` into an app after `app.Prepare()`.
- [ ] `cmd/plumego/commands/generate.go`: `plumego generate spec [--format json|yaml] [--output path]` — reads routes from a provided Go file that calls `app.Prepare()` and registers a spec handler at `/_plumego/openapi.json`.
- [ ] `x/openapi/module.yaml`: update `status` to `beta`.
- [ ] `specs/extension-maturity.yaml`: promote `x/openapi` to `beta`.
- [ ] `docs/concepts/extension-maturity.md`: update maturity table row for `x/openapi`.

### Phase 3 — Test (sequential)

9. Run `go test -race -timeout 60s ./x/openapi/...`.
10. Confirm generated JSON is valid OpenAPI 3.1.0 (manual schema check or simple structural assertion in test).
11. Run `go test -race -timeout 60s ./cmd/plumego/...`.

### Phase 4 — Validate and Ship (sequential)

12. Run the **Acceptance Criteria** commands; fix any failures.
13. Run `gofmt -w ./x/openapi/ ./cmd/plumego/` and confirm `gofmt -l` is empty.
14. Commit: `feat(x/openapi): promote to beta with Builder and generate spec CLI [M-014]`
15. Final commit: `milestone(M-014): OpenAPI 3.1 Generation`
16. Push to `milestone/M-014-openapi-generation`.

---

## Acceptance Criteria

```bash
go test -race -timeout 60s ./x/openapi/...
go test -race -timeout 60s ./cmd/plumego/...
go vet ./x/openapi/...
go vet ./cmd/plumego/...
gofmt -l ./x/openapi/
gofmt -l ./cmd/plumego/
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/public-entrypoints-sync
```

Expected: all exit 0; `gofmt -l` outputs nothing.

---

## Out of Scope

- Do not add OpenAPI support to stable root packages.
- Do not add struct tag scanning or reflection over handler types.
- Do not generate client SDKs in this milestone.
- Do not add an external OpenAPI validation library to `go.mod`.
- Do not change `router.RouteInfo` public API unless Phase 1 analysis reveals it is strictly necessary (document as Open Question first).

---

## Open Questions

(none at spec time — Phase 1 may surface a RouteInfo gap)

---

## Done Definition

- [ ] All Acceptance Criteria commands exit 0.
- [ ] Generated OpenAPI 3.1.0 JSON passes structural validation test.
- [ ] `x/openapi/module.yaml` status is `beta`.
- [ ] `specs/extension-maturity.yaml` reflects beta status.
- [ ] `docs/concepts/extension-maturity.md` maturity table updated.
- [ ] `plumego generate spec` command is documented in `cmd/plumego/README.md`.
- [ ] Branch `milestone/M-014-openapi-generation` pushed.
- [ ] PR open, title `milestone(M-014): OpenAPI 3.1 Generation`.
