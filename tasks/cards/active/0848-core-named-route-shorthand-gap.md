# Card 0848

Priority: P2
State: active
Primary Module: core
Owned Files:
- `core/routing.go`
Depends On:

Goal:
- Close the gap between the HTTP method shorthand API and the named route registration API so that every method shorthand also has a `WithName` variant.

Problem:
- `core/routing.go` exposes two layers of route registration:
  1. HTTP method shorthands: `Get`, `Post`, `Put`, `Delete`, `Patch`, `Any` — convenient one-liners, take only `(path, handler)`.
  2. Named route registration: `AddRouteWithName(method, path, name, handler)` — verbose, requires passing the HTTP method as a string.
- There is no `GetWithName`, `PostWithName`, `PutWithName`, `DeleteWithName`, `PatchWithName`, or `AnyWithName` shorthand.
- As a result, any caller who wants to register a named route cannot use the shorthand API. They must fall back to `AddRouteWithName` and supply the HTTP method as a raw string (e.g. `"GET"`, `"POST"`), reintroducing the possibility of typos and bypassing the documented preferred API.
- The asymmetry is visible at a glance: the six shorthands (`routing.go:72-99`) call `addRoute`, which calls `registerRoute` with an empty name. None call `registerRoute` with a non-empty name. Named routes are effectively a second-class API.
- The `router` package already supports named routes (`router.AddRouteWithName`), so there is no implementation barrier.

Scope:
- Add six shorthand methods to `App`:
  `GetWithName(path, name string, handler http.Handler) error`
  `PostWithName(path, name string, handler http.Handler) error`
  `PutWithName(path, name string, handler http.Handler) error`
  `DeleteWithName(path, name string, handler http.Handler) error`
  `PatchWithName(path, name string, handler http.Handler) error`
  `AnyWithName(path, name string, handler http.Handler) error`
- Each delegates to `registerRoute` with the correct method constant and the supplied name.
- Add basic tests for each new method mirroring the existing `AddRouteWithName` test.

Non-goals:
- Do not remove `AddRouteWithName`; it remains for callers who need to select the method dynamically.
- Do not change the existing shorthand methods.

Files:
- `core/routing.go`
- `core/routing_test.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None required.

Done Definition:
- Six `*WithName` shorthand methods exist on `App`, each tested.
- All tests pass.

Outcome:
- Pending.
