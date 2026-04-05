# Card 0763

Priority: P2
State: done
Primary Module: core
Owned Files:
- `core/app.go`
- `core/options.go`
- `core/introspection.go`
- `core/options_test.go`
- `x/observability/observability.go`
- `x/observability/observability_test.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Collapse HTTP metrics attachment in `core` to one explicit contract instead of
  keeping option-time, runtime, and accessor-shaped variants for the same
  capability.

Problem:
- `WithHTTPMetrics` writes one observer during app construction.
- `AttachHTTPObserver` adds observers later through fan-out composition.
- `HTTPMetrics()` exposes a dynamic wrapper object that reads whatever observer
  is currently stored.
- No first-party code calls the wrapper/accessor path, while
  `x/observability.Configure` still models the integration as a `SetHTTPMetrics`
  hook.
- This leaves `core` with multiple overlapping observer surfaces and no clear
  canonical wiring path.

Scope:
- Choose one canonical HTTP metrics attachment contract in `core`.
- Remove redundant option/accessor surfaces that only mirror the chosen
  contract.
- Keep `x/observability` aligned with the same attachment model instead of
  talking to a second pseudo-setter shape.

Non-goals:
- Do not redesign `middleware/httpmetrics`.
- Do not redesign Prometheus exporter behaviour.
- Do not move metrics collection into `core`.

Files:
- `core/app.go`
- `core/options.go`
- `core/introspection.go`
- `core/options_test.go`
- `x/observability/observability.go`
- `x/observability/observability_test.go`
- `docs/modules/core/README.md`

Tests:
- Add coverage for the chosen single attachment path and its composition
  behaviour.
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./x/observability/...`
- `go vet ./core/... ./x/observability/...`

Docs Sync:
- Update the core primer so metrics wiring is described through one attachment
  API.

Done Definition:
- HTTP metrics observer attachment has one canonical `core` contract.
- Redundant wrapper/setter surfaces are removed.
- `x/observability` wiring matches the same single contract.

Outcome:
- Removed `WithHTTPMetrics` and `(*App).HTTPMetrics()` so `core` no longer
  carries constructor-time and accessor-shaped duplicates for the same
  observer state.
- Kept `(*App).AttachHTTPObserver(...)` as the single explicit attachment path
  and added core coverage for fan-out composition and nil handling.
- Updated `x/observability.Configure` to use the same attachment contract via
  `Hooks.AttachHTTPObserver`, keeping metrics wiring aligned across modules.
