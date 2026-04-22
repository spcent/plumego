# x/ipc

## Purpose

`x/ipc` provides lower-level inter-process communication helpers and explicit IPC adapters.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is explicit IPC transport behavior
- the task is client/server communication between processes

## Do not use this module for

- application bootstrap
- general messaging family discovery
- business workflow orchestration

## First files to read

- `x/ipc/module.yaml`
- `x/ipc/ipc.go`
- `docs/modules/x-messaging/README.md`

## Canonical change shape

- keep transport contracts explicit
- keep process-wide side effects reviewable
- prefer family-level discovery in `x/messaging` before widening `x/ipc`

## Boundary rules

- `x/ipc` provides explicit IPC transport; do not use it as a substitute for in-process messaging (`x/pubsub`) or durable queues (`x/mq`)
- keep transport contracts and process-wide side effects explicit and reviewable; do not add implicit channel registration at import time
- do not expose IPC connection strings or socket paths through stable roots; keep them local to `x/ipc` adapters
- prefer `x/messaging` as the app-facing entrypoint for cross-process messaging before widening `x/ipc` callers
