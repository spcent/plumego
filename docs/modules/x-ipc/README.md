# x/ipc

## Purpose

`x/ipc` provides lower-level inter-process communication helpers and explicit
IPC adapters under the broader gateway family.

## v1 Status

- `experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is explicit IPC transport behavior
- the task is client/server communication between processes

## Do not use this module for

- application bootstrap
- general gateway family discovery
- business workflow orchestration

## First files to read

- `x/ipc/module.yaml`
- `x/ipc/ipc.go`
- `docs/modules/x-gateway/README.md`

## Public entrypoints

- `NewServer`
- `NewHeartbeatClient`
- `NewPool`
- `NewFramedClient`
- `NewStreamClient`
- `NewRateLimitedServer`

## Main risks when changing this module

- framing or reconnect regression
- hidden process-wide state or socket lifecycle leaks
- failure-path visibility drift

## Canonical change shape

- keep transport contracts explicit
- keep process-wide side effects reviewable
- prefer family-level discovery in `x/gateway` before widening `x/ipc`

## Boundary rules

- `x/ipc` provides explicit IPC transport; do not use it as a substitute for
  in-process messaging (`x/pubsub`) or durable queues (`x/mq`)
- keep transport contracts and process-wide side effects explicit and reviewable; do not add implicit channel registration at import time
- do not expose IPC connection strings or socket paths through stable roots; keep them local to `x/ipc` adapters
- prefer `x/gateway` as the app-facing entrypoint for edge or cross-process
  transport before widening `x/ipc` callers

## Validation commands

- `go test -race -timeout 60s ./x/ipc/...`
- `go test -timeout 20s ./x/ipc/...`
- `go vet ./x/ipc/...`
