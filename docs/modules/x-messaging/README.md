# x/messaging

## Purpose

`x/messaging` is the canonical app-facing entrypoint for the messaging family.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is message send orchestration
- the task is app-facing queue, pubsub, scheduler, or webhook wiring
- the task needs a family-level messaging entrypoint instead of a lower-level primitive

## Do not use this module for

- application bootstrap
- stable-root abstractions
- direct low-level primitive work when the task is clearly only `x/mq` or `x/pubsub`

## First files to read

- `x/messaging/module.yaml`
- `x/messaging/entrypoints.go`
- `specs/extension-taxonomy.yaml`

## Main risks when changing this module

- message flow regression
- provider global state leakage
- family entrypoint ambiguity returning across sibling packages

## Boundary rules

- `x/messaging` is the app-facing family entrypoint; internal messaging primitives (`x/mq`, `x/pubsub`, `x/scheduler`, `x/webhook`) must not bypass it for cross-family wiring
- keep send orchestration explicit in `x/messaging`; do not add implicit retry, deduplication, or routing policies at import time
- do not expose transport-layer internals (broker connection strings, topic naming conventions) through `x/messaging` API; keep those local to the owning subordinate package
- keep `x/messaging` transport-agnostic at the family boundary; owning handlers choose the subordinate primitive

## Canonical change shape

- start app-facing messaging work here
- keep orchestration explicit
- keep `x/mq`, `x/pubsub`, `x/scheduler`, and `x/webhook` subordinate to this family
- app-facing and admin HTTP success responses use `contract.WriteResponse`; direct JSON writes are reserved for non-JSON streaming surfaces outside this family
- HTTP error responses expose stable contract codes and user-safe messages; provider, broker, scheduler, and store errors must not be written with raw `err.Error()` text

## Subordinate packages

Open sibling packages only when the task is already known to be narrow:

- `x/mq`: durable queue primitives and worker coordination
- `x/pubsub`: in-process broker primitives
- `x/scheduler`: scheduling primitives
- `x/webhook`: inbound verification or outbound delivery mechanics
