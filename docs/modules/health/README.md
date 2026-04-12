# health

## Purpose

`health` owns readiness state, component checker contracts, and component health models.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- representing liveness or readiness state
- sharing component health result models across stable and extension packages

## Do not use this module for

- health manager ownership
- component registration and check orchestration
- retry, timeout, or concurrency policy for checks
- HTTP endpoint ownership
- protected ops endpoints
- gateway diagnostics
- build metadata
- health history export or retention
- metrics, trend analysis, or ops reporting

## First files to read

- `health/module.yaml`
- `health/*.go`
- `x/ops/healthhttp` when the task is HTTP exposure or check orchestration

## Canonical change shape

- keep health state transport-agnostic
- keep execution policy in `x/ops/healthhttp`
- expose HTTP handlers from reference or extensions, not from health itself
- keep analytics and reporting in owning extensions, not in stable `health`

## Boundary with HTTP exposure

- `health` owns state and models, not HTTP routes
- `x/ops/healthhttp` owns health managers, check execution, retries, timeouts, history, metrics, and HTTP handlers
- expose health endpoints from `reference/standard-service` or extension packages such as `x/ops/healthhttp`
- do not let `health` grow transport helpers or endpoint registration APIs
- keep build info, history export, and report-generation surfaces in `x/ops/healthhttp` or other owning extensions
