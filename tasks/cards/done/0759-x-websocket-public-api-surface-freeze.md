# Card 0759

Milestone: M-003
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/errors.go`
- `x/websocket/validation.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
Depends On: 0758

Goal:
- Freeze a smaller, stable-worthy public API surface for `x/websocket`.

Problem:
The package still exposes helper functions and error constructors that are not core websocket transport APIs, plus a panic convenience constructor. Leaving them public makes future stable compatibility harder.

Scope:
- Remove the panic-prone `NewHub` constructor in favor of `NewHubWithConfigE`.
- Internalize or delete non-core helper exports that are only used by package tests or internals.
- Update module manifest public entrypoints.
- Update tests, docs, and examples to use the remaining stable candidates.
- Follow `AGENTS.md §7.1` for every exported symbol removal.

Non-goals:
- Do not remove core entrypoints needed for explicit route registration and serving.
- Do not change protocol behavior.
- Do not promote module maturity.

Files:
- `x/websocket/hub.go`
- `x/websocket/errors.go`
- `x/websocket/validation.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for removed public symbols and public entrypoint inventory.

Done Definition:
- Public surface is limited to intentional websocket transport APIs.
- No production or test caller references removed exported symbols.
- Module manifest public entrypoints match the package.

Outcome:
- Removed the panic-prone `NewHub` public constructor.
- Migrated runtime, scaffold templates, and tests to `NewHubWithConfigE`.
- Removed unused exported websocket error helper types/functions that were not part of the transport core.
- Updated `x/websocket/module.yaml` and module docs public entrypoints.
- Re-ran old-symbol searches; removed names have no Go call sites.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go build ./...`
  - `go run ./internal/checks/module-manifests`
