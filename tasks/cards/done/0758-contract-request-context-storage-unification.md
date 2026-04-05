# Card 0758

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/context_core.go`
- `contract/context_test.go`
- `contract/context_extended_test.go`
- `router/dispatch.go`
- `router/params.go`
- `router/static.go`
- `router/router_contract_test.go`
- `x/rest/resource.go`
- `middleware/accesslog/accesslog.go`

Goal:
- Store request routing metadata through one canonical context payload instead of
  preserving both aggregate and legacy param-only storage models.

Problem:
- `contract` currently supports `WithRequestContext` / `RequestContextFromContext`
  and a second param-only path through `WithParams` / `ParamsFromContext`.
- `RequestContextFromContext` falls back to `ParamsFromContext`, so callers can
  still rely on the legacy partial payload.
- Router and middleware consumers now read a mix of `RequestContext` and
  `ParamsFromContext`, which keeps two storage contracts alive inside the
  transport layer.
- This makes route metadata propagation harder to reason about and keeps
  compatibility-shaped branching in the stable path.

Scope:
- Choose one canonical request metadata payload for context propagation.
- Remove the duplicate storage/access path once router and middleware callers
  are migrated.
- Keep direct helpers for route pattern / route name only if they are thin views
  over the canonical aggregate payload.

Non-goals:
- Do not redesign router matching semantics.
- Do not add business data to the request context payload.
- Do not add new middleware-only context wrappers.

Files:
- `contract/context_core.go`
- `contract/context_test.go`
- `contract/context_extended_test.go`
- `router/dispatch.go`
- `router/params.go`
- `router/static.go`
- `router/router_contract_test.go`
- `x/rest/resource.go`
- `middleware/accesslog/accesslog.go`

Tests:
- Add or update tests so router dispatch and downstream consumers depend on the
  single canonical request metadata payload.
- `go test -race -timeout 60s ./contract/...`
- `go test -timeout 20s ./router/... ./middleware/accesslog/... ./x/rest/...`

Docs Sync:
- Update any contract or router docs that still imply dual params/request-context
  storage.

Done Definition:
- Request routing metadata is stored through one canonical context payload.
- Legacy param-only context helpers are removed.
- Router and downstream consumers read the same canonical payload.

Outcome:
- Removed the legacy param-only context storage path from `contract` by deleting
  `WithParams` and `ParamsFromContext`.
- Standardized request routing metadata on the aggregate `RequestContext`
  payload, with router and downstream readers now using
  `RequestContextFromContext(...).Params`.
- Updated router/static-file handling and contract/router tests so they no
  longer rely on fallback behavior or legacy key types.
- Validation:
  - `go test -race -timeout 60s ./contract/...`
  - `go test -timeout 20s ./router/... ./middleware/accesslog/... ./x/rest/...`
  - `go vet ./contract/... ./router/... ./middleware/accesslog/... ./x/rest/...`
