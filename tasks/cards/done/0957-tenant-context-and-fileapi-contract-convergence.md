# Card 0957: Tenant Context And File API Contract Convergence

Priority: P1
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/tenant
Depends On: —

## Goal

Normalize tenant context access onto the repository's canonical
`With{Type}`/`{Type}FromContext` shape and make `x/fileapi` consume that
contract directly instead of relying on raw string-key context lookups and a
duplicated route table.

## Problem

- `x/tenant/core/context.go` still exports `ContextWithTenantID`, which breaks
  the repository-wide accessor naming rule that already applies in `contract`
  and `security`.
- `x/fileapi/handler.go` reads tenant and user identity from
  `ctx.Value("tenant_id")` / `ctx.Value("user_id")`, which is a hidden string
  key convention instead of an explicit accessor contract.
- `x/fileapi/module.yaml` and `docs/modules/x-fileapi/README.md` currently
  teach callers to inject `"tenant_id"` into context, so the drift is encoded
  in the control plane as well as the code.
- `x/fileapi/handler.go` exports both `RegisterRoutes(*http.ServeMux)` and
  `ServeHTTP`, which duplicates the route table in one package and makes route
  ownership less explicit than the repo's preferred wiring style.

## Scope

- Rename tenant context attachment to `WithTenantID` and migrate all callers in
  the same change; remove `ContextWithTenantID` after the last caller moves.
- Keep or refine `RequestWithTenantID` only if it still adds value after the
  canonical accessor pair exists.
- Replace raw string-key tenant lookup in `x/fileapi` with the `x/tenant`
  accessor.
- Replace raw string-key user lookup in `x/fileapi` with an explicit contract:
  either a local typed accessor or an injected resolver. Do not keep another
  undocumented string-key convention.
- Choose one canonical route surface for `x/fileapi` (`ServeHTTP` or explicit
  app-local route wiring) instead of maintaining both route tables forever.
- Update docs and tests so the new context contract is the only documented path.

## Non-Goals

- Do not move tenant logic back into stable `middleware` or stable `store`.
- Do not redesign `x/data/file` storage behavior.
- Do not add feature-specific user/session policy into `x/fileapi`.

## Expected Files

- `x/tenant/core/context.go`
- `x/tenant/core/context_test.go`
- `x/fileapi/handler.go`
- `x/fileapi/handler_test.go`
- `x/fileapi/module.yaml`
- `docs/modules/x-fileapi/README.md`

## Validation

```bash
rg -n 'ContextWithTenantID|Value\("tenant_id"\)|Value\("user_id"\)|RegisterRoutes\(mux \*http\.ServeMux\)' x/tenant x/fileapi . -g '*.go'
go test -timeout 20s ./x/tenant/... ./x/fileapi/...
go vet ./x/tenant/... ./x/fileapi/...
```

## Docs Sync

- `x/fileapi/module.yaml`
- `docs/modules/x-fileapi/README.md`
- any tenant context examples that still teach `ContextWithTenantID`

## Done Definition

- Tenant context attachment follows `WithTenantID` + `TenantIDFromContext`.
- `ContextWithTenantID` has zero remaining references.
- `x/fileapi` no longer reads tenant or user identity via raw string-key
  `context.Value` lookups.
- `x/fileapi` has one canonical route-definition surface instead of duplicated
  `RegisterRoutes` and `ServeHTTP` tables.
- Focused `x/tenant` and `x/fileapi` tests pass.

## Outcome

- Renamed tenant context attachment to `WithTenantID` and migrated all in-repo callers.
- Added `x/fileapi.WithUserID` and `UserIDFromContext` so file handlers no longer depend on raw string-key context values.
- Removed `x/fileapi`'s duplicated `RegisterRoutes` and `ServeHTTP` surfaces in favor of explicit app-local route wiring.
- Updated `x/fileapi` docs and module manifest so the control plane now teaches the same explicit context contract as the code.

## Validation Run

```bash
rg -n 'ContextWithTenantID|Value\("tenant_id"\)|Value\("user_id"\)|RegisterRoutes\(mux \*http\.ServeMux\)|func \(h \*Handler\) ServeHTTP' x/tenant x/fileapi . -g '*.go'
go test -timeout 20s ./x/tenant/... ./x/fileapi/...
go vet ./x/tenant/... ./x/fileapi/...
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```
