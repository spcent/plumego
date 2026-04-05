# Card 0765

Priority: P2
State: done
Primary Module: core
Owned Files:
- `core/app.go`
- `core/app_helpers.go`
- `core/options_test.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Remove the duplicate app-to-router logger shadow state so `core` has one
  kernel logger owner.

Problem:
- `App` already owns the application logger through `a.logger`.
- `core` still mirrors that logger into the router through
  `syncRouterConfig(...)`, and `New(...)` / `ensureRouter()` reapply that mirror
  state each time a router is installed or ensured.
- After `contract.AdaptCtxHandler(...)` was simplified, first-party code no
  longer depends on router-carried logger state for route registration.
- This leaves `core` maintaining two copies of the same logger contract with no
  clear consumer in first-party code.

Scope:
- Remove app-to-router logger mirroring from `core`.
- Keep `App.Logger()` as the single logger owner for app wiring.
- Update tests/docs so they no longer imply that configuring the app logger also
  mutates router state.

Non-goals:
- Do not redesign the standalone `router` package logger API in this card.
- Do not remove `App.Logger()`.
- Do not change middleware logger usage outside the kernel boundary.

Files:
- `core/app.go`
- `core/app_helpers.go`
- `core/options_test.go`
- `docs/modules/core/README.md`

Tests:
- Add or update coverage so core logger configuration is asserted through
  `App.Logger()` only.
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Update the core primer to treat the app logger as kernel-owned state rather
  than implicit router configuration.

Done Definition:
- `core` no longer mirrors app logger state into router state.
- The app logger has one kernel owner and one documented access path.
- Core tests/docs no longer rely on router logger shadow behaviour.

Outcome:
- Removed app-to-router logger mirroring from `core`, so `syncRouterConfig(...)`
  now only applies router-owned method-not-allowed state.
- Kept `App.Logger()` as the single kernel logger owner and added tests proving
  the default router stays logger-free while custom router loggers are not
  overwritten by `core`.
- Updated the core primer to describe logger ownership through `App.Logger()`
  instead of implying implicit router shadow state.
