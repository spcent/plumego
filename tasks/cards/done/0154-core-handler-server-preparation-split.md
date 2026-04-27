# Card 0154

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/introspection.go`
- `core/app_test.go`
- `core/introspection_test.go`
Depends On:
- `0153-core-health-readiness-ownership-removal.md`

Goal:
- Split handler preparation from server preparation so `core` has one clear
  contract for embedded-handler use and one clear contract for owned-server use.

Problem:
- `ServeHTTP` currently calls `ensurePrepared()`, which freezes config, builds
  the handler, and also allocates `http.Server` plus connection tracking.
- This means using `*core.App` as a plain `http.Handler` implicitly creates
  server state even when callers never ask `core` for a server.
- `RuntimeSnapshot.ServerPrepared` therefore flips true after first request
  handling, even if `Prepare()` / `Server()` were never used.
- The kernel currently conflates "handler is usable" with "owned server is
  prepared".

Scope:
- Split the internal preparation flow into handler-only preparation and
  explicit server preparation.
- Keep `ServeHTTP` net/http-compatible, but make it prepare only what direct
  handler use actually needs.
- Keep `Prepare()` / `Server()` as the explicit path that creates `http.Server`
  and connection tracking.
- Update snapshot/tests to reflect the narrower `ServerPrepared` meaning.

Non-goals:
- Do not remove `ServeHTTP`.
- Do not redesign TLS semantics.
- Do not reintroduce a private serve wrapper.

Files:
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/introspection.go`
- `core/app_test.go`
- `core/introspection_test.go`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Keep docs aligned on the distinction between handler readiness and prepared
  server state.

Done Definition:
- Direct `ServeHTTP` usage no longer allocates server-only state.
- `Prepare()` / `Server()` remain the only path that prepares `http.Server`.
- `ServerPrepared` reflects prepared-server state only, not first request use.

Outcome:
- Split preparation into `ensureHandlerPrepared()` and
  `ensureServerPrepared()`.
- Narrowed `ServeHTTP` so direct handler use freezes config/router state and
  builds only the composed handler.
- Kept `Prepare()` as the only path that allocates `http.Server` and
  connection tracking.
- Updated tests and docs so `RuntimeSnapshot.ServerPrepared` reflects prepared
  server state only.
