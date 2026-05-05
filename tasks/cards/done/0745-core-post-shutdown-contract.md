# 0745 - core Post-Shutdown Contract

State: done
Priority: P1
Primary Module: core

## Goal

Make core's behavior after successful `Shutdown` explicit and covered by tests.

## Scope

- Lock the current post-shutdown contract without adding new lifecycle states.
- Document that `Shutdown` closes the prepared `http.Server`, while the app
  remains `server_prepared` and handler serving through `ServeHTTP` remains
  available.
- Cover repeated `Shutdown`, repeated `Prepare`, `Server`, and `ServeHTTP`
  behavior after a real served shutdown.

## Non-goals

- Do not introduce a new exported `stopped` or `shutdown` state.
- Do not make core restart or recreate `http.Server` after shutdown.
- Do not change `http.Server.Shutdown` delegation semantics.

## Files

- `core/lifecycle_test.go`
- `docs/modules/core/README.md`
- `README.md`
- `README_CN.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go vet ./core/...`
- `go run ./internal/checks/module-manifests`

## Docs Sync

Required because this clarifies lifecycle semantics in public docs.

## Done Definition

- Post-shutdown behavior is explicitly documented.
- A regression test proves the stable behavior after a real served shutdown.
- Core tests, vet, and module manifest checks pass.

## Outcome

- Documented that successful shutdown keeps the app in `server_prepared`, keeps
  the same closed `*http.Server`, accepts repeated shutdown, and leaves
  `ServeHTTP` available as a prepared handler.
- Added a real served shutdown regression test for post-shutdown `Prepare`,
  `Server`, repeated `Shutdown`, `ServeHTTP`, and no-restart behavior.
- Verified with `go test -timeout 20s ./core/...`, `go vet ./core/...`, and
  `go run ./internal/checks/module-manifests`.
