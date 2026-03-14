# x/mq

## Purpose

`x/mq` provides subordinate durable queue primitives and worker coordination under the broader messaging family.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is queue persistence or worker coordination
- the task is already known to be narrower than the `x/messaging` family entrypoint

## Do not use this module for

- application bootstrap
- general messaging family discovery
- business workflow orchestration

## First files to read

- `x/mq/module.yaml`
- the owning package under `x/mq/*`
- `docs/modules/x-messaging/README.md`

## Canonical change shape

- keep queue semantics explicit
- keep worker wiring local and reviewable
- preserve persistence correctness and retry visibility
