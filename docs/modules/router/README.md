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
- logger lookup or logger carriage
- service construction
- middleware ownership or middleware execution policy
- frontend asset policy such as cache headers, SPA fallback, precompressed assets, or ETag generation

## First files to read

- `router/module.yaml`
- `router/router.go`
- `router/registration.go`

## Public entrypoints

- `Router`
- `Group`
- `AddRoute`
- `Param`
- `Static`
- `StaticFS`

## Main risks when changing this module

- dispatch regression
- param extraction regression
- reverse routing regression

## Canonical change shape

- preserve deterministic dispatch
- keep explicit method-plus-path registration behavior
- keep router APIs stdlib-shaped instead of alias-heavy
- use `net/http` method constants for standard HTTP methods
- keep one public param helper (`Param`)
- keep route metadata attached through `AddRoute(..., WithRouteName(...))`
- avoid bleeding response or middleware policy into router internals
- keep static serving as a small file-mount primitive; use `x/frontend` for frontend asset policy

## Boundary notes

- `router` does not own middleware registration or middleware chains.
- App-wide middleware belongs to `core.App.Use(...)` and the stable `middleware` package.
- `router` keeps an internal ANY sentinel for wildcard method dispatch; callers should prefer `core.App.Any(...)` for app-level catch-all routes.
- `router.Static` and `router.StaticFS` are primitive GET mounts. Cache headers, SPA fallback, precompressed files, custom headers, and MIME policy belong to `x/frontend`.
