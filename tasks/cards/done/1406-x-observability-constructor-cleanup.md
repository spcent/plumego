# Card 1406

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/observability
Owned Files:
- x/observability/exporter.go
- x/observability/observability.go
- x/observability/observability_test.go
- x/observability/tracer/tracer.go
- x/observability/tracer/tracer_test.go
Depends On:
- 1405

Goal:
- Make dynamic observability construction prefer error-returning paths while preserving existing compatibility constructors.

Scope:
- Inventory panic-style constructors in `x/observability` and tracer subpackages.
- Add or document `E` constructor paths where dynamic configuration can fail.
- Keep existing panic constructors as compatibility wrappers for known-good config.
- Add negative tests for invalid dynamic config.

Non-goals:
- Do not change exporter/tracer public type names.
- Do not introduce OpenTelemetry or Prometheus dependencies.
- Do not move transport observability into stable roots.

Files:
- x/observability/exporter.go
- x/observability/observability.go
- x/observability/observability_test.go
- x/observability/tracer/tracer.go
- x/observability/tracer/tracer_test.go

Tests:
- go test -timeout 20s ./x/observability/... 
- go vet ./x/observability/...
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update `docs/modules/x-observability/README.md` if constructor guidance changes.

Done Definition:
- Dynamic observability construction has an error-returning path.
- Compatibility wrappers are documented and registered if they retain compatibility markers.
- Observability tests and boundary checks pass.

Outcome:
- Completed on May 15, 2026.
- Updated `Configure` metrics wiring to use `NewPrometheusExporterE` instead of the panic-style exporter constructor.
- Added `tracer.NewProbabilitySamplerE` and `tracer.NewTracerE` for dynamic tracer configuration validation while keeping the existing constructors unchanged for current callers.
- Added negative tracer constructor tests and documented the error-returning constructor guidance in `docs/modules/x-observability/README.md`.
- Validation passed:
  - `go test -timeout 20s ./x/observability/...`
  - `go vet ./x/observability/...`
  - `go run ./internal/checks/dependency-rules`
