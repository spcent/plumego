# x/ops

## Purpose

`x/ops` provides protected operations endpoints for queues, receipts, health HTTP orchestration, and runtime diagnostics.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is an auth-gated admin endpoint
- the task is queue or receipt diagnostics
- the task is health HTTP manager orchestration, history, metrics, or protected health diagnostics
- the task is a protected runtime operations surface

## Do not use this module for

- generic observability adapter work
- application bootstrap
- unauthenticated public diagnostics

## First files to read

- `x/ops/module.yaml`
- `specs/task-routing.yaml`
- `x/ops/ops.go`
- `x/ops/healthhttp`

## Main risks when changing this module

- weakened auth boundaries
- hidden admin side effects
- mixing broader observability concerns into protected ops endpoints

## Canonical change shape

- keep admin routing explicit and auth-gated
- keep health manager execution policy in `x/ops/healthhttp`, not stable `health`
- use `x/ops` only for protected admin surfaces
- keep broader observability adapter work in `x/observability`
- `x/ops/healthhttp` JSON success bodies use `contract.WriteResponse`; raw probe/export exceptions are limited to the `LiveHandler` liveness text response and health history CSV export
- protected ops JSON errors use `contract.WriteError` with explicit uppercase codes; not-configured hooks use `*_NOT_CONFIGURED`, hook failures use `*_FAILED`, and health-history query failures use `INVALID_QUERY` with safe messages

## Boundary with x/observability

- `x/ops`: protected admin endpoints, health HTTP manager orchestration, auth-gated diagnostics, and runtime control surfaces
- `x/observability`: broader adapter, export, and diagnostics integration work
- stable `middleware/*`: transport-only observability primitives
