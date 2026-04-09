# metrics

## Purpose

`metrics` holds stable metrics contracts and the small in-memory collectors that
other modules can depend on safely.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- changing collector contracts
- adding base collectors or aggregate collector composition
- wiring stable instrumentation against `Recorder` or `HTTPObserver`

## Do not use this module for

- Prometheus or tracing implementations
- dev-only dashboard collectors
- feature-specific metrics reporters or exporters
- rolling-window aggregation helpers
- repo-wide metrics test helpers
- app bootstrap

## First files to read

- `metrics/module.yaml`
- `metrics/collector.go`
- `metrics/multi.go`
- owning extension docs when the change is implementation-specific

## Canonical change shape

- keep collector APIs small
- keep base collectors generic and transport-agnostic
- keep only aggregate collector composition in stable `metrics`
- keep `AggregateCollector` limited to `Record`, shared `ObserveHTTP`, stats, and reset semantics
- keep metric identity canonical as `MetricRecord.Name`; use `Labels` for dimensions instead of parallel type catalogs
- keep feature-specific observer interfaces in their owning package; only the shared HTTP observer stays in stable `metrics`
- keep non-HTTP feature helper record builders in owning extensions or `x/observability` helper packages
- keep Prometheus and tracing adapters in `x/observability`
- keep record-buffer inspection and retention tuning in `x/observability/recordbuffer`
- keep rolling-window aggregation in `x/observability/windowmetrics`
- keep metrics test helpers in `x/observability/testmetrics`
- keep dev-only collectors in `x/devtools`
- keep feature-specific metrics ownership in the owning extension
