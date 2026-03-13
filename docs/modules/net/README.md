# Net Module

> **Status**: Legacy public root removed

## Current Mapping

The historical `net/*` root is no longer part of the public repository layout.

Use these entrypoints instead:

| Legacy area | Current entrypoint |
|-------------|--------------------|
| service discovery | `x/discovery` |
| IPC | `x/ipc` |
| messaging orchestration | `x/messaging` |
| durable queueing primitive | `x/mq` |
| inbound/outbound webhooks | `x/webhook` |
| websocket utilities | `x/websocket` |

`internal/nethttp` exists only for framework-owned helpers inside this repository. It is not a supported public import path for applications outside this module.

## Scope Note

The subdocuments under this directory are migration notes or archived references. New application code should import `x/*` capability packs directly.

For messaging-related work, start in `x/messaging` first. Open `x/mq` or `x/pubsub` only when the task is explicitly about queue or broker primitives.
