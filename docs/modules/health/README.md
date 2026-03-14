# health

## Purpose

`health` owns readiness state and health models.

## Use this module when

- representing liveness or readiness state
- aggregating health status in-process

## Do not use this module for

- HTTP endpoint ownership
- protected ops endpoints
- gateway diagnostics

## First files to read

- `health/module.yaml`
- `health/*.go`
- `x/ops/healthhttp` when the task is HTTP exposure

## Canonical change shape

- keep health state transport-agnostic
- expose HTTP handlers from reference or extensions, not from health itself

## Boundary with HTTP exposure

- `health` owns state and models, not HTTP routes
- expose health endpoints from `reference/standard-service` or extension packages such as `x/ops/healthhttp`
- do not let `health` grow transport helpers or endpoint registration APIs
