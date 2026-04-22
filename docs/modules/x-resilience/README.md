# x/resilience

## Purpose

`x/resilience` contains reusable resilience components that are not part of the stable core and do not belong to a single feature family.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is a reusable circuit breaker or similar resilience primitive
- the behavior should be shared across multiple extension families

## Do not use this module for

- app bootstrap
- stable security policy
- feature-specific orchestration that belongs in `x/ai`, `x/gateway`, or another owning extension

## First files to read

- `x/resilience/module.yaml`
- `x/resilience/circuitbreaker`
- `specs/extension-taxonomy.yaml`

## Canonical change shape

- keep reusable resilience primitives here instead of in stable roots
- keep HTTP or transport adapters explicit and colocated with the primitive when they are generic
- keep feature-specific orchestration in the owning extension package

## Boundary rules

- `x/resilience` owns reusable circuit breaker and rate limit primitives; do not add these to stable `security` or stable `middleware`
- keep resilience primitive state instance-scoped; do not introduce package-level global state or implicit registration
- keep HTTP or transport adapters local to the owning extension when they are generic; do not push them into stable roots
- feature-specific orchestration (retry strategies tied to business rules) belongs in the owning extension, not in `x/resilience`
