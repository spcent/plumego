# x/pubsub

## Purpose

`x/pubsub` provides subordinate in-process publish-subscribe primitives used by messaging, webhook, and local debug flows.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is broker semantics or fan-out behavior
- the task is already narrower than the `x/messaging` family entrypoint

## Do not use this module for

- application bootstrap
- general messaging family discovery
- business event workflow logic

## First files to read

- `x/pubsub/module.yaml`
- `x/pubsub/*.go`
- `docs/modules/x-messaging/README.md`

## Canonical change shape

- keep broker injection explicit
- avoid hidden global brokers
- preserve deterministic fan-out behavior
