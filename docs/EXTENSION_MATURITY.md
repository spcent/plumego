# Extension Maturity Dashboard

This dashboard is the human-readable triage view for `x/*` capability families.
Module manifests remain the machine-readable source of truth for status, owner,
risk, responsibilities, and validation commands.

Status values follow `docs/EXTENSION_STABILITY_POLICY.md`.

Drift check:

```bash
go run ./internal/checks/extension-maturity
```

Review source data:

```bash
go run ./internal/checks/extension-maturity -report
```

The drift check verifies status, risk, and owner against each extension
`module.yaml`. For beta candidates, it also verifies the evidence link and
blocker text against `specs/extension-beta-evidence.yaml`.

## App-Facing Families

| Family | Status | Risk | Owner | Recommended entrypoint | Validation | Evidence / blocker |
| --- | --- | --- | --- | --- | --- | --- |
| `x/ai` | experimental | high | ai-gateway | `x/ai/provider`, `x/ai/session`, `x/ai/streaming`, `x/ai/tool` for stable-tier adoption | `go test -timeout 20s ./x/ai/...` | Root family is not beta-ready; evaluate stable-tier subpackages separately |
| `x/data` | experimental | medium | persistence | `docs/modules/x-data/README.md` before subordinate data packages | `go test -timeout 20s ./x/data/...` | Topology-heavy package set; needs feature-level evaluation |
| `x/devtools` | experimental | medium | observability | Explicit local/protected debug mounting only | `go test -timeout 20s ./x/devtools/...` | Debug tooling; not a production admin surface |
| `x/discovery` | experimental | medium | edge | Caller-selected discovery backend | `go test -timeout 20s ./x/discovery/...` | Kubernetes/etcd backends need release observation |
| `x/fileapi` | experimental | medium | persistence | HTTP file transport over `x/data/file` and stable `store/file` contracts | `go test -timeout 20s ./x/fileapi/...` | Needs persistence and transport scenario evidence as behavior expands |
| `x/frontend` | experimental | medium | frontend | Explicit static or embedded asset serving | `go test -timeout 20s ./x/frontend/...` | Keep frontend serving out of canonical bootstrap defaults |
| `x/gateway` | experimental | medium | edge | `x/gateway` for proxy, rewrite, balancing, and edge transport | `go test -timeout 20s ./x/gateway/...` | [beta evidence](extension-evidence/x-gateway.md): release history, API snapshot, and owner sign-off missing |
| `x/messaging` | experimental | medium | messaging | App-facing messaging entrypoint before queue/pubsub primitives | `go test -timeout 20s ./x/messaging/...` | Evaluate after subordinate contracts settle |
| `x/observability` | experimental | medium | observability | Exporter, tracer, collector, and adapter wiring | `go test -timeout 20s ./x/observability/...` | [beta evidence](extension-evidence/x-observability.md): release history, API snapshot, and owner sign-off missing |
| `x/ops` | experimental | medium | observability | Protected admin and runtime diagnostics routes | `go test -timeout 20s ./x/ops/...` | Requires explicit auth boundary; not a public diagnostics default |
| `x/resilience` | experimental | medium | runtime | Reusable extension-layer circuit breaker and rate-limit primitives | `go test -timeout 20s ./x/resilience/...` | Cross-family primitive; needs adoption evidence before promotion |
| `x/rest` | experimental | medium | platform-api | Resource controller and CRUD route conventions | `go test -timeout 20s ./x/rest/...` | [beta evidence](extension-evidence/x-rest.md): release history, API snapshot, and owner sign-off missing |
| `x/tenant` | experimental | high | multitenancy | Resolution, policy, quota, rate limit, session, and tenant-aware stores | `go test -timeout 20s ./x/tenant/...` | [beta evidence](extension-evidence/x-tenant.md): release history, API snapshot, and owner sign-off missing |
| `x/websocket` | experimental | medium | realtime | WebSocket hub and explicit route registration | `go test -timeout 20s ./x/websocket/...` | [beta evidence](extension-evidence/x-websocket.md): release history, API snapshot, and owner sign-off missing |

## Subordinate Primitives

| Package | Parent family | Status | Risk | Owner | Recommended entrypoint | Validation | Blocker |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `x/cache` | `x/data` | experimental | medium | platform | Start from `x/data` unless cache topology is the direct task | `go test -timeout 20s ./x/cache/...` | Evaluate with data topology maturity |
| `x/ipc` | `x/gateway` | experimental | medium | edge | Start from `x/gateway` unless IPC transport is the direct task | `go test -timeout 20s ./x/ipc/...` | Subordinate edge primitive |
| `x/mq` | `x/messaging` | experimental | medium | messaging | Start from `x/messaging` unless queue primitive work is direct | `go test -timeout 20s ./x/mq/...` | Subordinate messaging primitive |
| `x/pubsub` | `x/messaging` | experimental | medium | messaging | Start from `x/messaging` unless pub/sub primitive work is direct | `go test -timeout 20s ./x/pubsub/...` | Subordinate messaging primitive |
| `x/scheduler` | `x/messaging` | experimental | medium | messaging | Start from `x/messaging` unless scheduling primitive work is direct | `go test -timeout 20s ./x/scheduler/...` | Subordinate messaging primitive |
| `x/webhook` | `x/messaging` | experimental | medium | integration | Start from `x/messaging` unless webhook transport is direct | `go test -timeout 20s ./x/webhook/...` | Subordinate messaging/integration primitive |

## Promotion Rule

Do not promote a module from `experimental` to `beta` from this dashboard alone.
Promotion requires:

- complete evidence in `specs/extension-beta-evidence.yaml`
- exported API snapshot comparison
- two consecutive minor release refs with no exported API changes
- owner sign-off
- updated module manifest, primer, and roadmap
