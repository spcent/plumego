# Card 0781

Priority: P0
State: active
Primary Module: core
Owned Files:
- `core/config.go`
- `core/introspection.go`
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/app_test.go`
- `core/introspection_test.go`
- `core/lifecycle_test.go`
- `x/devtools/devtools.go`
Depends On:

Goal:
- Collapse `core` runtime introspection onto one clear preparation-state
  contract so callers stop reasoning about overlapping booleans that describe
  internal transitions rather than meaningful kernel phases.

Problem:
- `core.RuntimeSnapshot` currently exposes `Started`, `ConfigFrozen`, and
  `ServerPrepared`.
- These flags are not one coherent lifecycle model:
  `ServeHTTP()` can freeze config without preparing a server, while
  `Prepare()` starts logger hooks and prepares a server without actually
  serving traffic.
- As a result, first-party tooling receives three partially overlapping state
  booleans with no single canonical interpretation.

Scope:
- Replace the current overlapping runtime booleans with one canonical exported
  preparation-state contract.
- Keep the new contract aligned with the actual kernel states that matter to
  callers.
- Update first-party tooling and tests to consume the normalized state model.

Non-goals:
- Do not reintroduce readiness or serving-state ownership into `core`.
- Do not keep legacy booleans as aliases.
- Do not broaden runtime introspection beyond the normalized kernel state.

Files:
- `core/config.go`
- `core/introspection.go`
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/app_test.go`
- `core/introspection_test.go`
- `core/lifecycle_test.go`
- `x/devtools/devtools.go`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./x/devtools/...`
- `go vet ./core/... ./x/devtools/...`

Docs Sync:
- Update docs only if the exported runtime snapshot fields or meanings change.

Done Definition:
- `core` exposes one canonical preparation-state contract instead of three
  overlapping booleans.
- First-party tooling and tests assert the new normalized state directly.
- No stale `Started` / `ConfigFrozen` / `ServerPrepared` contract remains.
