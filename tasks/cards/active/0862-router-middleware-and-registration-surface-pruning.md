# Card 0862: Router Middleware And Registration Surface Pruning

Priority: P1
State: active
Primary Module: router

## Goal

Make `router` a pure route-structure package again: matching, params, groups, metadata needed for reverse routing, and one canonical registration path. Middleware composition must be owned by `core`/`middleware`, not by `router`.

## Problem

`router` currently imports the stable `middleware` package and owns a parallel middleware chain:

- `Router.Use(...)` stores middleware in router state.
- `MatchResult.RouteMiddlewares` exposes middleware through route matching.
- `dispatch.go` wraps handlers through `middleware.NewChain(...)`.
- route groups inherit middleware inside router internals.

This conflicts with the module manifest, which forbids `middleware/**` imports from `router`, and duplicates the canonical `core.App.Use(...)` middleware path.

`router` also exposes multiple overlapping registration styles:

- panic-oriented shortcuts: `Get`, `Post`, `Put`, `Delete`, `Patch`, `Options`, `Head`, `Any`
- named panic shortcuts: `GetNamed`, `PostNamed`, ...
- lower-level error-returning paths: `AddRoute`, `AddRouteWithName`, `AddRouteWithOptions`
- convenience wrappers such as `GroupFunc`
- exported method constants that shadow `net/http` method constants

Because compatibility is not required for this cleanup, the public surface should be reduced to one explicit route-registration path and the stdlib `http.Method*` names should be used by callers where possible.

## Scope

- Remove the `router -> middleware` dependency.
- Remove router-owned middleware state, inheritance, route middleware metadata, and handler wrapping.
- Keep app-level middleware behavior in `core.App.Use(...)` and `middleware.Chain(...)`.
- Collapse router registration to one canonical error-returning path.
- Remove panic shortcut wrappers and duplicate named shortcut wrappers in the same change that migrates their callers/tests.
- Replace exported HTTP method constants with `net/http` method constants at call sites where possible.
- Keep only an internal wildcard method sentinel if the `Any` route behavior still needs one.
- Update router tests to assert matching, params, groups, route names, reverse routing, and method handling through the canonical API.
- Update `core` route wiring to use the reduced router API.
- Update docs and module manifests so the listed public entrypoints match actual exported symbols.

## Non-Goals

- Do not change route matching semantics.
- Do not change path parameter extraction semantics.
- Do not remove reverse routing.
- Do not add feature-specific middleware behavior to `core`.
- Do not move route registration into `middleware`.

## Expected Files

- `router/router.go`
- `router/registration.go`
- `router/dispatch.go`
- `router/types.go`
- `router/group.go`
- `router/methods.go`
- `router/*_test.go`
- `core/routing.go`
- `core/*_test.go`
- `docs/modules/router/README.md`
- `router/module.yaml`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./router ./core ./x/devtools ./x/frontend ./x/websocket ./x/webhook
go test -race -timeout 60s ./router ./core
go vet ./router ./core ./x/devtools ./x/frontend ./x/websocket ./x/webhook
```

Then run the required repo-wide gates before committing.

## Done Definition

- `router` no longer imports `middleware`.
- No exported router middleware API remains.
- No router match result exposes middleware.
- No panic-only route-registration shortcut remains.
- The router public API has one canonical route-registration path.
- All old removed exported symbols have zero residual references.
- Focused gates and repo-wide gates pass.
