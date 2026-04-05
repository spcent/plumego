# Card 0774

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/app.go`
- `core/introspection.go`
- `x/observability/observability.go`
- `x/observability/observability_test.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Remove dead HTTP observer attachment state from `core` so metrics wiring has
  one explicit owner and one real execution path.

Problem:
- `core.App` still stores `httpMetrics`, and `AttachHTTPObserver(...)` still
  mutates that field.
- The kernel never reads `httpMetrics` when building the handler or prepared
  server, so the field is write-only shadow state.
- `x/observability.Configure(...)` still calls `hooks.AttachHTTPObserver(...)`
  as if it were enough to enable request metrics, but actual observation only
  happens when callers wire `middleware/httpmetrics.Middleware(collector)`.
- This leaves a misleading `core` API that looks live but has no transport
  effect.

Scope:
- Remove `httpMetrics` storage and `AttachHTTPObserver(...)` from `core`.
- Make `x/observability` use an explicit middleware-owned collector path
  instead of mutating dead kernel state.
- Update tests and module docs to match the explicit metrics wiring contract.

Non-goals:
- Do not redesign the `metrics` package.
- Do not move HTTP metrics collection into `core`.
- Do not add global observer registries.

Files:
- `core/app.go`
- `core/introspection.go`
- `x/observability/observability.go`
- `x/observability/observability_test.go`
- `docs/modules/core/README.md`

Tests:
- `go test -race -timeout 60s ./core/... ./x/observability/...`
- `go test -timeout 20s ./...`
- `go vet ./...`

Docs Sync:
- Remove any claim that `core` can attach live HTTP metrics observers by
  itself.

Done Definition:
- `core` no longer stores write-only HTTP observer state.
- `x/observability` uses one explicit request-metrics wiring path.
- First-party docs no longer describe dead observer attachment behavior.

Outcome:
- Removed write-only `httpMetrics` state and `AttachHTTPObserver(...)` from
  `core`.
- Simplified `x/observability.Hooks` so metrics wiring no longer pretends the
  kernel can attach a live request observer.
- Updated tests and core module docs to reflect explicit middleware-owned HTTP
  metrics wiring.
