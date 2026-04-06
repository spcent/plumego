# metrics

## Purpose

`metrics` holds stable metrics contracts and the small in-memory collectors that
other modules can depend on safely.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- changing collector contracts
- adding base metrics helpers
- wiring stable instrumentation against `Recorder` or the focused observer interfaces

## Do not use this module for

- Prometheus or tracing implementations
- dev-only dashboard collectors
- feature-specific metrics reporters or exporters
- app bootstrap

## First files to read

- `metrics/module.yaml`
- `metrics/collector.go`
- `metrics/helpers.go`
- owning extension docs when the change is implementation-specific

## Canonical change shape

- keep collector APIs small
- keep base collectors generic and transport-agnostic
- keep Prometheus and tracing adapters in `x/observability`
- keep dev-only collectors in `x/devtools`
- keep feature-specific metrics ownership in the owning extension
