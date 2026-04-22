# Card 2115: Observability Exporter Constructor Contract

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: done
Primary Module: x/observability
Owned Files:
- x/observability/exporter.go
- x/observability/prometheus_test.go
- x/observability/observability_test.go
- docs/modules/x-observability/README.md
Depends On: none

Goal:
- Make Prometheus exporter construction explicit when the collector dependency is nil.
- Remove panic-only behavior as the only way to discover invalid exporter wiring.

Scope:
- Audit `NewPrometheusExporter` and related tests for nil collector handling.
- Add an error-returning constructor or documented compatibility wrapper that keeps export behavior explicit.
- Preserve collector interfaces, exporter registration behavior, and app-facing observability ownership.
- Add focused tests for nil collector and normal construction.

Non-goals:
- Do not move transport-only middleware primitives into `x/observability`.
- Do not add hidden globals, auto-registration, or core bootstrap ownership.
- Do not change metric collection semantics or Prometheus output format.

Files:
- `x/observability/exporter.go`: normalize nil collector construction behavior.
- `x/observability/prometheus_test.go`: update exporter coverage if appropriate.
- `x/observability/observability_test.go`: add constructor compatibility coverage if appropriate.
- `docs/modules/x-observability/README.md`: document constructor guidance if public behavior changes.

Tests:
- `go test -race -timeout 60s ./x/observability/...`
- `go test -timeout 20s ./x/observability/...`
- `go vet ./x/observability/...`

Docs Sync:
- Required if public constructor behavior or docs for exporter wiring change.

Done Definition:
- Invalid nil collector wiring has a non-panic constructor path.
- Existing constructor compatibility is preserved or migrated with symbol-change protocol evidence.
- Focused tests cover nil and valid collector construction.
- The three listed validation commands pass.

Outcome:
- Added `NewPrometheusExporterE` and `ErrNilCollector` for explicit non-panic nil collector handling.
- Preserved `NewPrometheusExporter` compatibility by delegating to the error-returning constructor and panicking on invalid input.
- Added tests for nil collector handling and successful exporter construction through the error-returning path.
- Documented exporter constructor guidance in `docs/modules/x-observability/README.md`.
- Validation passed:
  - `go test -race -timeout 60s ./x/observability/...`
  - `go test -timeout 20s ./x/observability/...`
  - `go vet ./x/observability/...`
