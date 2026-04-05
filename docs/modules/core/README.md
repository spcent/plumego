# core

## Purpose

`core` is the HTTP application kernel. It owns app construction, route attachment, middleware attachment, and server lifecycle.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- assembling an application from explicit routes and middleware
- starting or stopping an HTTP server
- wiring shared runtime facilities such as logging, health, and metrics

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
- `Option`
- `(*App).Prepare`
- `(*App).Start`
- `(*App).Server`
- `(*App).Shutdown`

## Main risks when changing this module

- startup regression
- shutdown regression
- route assembly regression

## Canonical change shape

- keep bootstrap explicit
- construct apps through `DefaultConfig()` / `AppConfig` plus non-config `Option`s, instead of per-field config helper options
- keep lifecycle behavior reviewable
- keep one canonical lifecycle path: `Prepare` + `Start` + `Server` + `Shutdown`
- treat `Prepare` and first-use `ServeHTTP` as the same one-time preparation boundary: both freeze config/router state and make the app serve-ready through the same internal transition
- keep server preparation and runtime snapshots aligned through one shared internal projection, instead of duplicating field-by-field remaps across lifecycle helpers
- keep TLS on the same public serve path: `Prepare` loads configured certificates into the returned `*http.Server`, and callers use `ListenAndServeTLS("", "")` on that prepared server when TLS is enabled
- keep `core` as the first-party router owner: route wiring goes through `App.AddRoute` / `App.Get` / `App.Post` and reverse URL lookup goes through `App.URL(...)`, not raw router replacement or mutation
- attach HTTP metrics observers through `AttachHTTPObserver` when app-local wiring needs fan-out metrics collection
- keep app-local debug flags and env-file metadata outside `core`; the kernel owns HTTP runtime state, not devtools metadata transport
- keep the app logger kernel-owned on `App.Logger()`; `core` does not mirror it into router state
- treat router-affecting options as declarative app construction input, not eager side effects
- push feature-specific wiring back to app-local code or the owning extension
- preserve `net/http` compatibility while keeping `core` as a kernel
