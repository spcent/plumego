# Card 0273: Core Route Registration Surface Convergence

Priority: P1
State: done
Primary Module: core

## Goal

Keep `core` route registration to one canonical route-registration path plus the existing method helpers. Remove duplicate named-route entrypoints now that `router.RouteOption` supports naming directly.

## Problem

`core/routing.go` currently exposes two public ways to register named routes:

- `AddRoute(method, path, handler, router.WithRouteName("name"))`
- `AddRouteWithName(method, path, name, handler)`

This duplicates the router-owned route metadata path and keeps `core` carrying a named-route convenience wrapper that is no longer needed. The stable route style should remain:

- one method + one path + one handler per line
- optional route metadata through `router.RouteOption`
- no per-feature aliases for the same registration behavior

The package also has internal `addRoute` and `registerRoute` layering that should remain only if it still helps after the duplicate public entrypoint is removed.

## Scope

- Enumerate all references to `AddRouteWithName` before editing.
- Migrate callers/tests to `AddRoute(..., router.WithRouteName(...))`.
- Remove `AddRouteWithName`.
- Keep `AddRoute`, `Get`, `Post`, `Put`, `Delete`, `Patch`, and `Any` unless analysis shows further duplication that can be safely collapsed.
- Keep `URL` as the app-owned access to reverse routing if still used.
- Update docs and tests for the canonical named-route registration path.

## Non-Goals

- Do not remove method helpers unless they are proven unused and redundant in the same implementation pass.
- Do not change router reverse-routing behavior.
- Do not add new per-method named route helpers.
- Do not change stable `net/http` handler compatibility.

## Expected Files

- `core/routing.go`
- `core/routing_test.go`
- `docs/modules/core/README.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
- affected reference or x/* callers if any

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./core ./router
go test -race -timeout 60s ./core ./router
go vet ./core ./router
```

Then run the required repo-wide gates before committing.

## Done Definition

- `core.App` has one canonical named route registration path through `AddRoute` + `router.WithRouteName`.
- `AddRouteWithName` has zero residual references.
- Core route registration tests cover named route registration through the canonical path.
- Docs no longer advertise the removed duplicate entrypoint.
- Focused gates and repo-wide gates pass.

## Outcome

- Removed `core.App.AddRouteWithName`.
- Kept named route registration on the canonical `App.AddRoute(..., router.WithRouteName(...))` path.
- Simplified the internal route registration implementation by removing the duplicate name parameter.
- Updated current core/style documentation and tests to use `router.WithRouteName`.
