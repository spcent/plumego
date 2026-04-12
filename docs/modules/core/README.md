# core

## Purpose

`core` is the HTTP application kernel. It owns app construction, route attachment, middleware attachment, and server lifecycle.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- assembling an application from explicit routes and middleware
- starting or stopping an HTTP server
- wiring shared runtime facilities such as logging and metrics

## Do not use this module for

- route matching
- feature catalogs or plugin containers
- tenant policy
- persistence behavior

## First files to read

- `core/module.yaml`
- `core/app.go`
- `core/options.go`
- `reference/standard-service/internal/app/app.go`

## Public entrypoints

- `New`
- `App`
- `AppDependencies`
- `(*App).Prepare`
- `(*App).Server`
- `(*App).Shutdown`

## Main risks when changing this module

- startup regression
- shutdown regression
- route assembly regression

## Canonical change shape

- keep bootstrap explicit
- construct apps through `DefaultConfig()` / `AppConfig`, and use `AppDependencies` as the explicit constructor dependency contract
- keep lifecycle behavior reviewable
- keep one canonical lifecycle path: `Prepare` + `Server` + `Shutdown`
- split handler preparation from server preparation: `ServeHTTP` only freezes config/router state and builds the handler, while `Prepare` is the explicit path that allocates `http.Server` and prepares connection tracking
- keep TLS on the same public serve path: `Prepare` loads configured certificates into the returned `*http.Server`, and callers use `ListenAndServeTLS("", "")` on that prepared server when TLS is enabled
- keep `core` as the first-party router owner: route wiring goes through `App.AddRoute` / `App.Get` / `App.Post`; named routes use `App.AddRoute(..., router.WithRouteName(...))`; reverse URL lookup goes through `App.URL(...)`, not raw router replacement or mutation
- keep HTTP request metrics explicit in middleware/app-local wiring; `core` does not own live observer attachment state
- keep readiness ownership out of `core`; callers own the outer serve loop, so readiness signaling must stay app-local instead of pretending the kernel knows when traffic can flow
- keep logger subsystem ownership out of `core`; `core.AppDependencies{Logger: ...}` injects a passive dependency and callers own logger initialization and flushing
- keep app-local debug flags and env-file metadata outside `core`; the kernel owns HTTP runtime state, not devtools metadata transport
- keep debug/runtime snapshot payloads in `x/devtools`; `core` owns lifecycle state, not tooling response contracts
- keep the app logger kernel-owned on `App.Logger()`; `router` stays logger-free and does not mirror app logger state
- keep router behavior policy in typed config, not in ad hoc constructor options
- push feature-specific wiring back to app-local code or the owning extension
- preserve `net/http` compatibility while keeping `core` as a kernel
