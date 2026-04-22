# x/scheduler

## Purpose

`x/scheduler` provides subordinate in-process scheduling primitives for cron, delayed jobs, and retries within the broader messaging extension family.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is delayed execution or retry coordination
- the task is already known to be narrower than the `x/messaging` family entrypoint
- the task is scheduler behavior rather than broader messaging orchestration

## Do not use this module for

- application bootstrap
- general messaging family discovery
- stable root entrypoints
- business workflow orchestration hidden in the scheduler layer

## First files to read

- `x/scheduler/module.yaml`
- `x/scheduler/*.go`
- `docs/modules/x-messaging/README.md`

## Canonical change shape

- keep job wiring explicit
- avoid hidden process-wide state
- preserve retry determinism and failure visibility

## Boundary rules

- `x/scheduler` is a subordinate primitive under `x/messaging`; do not use it as a cross-family entrypoint
- keep scheduler state instance-scoped; do not introduce process-wide job registries or implicit registration at import time
- retry determinism and failure visibility must stay explicit in `x/scheduler`; do not add hidden retry policies
- do not push scheduling-specific business rules (e.g. cron expressions tied to domain logic) into stable roots
