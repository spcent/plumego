# Card 0975: X/Messaging and X/Rest ContextResourceController Handler Migration

Priority: P2
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/messaging, x/rest
Depends On: 0974

## Goal

Migrate `x/messaging` HTTP handlers and the `x/rest` `ContextResourceController`
family from the local `adaptCtx`/`*contract.Ctx` pattern to the canonical
`func(w http.ResponseWriter, r *http.Request)` signature, then remove the
package-local `adaptCtx` adapters.

## Problem

### x/messaging

`x/messaging/routes.go` defines a local `adaptCtx` function (line 39) that
wraps `func(*contract.Ctx)` into `http.Handler`. All six exported handler
methods on `Service` (`HandleSend`, `HandleBatchSend`, `HandleStats`,
`HandleGetReceipt`, `HandleListReceipts`, `HandleChannelHealth`) and the
private `writeServiceError` helper accept `*contract.Ctx` instead of the
canonical `(w, r)` pair. This duplicates the adapter pattern present in x/ops,
x/rest, x/devtools, and cmd/devserver.

### x/rest ContextResourceController

`x/rest/resource.go` exports a `ContextResourceController` interface (line 447)
with eight Ctx-suffixed methods, a `BaseContextResourceController` struct (line
458) providing default not-implemented stubs, and `x/rest/entrypoints.go` wires
these through a local `adaptCtx`. `x/rest/resource_db.go` implements all eight
`*Ctx` methods on `DBResourceController[T]`. The canonical `ResourceController`
interface (line 22) already uses `(w, r)`, so `ContextResourceController` is an
unnecessary parallel surface.

## Scope

### x/messaging

- Change all six `Handle*` methods on `Service` to accept
  `(w http.ResponseWriter, r *http.Request)`.
- Update `writeServiceError` to accept `(w, r)` instead of `*contract.Ctx`.
- Replace all `ctx.W`, `ctx.R`, `ctx.PathParam()`, `ctx.BodyBytes()`,
  `ctx.QueryParam()`, `ctx.RequestID()` accesses with equivalent calls.
- Prefix every `contract.WriteError(...)` with `_ =`.
- Remove the `adaptCtx` function from `routes.go`; register handlers directly.

### x/rest

- Delete the `ContextResourceController` interface, `BaseContextResourceController`
  struct, and `adaptCtx` from `entrypoints.go`.
- Migrate the eight `*Ctx` methods on `DBResourceController[T]` to their
  canonical `(w, r)` names (drop the `Ctx` suffix), implementing the existing
  `ResourceController` interface directly.
- Remove the `RegisterContextRoutes` entrypoint if it exclusively served
  `ContextResourceController`; keep `RegisterRoutes` which targets
  `ResourceController`.
- Update all callers in tests or examples that use `ContextResourceController`
  or `Register*Context*`.

## Non-Goals

- Do not change route paths, HTTP methods, or response content.
- Do not change `Service` constructor, store interface, or receipt logic in
  x/messaging.
- Do not change the canonical `ResourceController` interface or
  `BaseResourceController` in x/rest.
- Do not migrate x/devtools or cmd/devserver adapters (covered by card 0976).

## Files

- `x/messaging/api.go`
- `x/messaging/routes.go`
- `x/messaging/*_test.go`
- `x/rest/resource.go` (remove ContextResourceController, BaseContextResourceController)
- `x/rest/resource_db.go` (migrate *Ctx methods to (w, r) shape)
- `x/rest/entrypoints.go` (remove adaptCtx, RegisterContextRoutes)
- `x/rest/*_test.go`

## Tests

```bash
rg -n 'adaptCtx|contract\.Ctx|ContextResourceController|RegisterContextRoutes|\.IndexCtx|\.ShowCtx|\.CreateCtx|\.UpdateCtx|\.DeleteCtx|\.PatchCtx|\.BatchCreateCtx|\.BatchDeleteCtx' x/messaging x/rest -g '*.go'
go test -timeout 20s ./x/messaging/... ./x/rest/...
go vet ./x/messaging/... ./x/rest/...
```

## Docs Sync

- `docs/modules/x-messaging/README.md` if it shows `HandleSend` or handler
  signatures.
- `docs/modules/x-rest/README.md` if it documents `ContextResourceController`
  or `RegisterContextRoutes`.

## Done Definition

- `rg 'adaptCtx|contract\.Ctx' x/messaging x/rest -g '*.go'` returns no
  results.
- `ContextResourceController`, `BaseContextResourceController`, and all `*Ctx`
  method names are absent from x/rest source.
- `DBResourceController[T]` implements `ResourceController` directly via
  canonical `(w, r)` methods.
- All x/messaging and x/rest tests pass; `go vet` is clean.

## Outcome

