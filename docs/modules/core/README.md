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
- `core/config.go`
- `core/routing.go`
- `core/middleware.go`
- `core/lifecycle.go`
- `core/options.go`
- `reference/standard-service/internal/app/app.go`

## Public entrypoints

- `New`
- `App`
- `AppDependencies`
- `DefaultConfig`
- `AppConfig`
- `TLSConfig`
- `RouterConfig`
- `PreparationState`
- `(*App).Use`
- `(*App).AddRoute`
- `(*App).Get`
- `(*App).Post`
- `(*App).Put`
- `(*App).Delete`
- `(*App).Patch`
- `(*App).Any`
- `(*App).URL`
- `(*App).Routes`
- `(*App).Logger`
- `(*App).Prepare`
- `(*App).Server`
- `(*App).Shutdown`

## Configuration contract

`DefaultConfig()` is the supported baseline for application code. Custom
`AppConfig` values are copied by value during `New`, so later mutations to the
caller-owned config value do not affect the app.

Server hardening fields follow standard `http.Server` zero-value semantics:
zero timeouts disable that specific timeout and `MaxHeaderBytes == 0` leaves the
server on the standard library default. `Prepare` rejects empty addresses and
negative timeout/header-size values before freezing route or middleware
mutation.

TLS ownership is intentionally narrow. `Prepare` loads configured certificate
and key material into the prepared `*http.Server`; callers own advanced TLS
policy such as minimum versions, cipher suites, client certificates, and custom
`GetCertificate` behavior by adjusting `Server().TLSConfig` before serving.

## Error contract

Core lifecycle and wiring entrypoints return normal Go errors. Errors use a
canonical `core <operation>:` prefix, include stable operation names such as
`prepare_server`, `get_server`, `shutdown_app`, `add_route`, and
`use_middleware`, and wrap lower-level causes where a caller-owned dependency
or the standard library produced the failure.

The stable core surface intentionally does not export lifecycle sentinel errors.
Callers should handle errors at the operation boundary, use `errors.Is` only for
wrapped lower-level causes that are already exported by the producing package,
and avoid depending on ad hoc string parsing for branching behavior.

## Main risks when changing this module

- startup regression
- shutdown regression
- route assembly regression

## Canonical change shape

- keep bootstrap explicit
- construct apps through `DefaultConfig()` / `AppConfig`, and use `AppDependencies` as the explicit constructor dependency contract
- treat `DefaultConfig()` as the supported baseline: `Prepare` rejects empty server addresses and negative server hardening values before freezing route/middleware mutation; zero timeout fields keep standard `http.Server` semantics
- keep lifecycle behavior reviewable
- keep one canonical lifecycle path: `Prepare` + `Server` + `Shutdown`
- split handler preparation from server preparation: `ServeHTTP` only freezes config/router state and builds the handler, while `Prepare` is the explicit path that validates server-only config, allocates `http.Server`, and prepares connection tracking
- keep server-only preparation validation non-destructive: TLS config and certificate load failures return before route/middleware mutation is frozen
- keep TLS on the same public serve path: `Prepare` loads configured certificates into the returned `*http.Server`, and callers use `ListenAndServeTLS("", "")` on that prepared server when TLS is enabled
- keep `core` as the first-party router owner: route wiring goes through `App.AddRoute` / `App.Get` / `App.Post`; named routes use route options such as `router.WithRouteName(...)` on `App.AddRoute` or `App.Any`; reverse URL lookup goes through `App.URL(...)`, not raw router replacement or mutation
- keep HTTP request metrics explicit in middleware/app-local wiring; `core` does not own live observer attachment state
- keep readiness ownership out of `core`; callers own the outer serve loop, so readiness signaling must stay app-local instead of pretending the kernel knows when traffic can flow
- keep logger subsystem ownership out of `core`; `core.AppDependencies{Logger: ...}` injects a passive dependency and callers own logger initialization and flushing
- keep app-local debug flags and env-file metadata outside `core`; the kernel owns HTTP runtime state, not devtools metadata transport
- keep debug/runtime snapshot payloads in `x/devtools`; `core` owns lifecycle state, not tooling response contracts
- keep the app logger kernel-owned on `App.Logger()`; `router` stays logger-free and does not mirror app logger state
- keep router behavior policy in typed config, not in ad hoc constructor options
- push feature-specific wiring back to app-local code or the owning extension
- preserve `net/http` compatibility while keeping `core` as a kernel

## Frozen behavior matrix

These behaviors are part of the current stable-root freeze baseline:

| Surface | Behavior |
| --- | --- |
| Construction | `New` copies `AppConfig` by value and resolves missing dependencies to safe defaults |
| Preparation state | `PreparationState` names the stable mutation, handler-prepared, and server-prepared phases exposed by core |
| Config validation | `Prepare` rejects invalid server config before freezing route/middleware mutation |
| Route wiring | `AddRoute` and method helpers delegate to the owned router with explicit method/path handlers |
| Middleware wiring | `Use` preserves registration order and rejects nil middleware without partial registration |
| `ServeHTTP` | lazily prepares the handler only, skips server-only config validation, freezes later route/middleware mutation, and remains `net/http` compatible |
| `Prepare` | freezes handler state, builds one `http.Server`, prepares active HTTP connection tracking, and is idempotent |
| `Prepare` failure | server-only config errors return before freezing route/middleware mutation |
| `Server` | returns an error before explicit server preparation and returns the prepared server after `Prepare` |
| `Shutdown` | requires a prepared server, tolerates nil contexts by using `context.Background()`, delegates shutdown to `http.Server`, and starts active-connection drain logging at most once |
| TLS | `Prepare` loads configured cert/key material into the returned `*http.Server` |
| Ownership | logger lifecycle, readiness signaling, debug routes, and feature wiring remain caller-owned |

Focused regression coverage lives in `core/freeze_test.go`,
`core/lifecycle_test.go`, `core/app_test.go`, and `core/routing_test.go`.
