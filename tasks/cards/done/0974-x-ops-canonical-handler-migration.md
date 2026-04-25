# Card 0974: X/Ops Canonical Handler Migration

Priority: P2
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/ops
Depends On: 0971

## Goal

Migrate all `x/ops` HTTP handlers from the local `adaptCtx`/`*contract.Ctx`
pattern to the canonical `func(w http.ResponseWriter, r *http.Request)` shape
and remove the package-local `adaptCtx` adapter.

## Problem

`x/ops/ops.go` registers six route handlers through a package-local
`adaptCtx` shim (line 474) that wraps `func(*contract.Ctx)` handlers into
`http.Handler`. The canonical handler shape for this repository is
`func(w http.ResponseWriter, r *http.Request)` (style guide §canonical defaults).
The `*contract.Ctx` shape was an earlier abstraction that has since been retired
from the stable layer; keeping it in `x/ops` means:

- `x/ops` maintains a redundant adapter that duplicates similar adapters in
  x/messaging, x/rest, x/devtools, and cmd/devserver.
- Bare `contract.WriteError(ctx.W, ctx.R, ...)` calls in the handlers cannot be
  fixed without `_ =` without first fixing the receiver type (the `ctx.*`
  accessors obscure the pattern).
- Seven handler functions (`handleSummary`, `handleQueueStats`,
  `handleQueueReplay`, `handleReceiptLookup`, `handleChannelHealth`,
  `handleTenantQuota`, plus helpers `writeHookError` and `writeNotImplemented`)
  accept `*contract.Ctx` instead of the canonical `(w, r)` pair.

## Scope

- Change all six handler methods (`handleSummary`, `handleQueueStats`,
  `handleQueueReplay`, `handleReceiptLookup`, `handleChannelHealth`,
  `handleTenantQuota`) to accept `(w http.ResponseWriter, r *http.Request)`.
- Update `writeHookError` and `writeNotImplemented` helpers to accept
  `(w, r)` instead of `*contract.Ctx`.
- Replace all `ctx.W`, `ctx.R`, `ctx.RequestID()`, `ctx.RequestHeaders()`,
  `ctx.PathParam()`, `ctx.BodyBytes()`, `ctx.QueryParam()` accesses with
  equivalent `httpx`, `contract`, or standard-library calls on `w` and `r`.
- Prefix every `contract.WriteError(...)` call in the rewritten handlers with
  `_ =` per style guide §17.3.
- Remove the `adaptCtx` function and the `contract.Ctx` import once no callers
  remain.
- Update route registration to pass handlers directly (no `adaptCtx` wrapping).

## Non-Goals

- Do not change route paths, HTTP methods, auth middleware wiring, or response
  body content.
- Do not change `x/ops/healthhttp/` handlers (those are already on the
  canonical `(w, r)` shape and covered by card 0971 for the `_ =` fix).
- Do not restructure the `Options` type or `RegisterRoutes` signature.

## Files

- `x/ops/ops.go`
- `x/ops/ops_test.go` (update any test helpers that use `*contract.Ctx`)

## Tests

```bash
rg -n 'adaptCtx|contract\.Ctx' x/ops -g '*.go'
go test -timeout 20s ./x/ops/...
go vet ./x/ops/...
```

## Docs Sync

- `docs/modules/x-ops/README.md` if it shows handler signatures or
  `adaptCtx` usage.

## Done Definition

- `rg 'adaptCtx|contract\.Ctx' x/ops -g '*.go'` returns no results.
- All six handlers accept `(w http.ResponseWriter, r *http.Request)`.
- All `contract.WriteError` calls in ops.go use `_ =` prefix.
- `x/ops` tests pass; `go vet` is clean.

## Outcome

