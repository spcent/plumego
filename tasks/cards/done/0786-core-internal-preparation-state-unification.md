# Card 0786

Priority: P0
State: done
Primary Module: core
Owned Files:
- `core/app.go`
- `core/app_helpers.go`
- `core/lifecycle.go`
- `core/http_handler.go`
- `core/introspection.go`
- `core/app_test.go`
- `core/lifecycle_test.go`
- `core/routing_test.go`
- `core/introspection_test.go`
Depends On:

Goal:
- Unify `core` internal mutability and preparation tracking onto the same
  preparation-state model already exposed to callers.

Problem:
- `core` now exposes `PreparationState` externally, but internally it still
  carries separate `started` and `configFrozen` booleans.
- `started` no longer represents a real runtime hook or serving phase; it is
  only used for mutability checks and gets reset on `Shutdown(ctx)`, while
  `configFrozen` remains true.
- This leaves the kernel with two internal state paths and one external state
  path that no longer mean the same thing.

Scope:
- Replace `started` / `configFrozen` with one internal preparation-state model.
- Make `ensureMutable(...)`, `Prepare()`, `ServeHTTP()`, and `Shutdown(ctx)`
  derive from that one state model.
- Remove fake state rollback on shutdown if the app remains intentionally
  terminal after preparation.

Non-goals:
- Do not reintroduce serving/readiness ownership.
- Do not widen runtime introspection beyond preparation state.
- Do not preserve dead transitional booleans as aliases.

Files:
- `core/app.go`
- `core/app_helpers.go`
- `core/lifecycle.go`
- `core/http_handler.go`
- `core/introspection.go`
- `core/app_test.go`
- `core/lifecycle_test.go`
- `core/routing_test.go`
- `core/introspection_test.go`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Update docs only if caller-visible preparation/shutdown semantics change.

Done Definition:
- `core` uses one internal preparation-state model instead of `started` plus
  `configFrozen`.
- Mutability checks and lifecycle transitions reflect that single model.
- Shutdown no longer pretends to roll back a separate runtime-start flag.
