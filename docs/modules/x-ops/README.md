# x/ops

## Purpose

`x/ops` provides protected operations endpoints for queues, receipts, and runtime diagnostics.

## Use this module when

- the task is an auth-gated admin endpoint
- the task is queue or receipt diagnostics
- the task is a protected runtime operations surface

## Do not use this module for

- generic observability adapter work
- application bootstrap
- unauthenticated public diagnostics

## First files to read

- `x/ops/module.yaml`
- `specs/agent-entrypoints.yaml`
- `x/ops/ops.go`

## Main risks when changing this module

- weakened auth boundaries
- hidden admin side effects
- mixing broader observability concerns into protected ops endpoints

## Canonical change shape

- keep admin routing explicit and auth-gated
- use `x/ops` only for protected admin surfaces
- keep broader observability adapter work in `x/observability`
