# docs/modules — Directory Index

This directory contains module primers for stable roots and extension packages.

## Naming Convention

Directories use the following naming scheme:

| Directory pattern | Meaning | Example |
|---|---|---|
| `<name>/` (no prefix) | Stable root package | `core/`, `contract/`, `middleware/` |
| `x/<family>/` | Top-level extension family | `x/ai/`, `x/gateway/`, `x/messaging/` |
| `x/<family>/<subpkg>/` | Sub-package of an extension family | `x/data/cache/`, `x/messaging/mq/` |

The directory tree under `docs/modules/x/` is a 1:1 mirror of the `x/` source tree.
The primer path is always `docs/modules/<import-path>/README.md`.

## Extension Primers

| Import path | Primer |
|---|---|
| `x/ai` | [`x/ai/README.md`](x/ai/README.md) |
| `x/data` | [`x/data/README.md`](x/data/README.md) |
| `x/data/cache` | [`x/data/cache/README.md`](x/data/cache/README.md) |
| `x/fileapi` | [`x/fileapi/README.md`](x/fileapi/README.md) |
| `x/frontend` | [`x/frontend/README.md`](x/frontend/README.md) |
| `x/gateway` | [`x/gateway/README.md`](x/gateway/README.md) |
| `x/gateway/discovery` | [`x/gateway/discovery/README.md`](x/gateway/discovery/README.md) |
| `x/gateway/ipc` | [`x/gateway/ipc/README.md`](x/gateway/ipc/README.md) |
| `x/messaging` | [`x/messaging/README.md`](x/messaging/README.md) |
| `x/messaging/mq` | [`x/messaging/mq/README.md`](x/messaging/mq/README.md) |
| `x/messaging/pubsub` | [`x/messaging/pubsub/README.md`](x/messaging/pubsub/README.md) |
| `x/messaging/scheduler` | [`x/messaging/scheduler/README.md`](x/messaging/scheduler/README.md) |
| `x/messaging/webhook` | [`x/messaging/webhook/README.md`](x/messaging/webhook/README.md) |
| `x/observability` | [`x/observability/README.md`](x/observability/README.md) |
| `x/observability/devtools` | [`x/observability/devtools/README.md`](x/observability/devtools/README.md) |
| `x/observability/ops` | [`x/observability/ops/README.md`](x/observability/ops/README.md) |
| `x/openapi` | [`x/openapi/README.md`](x/openapi/README.md) |
| `x/resilience` | [`x/resilience/README.md`](x/resilience/README.md) |
| `x/rest` | [`x/rest/README.md`](x/rest/README.md) |
| `x/rpc` | [`x/rpc/README.md`](x/rpc/README.md) |
| `x/tenant` | [`x/tenant/README.md`](x/tenant/README.md) |
| `x/validate` | [`x/validate/README.md`](x/validate/README.md) |
| `x/websocket` | [`x/websocket/README.md`](x/websocket/README.md) |

## Maturity Status

Extension maturity (experimental / beta / ga) is tracked in
`docs/EXTENSION_MATURITY.md`. Module manifests (`<package>/module.yaml`) are
the machine-readable source of truth for status, owner, risk, and validation
commands.
