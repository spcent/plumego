# x/messaging

## Purpose

`x/messaging` is the canonical app-facing entrypoint for the messaging family.

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
- `specs/extension-entrypoints.yaml`

## Main risks when changing this module

- message flow regression
- provider global state leakage
- family entrypoint ambiguity returning across sibling packages

## Canonical change shape

- start app-facing messaging work here
- keep orchestration explicit
- keep `x/mq`, `x/pubsub`, `x/scheduler`, and `x/webhook` subordinate to this family
