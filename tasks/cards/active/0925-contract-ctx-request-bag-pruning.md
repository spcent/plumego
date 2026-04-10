# Card 0925: Contract Ctx Request-Bag Pruning

Priority: P1
State: active
Primary Module: contract

## Goal

Keep stable `contract` focused on transport contracts, response/error helpers, request metadata, and narrow binding helpers. Prune `Ctx` behavior that acts like a framework request bag, hidden lifecycle manager, or feature-specific response surface.

## Problem

`contract.Ctx` still aggregates many responsibilities:

- stores `http.ResponseWriter`, `*http.Request`, route params, query values, client IP, deadline, config, cached body, and cancel function
- creates request timeouts internally through `RequestConfig.RequestTimeout`
- owns response shortcuts: `Response`, `Text`, `Bytes`, `Redirect`, `UnsafeRedirect`, `File`, `SetCookie`, `Cookie`
- owns body caching and query binding
- owns streaming dispatch through `StreamConfig`, `StreamFormat`, `SSEEvent`, `SSEWriter`, and `Ctx.Stream`
- supports a non-canonical handler shape through `CtxHandlerFunc` and `AdaptCtxHandler`

This conflicts with current style guidance:

- canonical handlers are `func(http.ResponseWriter, *http.Request)`
- data source should be visible at the read site
- no framework-style request bags or hidden request lifecycle behavior
- one canonical success-response path per layer

The goal is not to remove every convenience blindly, but to make the stable contract surface small and explicit.

## Scope

- Enumerate all exported `Ctx` and streaming symbols before editing.
- Decide which pieces remain stable primitives and which should be removed or moved to extension/application helpers.
- Remove hidden request timeout creation from `NewCtx`/`RequestConfig` if `Ctx` remains.
- Remove or de-emphasize non-canonical `CtxHandlerFunc`/`AdaptCtxHandler`.
- Collapse response shortcuts onto `WriteJSON`, `WriteResponse`, and `WriteError` where possible.
- Keep direct request metadata accessors (`WithRequestContext`, `RequestContextFromContext`, route pattern/name helpers) if still needed by router/middleware.
- Update tests and docs together with any exported symbol removals.

## Non-Goals

- Do not change `WriteError` error envelope shape.
- Do not remove `WriteJSON` or `WriteResponse`.
- Do not introduce middleware-based DTO binding.
- Do not add service lookup or app state to context.
- Do not move protocol-specific streaming policy into stable `contract` unless explicitly justified.

## Expected Files

- `contract/context_core.go`
- `contract/context_response.go`
- `contract/context_bind.go`
- `contract/context_stream.go`
- `contract/*_test.go`
- `docs/modules/contract/README.md`
- `contract/module.yaml`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./contract ./router ./middleware/...
go test -race -timeout 60s ./contract
go vet ./contract ./router ./middleware/...
```

Then run the required repo-wide gates before committing.

## Done Definition

- `contract` no longer exposes framework-style request-bag behavior beyond explicitly retained primitives.
- Hidden request timeout/lifecycle behavior is removed from stable `contract`.
- Success and error response paths remain canonical and documented.
- Removed exported symbols have zero residual references.
- Focused gates and repo-wide gates pass.
