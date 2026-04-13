# Card 0978: X/Webhook Outbound and Inbound Handler Migration

Priority: P2
State: active
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/webhook
Depends On: 0973

## Goal

Migrate all `x/webhook` HTTP handlers in `in.go` and `out.go` from the
package-local `adaptCtx`/`*contract.Ctx` pattern to the canonical
`func(w http.ResponseWriter, r *http.Request)` shape, then delete
`x/webhook/ctx_adapter.go`.

## Problem

Card 0969 cleaned up `x/webhook`'s `routesOnce` guards and `err*` helper
families, but left the `adaptCtx`/`*contract.Ctx` handler shape in place.
`x/webhook/ctx_adapter.go` is a dedicated shim file whose sole purpose is to
define `adaptCtx`. The inbound and outbound handlers still use it:

**`in.go`** — 2 methods accepting `*contract.Ctx`:
- `webhookInGitHub(ctx *contract.Ctx)` (line 73)
- `webhookInStripe(ctx *contract.Ctx)` (line 170)

**`out.go`** — 9 package-level handler functions accepting `*contract.Ctx`:
- `webhookCreateTarget`, `webhookListTargets`, `webhookGetTarget`,
  `webhookPatchTarget`, `webhookSetTargetEnabled`, `webhookTriggerEvent`,
  `webhookListDeliveries`, `webhookGetDelivery`, `webhookReplayDelivery`

All are wrapped at route registration time with `adaptCtx(func(ctx *contract.Ctx) { ... })`.

## Scope

- Change `webhookInGitHub` and `webhookInStripe` in `in.go` to accept
  `(w http.ResponseWriter, r *http.Request)` and register them directly (no
  `adaptCtx` wrapper).
- Change all nine `webhook*` functions in `out.go` to accept
  `(w http.ResponseWriter, r *http.Request, svc *Service)` (or equivalent
  closure form) and register them directly.
- Replace all `ctx.W`, `ctx.R`, `ctx.RequestID()`, `ctx.RequestHeaders()`,
  `ctx.PathParam()`, `ctx.BodyBytes()`, `ctx.QueryParam()` accesses with
  equivalent calls on `w` and `r`.
- Prefix every `contract.WriteError(...)` call in migrated handlers with `_ =`
  per style guide §17.3.
- Delete `x/webhook/ctx_adapter.go`.
- Remove the `contract.Ctx` import once no remaining callers exist.

## Non-Goals

- Do not change route paths, HTTP methods, authentication logic, or signature
  verification behavior.
- Do not change the outbound webhook store, delivery retry, or event-trigger
  semantics.
- Do not modify `inbound_errors.go`, `inbound_hmac.go`, or verification logic.

## Files

- `x/webhook/in.go`
- `x/webhook/out.go`
- `x/webhook/ctx_adapter.go` (delete)
- `x/webhook/inbound_test.go`, `x/webhook/outbound_test.go` (update as needed)

## Tests

```bash
rg -n 'adaptCtx\|contract\.Ctx' x/webhook -g '*.go'
go test -timeout 20s ./x/webhook/...
go vet ./x/webhook/...
```

## Docs Sync

- `docs/modules/x-webhook/README.md` if it shows handler registration or
  function signatures.

## Done Definition

- `rg 'adaptCtx\|contract\.Ctx' x/webhook -g '*.go'` returns no results.
- `x/webhook/ctx_adapter.go` no longer exists.
- All webhook handler functions accept `(w http.ResponseWriter, r *http.Request)`.
- All `contract.WriteError` calls in in.go and out.go use `_ =` prefix.
- `x/webhook` tests pass; `go vet` is clean.

## Outcome

