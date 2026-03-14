# x/observability

## Purpose

`x/observability` is the app-facing extension root for broader observability adapters and export wiring.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is exporter or adapter integration work
- the task is broader diagnostics or telemetry pipeline wiring

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
