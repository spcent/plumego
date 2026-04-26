# x/observability

## Purpose

`x/observability` is the app-facing extension root for broader observability adapters and export wiring.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen
- Beta candidate once the extension stability policy's two-release API freeze
  evidence is available. Current blocker: no repository release history proves
  two consecutive minor releases without exported `x/observability/*` API changes.

## Use this module when

- the task is exporter or adapter integration work
- the task is broader diagnostics or telemetry pipeline wiring
- the task is moving non-stable observability helpers or metrics test utilities out of stable roots

## Do not use this module for

- transport-only middleware primitives
- application bootstrap
- feature-specific business metrics policy

## First files to read

- `x/observability/module.yaml`
- the owning package under `x/observability/*`
- `docs/modules/middleware/README.md`

## Canonical change shape

- keep export wiring explicit
- keep adapter-local behavior reviewable
- keep transport observability primitives in stable `middleware/*`
- keep buffered metric-record inspection in `x/observability/recordbuffer`
- keep rolling-window aggregation helpers in `x/observability/windowmetrics`
- keep metrics test utilities in `x/observability/testmetrics`
- keep DB analytics and slow-query helpers in `x/observability/dbinsights`

## Boundary rules

- `x/observability` is for exporter, adapter, and pipeline wiring only; keep transport-layer observation (per-request counters, latency) in stable `middleware/httpmetrics`
- do not add hidden global collectors or automatic registration at import time; pass `Hooks` and config explicitly via `Configure`
- keep `PrometheusCollector` and `PrometheusExporter` isolated per service instance; do not share collectors across test cases or goroutines without coordination
- use `NewPrometheusExporterE` when invalid collector wiring should be handled as an error; `NewPrometheusExporter` preserves legacy panic behavior for nil collectors
- keep `OpenTelemetryTracer` scoped per service; `Clear()` and `WithMaxSpans()` exist for bounded test use, not for production hot-path eviction
- DB insight helpers (`dbinsights`) are analytics utilities only; do not use them as a query-layer abstraction or replace stable `store` APIs

## Current test coverage

- `PrometheusCollector`: observe/handler round-trip, multiple requests with different labels, stats (ActiveSeries, TotalRecords, NameBreakdown, StartTime), Clear, WithMaxMemory eviction, concurrency, Prometheus text format and label escaping (newline and quote injection), empty namespace default, zero-max-memory fallback, eviction of least-used series
- `PrometheusExporter.Handler()`: output format (all metric names, label values), Content-Type header (`text/plain` prefix), empty-collector 200 response with uptime line
- `OpenTelemetryTracer`: empty-name fallback, span lifecycle (Start/End), 4xx/5xx/success span classification, span attributes completeness, parent trace-ID propagation, `GetSpanStats`, `Clear`, `WithMaxSpans` bounding
- `Configure`: both metrics and tracing enabled, custom namespace, custom service name, concurrent calls, max-series, custom path, mutable-callback invocation
- Subpackages: `recordbuffer`, `windowmetrics`, `testmetrics`, `testlog`, `tracer`, `featuremetrics`, `dbinsights` — each has dedicated unit tests

## Beta readiness

`x/observability` satisfies the current coverage and boundary portions of
`docs/EXTENSION_STABILITY_POLICY.md`: collector/exporter behavior, tracer
lifecycle, configuration, and supporting record-buffer, window, test, feature,
and DB insight packages have focused tests.

The module remains `experimental` until the release-history criterion is
verifiable. Promotion to `beta` requires evidence that exported
`x/observability/*` symbols have not changed for two consecutive minor
releases, plus owner sign-off recorded with the promotion card. Transport
observability primitives remain in stable `middleware/*`; exporter and adapter
wiring stays here.
