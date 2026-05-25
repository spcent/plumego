# Card 1512

Milestone: M-022
Recipe: specs/change-recipes/docs-and-config.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/ai
Owned Files:
- `x/ai/metrics/adapter.go`
- `x/ai/metrics/metrics_test.go`
- `docs/modules/metrics/README.md`
- `docs/modules/x-ai/README.md`
Depends On: 1511

Goal:
- Reconnect `x/ai/metrics` to the stable `metrics` package through an explicit
  adapter so AI instrumentation can emit into the stable collector contracts
  without maintaining a fully isolated metrics stack.

Scope:
- Add a stable-collector adapter in `x/ai/metrics`.
- Test the adapter behavior against the stable collector contract.
- Document the adapter boundary and the tradeoff that AI metric kinds flatten
  into stable `MetricRecord` events.

Non-goals:
- Do not delete `x/ai/metrics.Collector` or `MemoryCollector`.
- Do not widen the stable `metrics` package with AI-specific helper methods.
- Do not refactor `x/ai/instrumentation` call sites in this card.

Files:
- `x/ai/metrics/adapter.go`
- `x/ai/metrics/metrics_test.go`
- `docs/modules/metrics/README.md`
- `docs/modules/x-ai/README.md`

Acceptance Tests:
- `go test -timeout 20s ./x/ai/metrics ./metrics`

Tests:
- `go test -timeout 20s ./x/ai/... ./metrics/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- `docs/modules/metrics/README.md`
- `docs/modules/x-ai/README.md`

Validation:
- `go test -timeout 20s ./x/ai/... ./metrics/...`
- `go run ./internal/checks/dependency-rules`
- `gofmt -l .`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Added `x/ai/metrics.NewAggregateCollectorAdapter(...)`, which lets AI
  instrumentation write into the stable `metrics.AggregateCollector` contract
  without deleting the richer AI-local collector facade.
- The adapter records AI counter, gauge, histogram, and timing calls as stable
  `MetricRecord` events and keeps tag translation explicit; it stays nil-safe so
  optional wiring remains easy to compose.
- Updated the stable `metrics` and `x/ai` module docs to make the adapter the
  official bridge and to document the flattening tradeoff: stable collectors do
  not preserve AI metric kind as a separate contract field.
- Validation:
  - `go test -timeout 20s ./x/ai/... ./metrics/...`
  - `go run ./internal/checks/dependency-rules`
  - `gofmt -l .`
