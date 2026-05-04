# router

## Purpose

`router` owns route matching, params, groups, and reverse routing.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- adding matching or grouping behavior
- changing path parameter extraction
- working on reverse routing or mount primitives
- registering routes directly with `Router.AddRoute(method, path, handler, opts...)`
- mounting a small static directory or `http.FileSystem` without frontend asset policy

## Do not use this module for

- JSON responses
- auth policy
- route parameter validation policy
- logger lookup or logger carriage
- service construction
- middleware ownership or middleware execution policy
- frontend asset policy such as cache headers, SPA fallback, precompressed assets, or ETag generation

## First files to read

- `router/module.yaml`
- `router/router.go`
- `router/registration.go`

## Public entrypoints

- `NewRouter`
- `MethodAny`
- `Router`
- `RouterOption`
- `RouteOption`
- `RouteMeta`
- `RouteInfo`
- `NamedRoute`
- `Group`
- `AddRoute`
- `Freeze`
- `HasRoute`
- `MethodNotAllowedEnabled`
- `NamedRoutes`
- `Param`
- `Print`
- `Routes`
- `ServeHTTP`
- `SetMethodNotAllowed`
- `URL`
- `URLMust`
- `WithRouteName`
- `WithMethodNotAllowed`
- `Static`
- `StaticFS`

## Main risks when changing this module

- dispatch regression
- param extraction regression
- reverse routing regression
- lifecycle ambiguity between mutable registration and immutable serving

## Canonical change shape

- preserve deterministic dispatch
- keep explicit method-plus-path registration behavior
- normalize registered route paths to leading-slash patterns, including grouped
  child routes that omit the leading slash; repeated leading slashes collapse
  to one stored leading slash while internal empty segments remain invalid
- reject empty or whitespace-containing methods, and keep route parameter names
  to ASCII identifiers such as `:id`, `:userID`, or `*rest_path`
- treat route registration failures as returned `error` values; do not model
  duplicate/conflict/frozen registration through panic-oriented helpers
- keep router APIs stdlib-shaped instead of alias-heavy
- use `net/http` method constants for standard HTTP methods
- keep one public param helper (`Param`)
- keep route metadata attached through `AddRoute(..., WithRouteName(...))`
- avoid bleeding response or middleware policy into router internals
- keep static serving as a small file-mount primitive; use `x/frontend` for frontend asset policy

## Lifecycle Contract

Direct `router.Router` callers own the build-and-serve boundary:

- register routes before serving traffic
- call `Freeze()` before serving when the runtime route table must be immutable
- do not call `AddRoute`, `Static`, `StaticFS`, or group registration while
  requests are concurrently being served
- treat registration after `Freeze()` as a normal returned error, not a panic
  path

`core.App` owns this boundary for app-managed routers. `core.App.Prepare()` and
the first `core.App.ServeHTTP(...)` path freeze the owned router before building
the handler chain, so callers using `core` do not call `Router.Freeze()`
directly.

## Boundary notes

- `router` does not own middleware registration or middleware chains.
- App-wide middleware belongs to `core.App.Use(...)` and the stable `middleware` package.
- `router.MethodAny` is the reserved fallback method sentinel for wildcard method dispatch; it is not available as a separate exact custom HTTP method.
- Callers should prefer `core.App.Any(...)` for app-level catch-all routes.
- HEAD requests fall back to matching GET handlers and suppress response bodies while preserving handler-visible write counts.
- Route misses use the standard-library `http.NotFound` plain-text response.
- When 405 handling is enabled, method mismatches use `contract.WriteError`
  with `method_not_allowed` and a sorted `Allow` header.
- Named route collisions fail registration; route names must be unique.
- `URL` consumes params as key/value pairs, percent-escapes segment params, preserves slash boundaries for wildcard params, returns empty string for unknown or incomplete routes, and ignores an unpaired trailing param key.
- `router.Static` and `router.StaticFS` are primitive GET mounts. Cache headers, SPA fallback, precompressed files, custom headers, and MIME policy belong to `x/frontend`.
- Static request paths are cleaned with slash-based URL semantics before local
  directory serving converts them to platform filesystem paths.
- Static prefixes are canonicalized before registration: relative prefixes gain
  a leading slash, trailing slashes are removed, and root mounts register as
  `/*filepath`.
- `Static` resolves and validates its local root during registration; missing
  roots and file roots fail fast.
- For embedded directories, pass a filesystem rooted at the mounted directory
  to `StaticFS`, for example `sub, _ := fs.Sub(public, "public")` followed by
  `r.StaticFS("/assets", http.FS(sub))`.

## Frozen behavior matrix

These behaviors are part of the current stable-root freeze baseline:

| Surface | Behavior |
| --- | --- |
| Registration | one method, one normalized path, one handler per route |
| Relative paths | route and group paths gain one leading slash, repeated leading slashes collapse, and internal double slashes remain invalid |
| Params | `Param(r, name)` and `contract.RequestContextFromContext` expose matched params |
| Groups | nested groups compose normalized prefixes and preserve named route metadata |
| Matching | static segments take precedence over params, and params take precedence over wildcards; warm cache preserves that result |
| Reverse routing | `URL` percent-escapes params and returns empty string for unknown or missing params |
| Route snapshots | `Routes` returns method/path-sorted route metadata snapshots |
| 404 handling | unmatched routes use standard-library `http.NotFound` |
| 405 handling | disabled by default; when enabled, returns sorted `Allow` and canonical `contract` error body |
| HEAD fallback | HEAD can use matching GET handlers while suppressing response body writes |
| Freeze | Direct router users call `Freeze` before immutable serving; `core.App` freezes owned routers during prepare/first serve |
| Static mounts | `Static` and `StaticFS` are small GET file mounts, not frontend asset policy |

Focused regression coverage lives in `router/freeze_test.go`,
`router/router_contract_test.go`, `router/reverse_routing_group_test.go`, and
`router/static_test.go`. Lightweight seed-based fuzz coverage for path
normalization and reverse routing lives in `router/fuzz_test.go` and runs under
normal `go test`.
