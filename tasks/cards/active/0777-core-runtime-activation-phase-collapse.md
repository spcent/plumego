# Card 0777

Priority: P0
State: active
Primary Module: core
Owned Files:
- `core/app.go`
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/lifecycle_test.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Collapse `core` pre-serve lifecycle into one explicit activation phase so the
  kernel stops exposing a separate `Start()` step that does not correspond to
  socket readiness or request serving.

Problem:
- `Prepare()` already freezes config, builds the handler, and prepares the
  `http.Server`.
- `Start()` now only starts logger lifecycle hooks and flips extra runtime
  state.
- Since `core` no longer owns `ListenAndServe()`, this second step is easy to
  misread as server activation or readiness even though it is only runtime
  bookkeeping.
- The kernel still carries fragmented lifecycle state through `configFrozen`,
  `started`, `loggerStarted`, and prepared-server checks.

Scope:
- Remove `Start()` as a separate public phase.
- Make `Prepare()` perform the full explicit pre-serve activation transition,
  including logger lifecycle start.
- Simplify runtime state so it reflects one clear activation path plus
  shutdown.
- Update tests and core docs to match the collapsed lifecycle.

Non-goals:
- Do not reintroduce a private serve wrapper.
- Do not bring readiness signaling back into `core`.
- Do not redesign `Shutdown(ctx)` ownership.

Files:
- `core/app.go`
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/lifecycle_test.go`
- `docs/modules/core/README.md`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./...`
- `go vet ./...`

Docs Sync:
- Replace `Prepare + Start + Server + Shutdown` with the collapsed canonical
  lifecycle in core module docs.

Done Definition:
- `core` exposes one explicit pre-serve activation phase instead of separate
  `Prepare()` and `Start()`.
- Runtime state no longer models a fake activation step.
- Tests and module docs reflect the collapsed lifecycle contract.

Outcome:
