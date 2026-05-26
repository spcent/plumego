# docs/modules — Directory Index

This directory contains module primers for stable roots and extension packages.

## Naming Convention

Directories use the following naming scheme:

| Directory pattern | Meaning | Example |
|---|---|---|
| `<name>/` (no prefix) | Stable root package | `core/`, `contract/`, `middleware/` |
| `x-<family>/` | Top-level extension family | `x-ai/`, `x-gateway/`, `x-messaging/` |
| `x-<subpkg>/` | Sub-package of an extension family | `x-cache/` = `x/data/cache`, `x-mq/` = `x/messaging/mq` |

The `x-{subpkg}` directories share the same `x-` prefix as top-level extensions.
Check the H1 title (or the import path blockquote at the top of each README) to
determine whether the primer covers a top-level family or a sub-package.

## Sub-package Primers

These primers document sub-packages of their parent family. The actual Go import
path differs from the directory name:

| Directory | Import path | Parent family |
|---|---|---|
| `x-cache/` | `x/data/cache` | [`x-data/`](x-data/README.md) |
| `x-devtools/` | `x/observability/devtools` | [`x-observability/`](x-observability/README.md) |
| `x-discovery/` | `x/gateway/discovery` | [`x-gateway/`](x-gateway/README.md) |
| `x-ipc/` | `x/gateway/ipc` | [`x-gateway/`](x-gateway/README.md) |
| `x-mq/` | `x/messaging/mq` | [`x-messaging/`](x-messaging/README.md) |
| `x-ops/` | `x/observability/ops` | [`x-observability/`](x-observability/README.md) |
| `x-pubsub/` | `x/messaging/pubsub` | [`x-messaging/`](x-messaging/README.md) |
| `x-scheduler/` | `x/messaging/scheduler` | [`x-messaging/`](x-messaging/README.md) |
| `x-webhook/` | `x/messaging/webhook` | [`x-messaging/`](x-messaging/README.md) |

## Maturity Status

Extension maturity (experimental / beta / ga) is tracked in
`docs/EXTENSION_MATURITY.md`. Module manifests (`<package>/module.yaml`) are
the machine-readable source of truth for status, owner, risk, and validation
commands.
