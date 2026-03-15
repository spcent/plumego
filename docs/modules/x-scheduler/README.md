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
