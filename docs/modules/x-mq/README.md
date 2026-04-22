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

## Boundary rules

- `x/mq` is a subordinate primitive under `x/messaging`; do not bypass `x/messaging` for cross-family wiring
- keep queue persistence and worker coordination local to `x/mq`; do not push worker state into stable roots
- keep retry and failure visibility explicit in `x/mq`; do not add implicit retry policies at import time
- do not expose store-backend connection strings or topic naming conventions through the `x/mq` API surface
