# health

## Purpose

`health` owns readiness state, component health models, and the minimum in-process manager surface.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- representing liveness or readiness state
- aggregating health status in-process
- registering component checkers and running in-process checks

## Do not use this module for

- HTTP endpoint ownership
- protected ops endpoints
- gateway diagnostics
- build metadata
- health history export or retention
- metrics, trend analysis, or ops reporting

## First files to read

- `health/module.yaml`
- `health/*.go`
- `x/ops/healthhttp` when the task is HTTP exposure

## Canonical change shape

- keep health state transport-agnostic
- expose HTTP handlers from reference or extensions, not from health itself
- keep analytics and reporting in owning extensions, not in stable `health`

## Boundary with HTTP exposure

- `health` owns state and models, not HTTP routes
- expose health endpoints from `reference/standard-service` or extension packages such as `x/ops/healthhttp`
- do not let `health` grow transport helpers or endpoint registration APIs
- keep build info, history export, and report-generation surfaces in `x/ops/healthhttp` or other owning extensions
