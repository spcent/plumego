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
`module.yaml`. Dashboard-only signals such as recommended entrypoints, docs
state, and coverage state come from `specs/extension-maturity.yaml`. For beta
candidates, it also verifies the evidence link and blocker text against
`specs/extension-beta-evidence.yaml`.

## Publishing Extensions

Community-authored extensions use the separate contract in
[`docs/EXTENSION_AUTHORING.md`](EXTENSION_AUTHORING.md). That guide covers
`community-extension.yaml`, `plumego add`, explicit constructor ownership, and
the compatibility checklist for modules published outside this repository. This
dashboard remains the maturity view for first-party `x/*` families.

## Release History

`x/rest`, `x/websocket`, `x/gateway`, and `x/observability` were promoted to
beta at d2c25c3–ec70358. All four modules showed no exported API changes across
both release refs. At v1.1.0, `x/tenant` was promoted to beta, and selected
surfaces under `x/ai` and `x/data` gained beta evidence without promoting their
root families. The same release promoted `x/frontend` and the app-facing
`x/messaging` service while leaving subordinate messaging primitives
experimental. Release-backed API snapshots are recorded in
`docs/extension-evidence/snapshots/`. Remaining `x/*` modules retain
`experimental` status until their own release evidence is complete.

## Selected Beta Surfaces

| Surface | Parent | Status | Owner | Release refs | Evidence |
| --- | --- | --- | --- | --- | --- |
| `x/ai/provider` | x/ai | beta surface | ai-gateway | `v1.0.0`, `v1.1.0` | [beta evidence](extension-evidence/x-ai-provider.md): API unchanged; owner sign-off recorded |
| `x/ai/session` | x/ai | beta surface | ai-gateway | `v1.0.0`, `v1.1.0` | [beta evidence](extension-evidence/x-ai-session.md): API unchanged; owner sign-off recorded |
| `x/ai/streaming` | x/ai | beta surface | ai-gateway | `v1.0.0`, `v1.1.0` | [beta evidence](extension-evidence/x-ai-streaming.md): API unchanged; owner sign-off recorded |
| `x/ai/tool` | x/ai | beta surface | ai-gateway | `v1.0.0`, `v1.1.0` | [beta evidence](extension-evidence/x-ai-tool.md): API unchanged; owner sign-off recorded |
| `x/data/file` | x/data | beta surface | persistence | `v1.0.0`, `v1.1.0` | [beta evidence](extension-evidence/x-data.md): API unchanged; owner sign-off recorded |
| `x/data/idempotency` | x/data | beta surface | persistence | `v1.0.0`, `v1.1.0` | [beta evidence](extension-evidence/x-data.md): API unchanged; owner sign-off recorded |

## App-Facing Families

| Family | Status | Risk | Owner | Recommended entrypoint | Signals | Validation | Evidence / blocker |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `x/ai` | experimental | high | ai-gateway | `x/ai/provider`, `x/ai/session`, `x/ai/streaming`, `x/ai/tool` for beta-surface adoption | docs:primer; coverage:stable-tier | `go test -timeout 20s ./x/ai/...` | Root family is not beta-ready; stable-tier subpackages have beta evidence at v1.1.0 |
| `x/data` | experimental | medium | persistence | `docs/modules/x-data/README.md` before subordinate data packages | docs:primer; coverage:sub-surface-inventory | `go test -timeout 20s ./x/data/...` | [maturity note](extension-evidence/x-data.md): `x/data/file` and `x/data/idempotency` are beta surfaces at v1.1.0; topology surfaces remain experimental |
| `x/observability/devtools` | experimental | medium | observability | Explicit local/protected debug mounting only | docs:primer; coverage:debug-surface | `go test -timeout 20s ./x/observability/devtools/...` | Debug tooling; not a production admin surface |
| `x/gateway/discovery` | experimental | medium | edge | Caller-selected discovery backend | docs:primer; coverage:backend-tests | `go test -timeout 20s ./x/gateway/discovery/...` | [maturity note](extension-evidence/x-discovery.md): core/static surface blocked because `x/gateway/discovery` did not exist at `v1.0.0`; path-migration evidence is required |
| `x/fileapi` | experimental | medium | persistence | HTTP file transport over `x/data/file` and stable `store/file` contracts | docs:primer; coverage:transport-tests | `go test -timeout 20s ./x/fileapi/...` | Needs persistence and transport scenario evidence as behavior expands |
| `x/frontend` | beta | medium | frontend | Explicit static or embedded asset serving | docs:primer; coverage:asset-serving; hardening:directory-safety/precompressed/negotiation | `go test -timeout 20s ./x/frontend/...` | [beta evidence](extension-evidence/x-frontend.md): promoted at v1.1.0; API unchanged across `v1.0.0` to `v1.1.0` |
| `x/gateway` | beta | medium | edge | `x/gateway` for proxy, rewrite, balancing, and edge transport | docs:primer; coverage:edge-tests | `go test -timeout 20s ./x/gateway/...` | [beta evidence](extension-evidence/x-gateway.md): promoted at d2c25c3–ec70358; API unchanged across both refs |
| `x/messaging` | beta | medium | messaging | App-facing messaging entrypoint before queue/pubsub primitives | docs:primer; coverage:sub-surface-inventory | `go test -timeout 20s ./x/messaging/...` | [beta evidence](extension-evidence/x-messaging.md): app-facing service promoted at v1.1.0; mq/pubsub/scheduler/webhook primitives remain experimental |
| `x/observability` | beta | medium | observability | Exporter, tracer, collector, and adapter wiring | docs:primer; coverage:exporter-tracer-tests | `go test -timeout 20s ./x/observability/...` | [beta evidence](extension-evidence/x-observability.md): promoted at d2c25c3–ec70358; API unchanged across both refs |
| `x/openapi` | experimental | medium | api | RouteInfo-driven OpenAPI document generation | docs:primer; coverage:generator-tests | `go test -timeout 20s ./x/openapi/...` | OpenAPI 3.1 generator with dependency-free JSON/YAML serialization and `plumego generate spec` CLI wiring |
| `x/observability/ops` | experimental | medium | observability | Protected admin and runtime diagnostics routes | docs:primer; coverage:protected-ops | `go test -timeout 20s ./x/observability/ops/...` | Requires explicit auth boundary; not a public diagnostics default |
| `x/resilience` | experimental | medium | runtime | Reusable extension-layer circuit breaker and rate-limit primitives | docs:primer; coverage:runtime-primitive | `go test -timeout 20s ./x/resilience/...` | Cross-family primitive; needs adoption evidence before promotion |
| `x/rest` | beta | medium | platform-api | Resource controller and CRUD route conventions | docs:primer; coverage:crud-tests | `go test -timeout 20s ./x/rest/...` | [beta evidence](extension-evidence/x-rest.md): promoted at d2c25c3–ec70358; API unchanged across both refs |
| `x/rpc` | experimental | medium | rpc | Optional gRPC server, client, and gateway adapters | docs:primer; coverage:server-lifecycle | `go test -timeout 20s ./x/rpc/...` | Optional RPC transports are caller-owned adapters outside the Plumego core surface |
| `x/tenant` | beta | high | multitenancy | Resolution, policy, quota, rate limit, session, and tenant-aware stores | docs:primer; coverage:tenant-chain-tests | `go test -timeout 20s ./x/tenant/...` | [beta evidence](extension-evidence/x-tenant.md): promoted at v1.1.0; API unchanged across `v1.0.0` to `v1.1.0` |
| `x/validate` | experimental | medium | api | Explicit `Bind` or `BindJSON` call sites in handlers | docs:primer; coverage:binding-tests | `go test -timeout 20s ./x/validate/...` | Validation bridge has no third-party adapter or tag parser in the package |
| `x/websocket` | beta | medium | realtime | WebSocket hub and explicit route registration | docs:primer; coverage:hub-lifecycle-tests | `go test -timeout 20s ./x/websocket/...` | [beta evidence](extension-evidence/x-websocket.md): promoted at d2c25c3–ec70358; API unchanged across both refs |

## Subordinate Primitives

| Package | Parent family | Status | Risk | Owner | Recommended entrypoint | Signals | Validation | Blocker |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `x/data/cache` | `x/data` | experimental | medium | platform | Start from `x/data` unless cache topology is the direct task | docs:primer; coverage:topology | `go test -timeout 20s ./x/data/cache/...` | Evaluate with data topology maturity |
| `x/gateway/ipc` | `x/gateway` | experimental | medium | edge | Start from `x/gateway` unless IPC transport is the direct task | docs:primer; coverage:transport-primitive | `go test -timeout 20s ./x/gateway/ipc/...` | Subordinate edge primitive |
| `x/messaging/mq` | `x/messaging` | experimental | medium | messaging | Start from `x/messaging` unless queue primitive work is direct | docs:primer; coverage:queue-primitive | `go test -timeout 20s ./x/messaging/mq/...` | Subordinate messaging primitive |
| `x/messaging/pubsub` | `x/messaging` | experimental | medium | messaging | Start from `x/messaging` unless pub/sub primitive work is direct | docs:primer; coverage:broker-primitive | `go test -timeout 20s ./x/messaging/pubsub/...` | Subordinate messaging primitive |
| `x/messaging/scheduler` | `x/messaging` | experimental | medium | messaging | Start from `x/messaging` unless scheduling primitive work is direct | docs:primer; coverage:scheduler-primitive | `go test -timeout 20s ./x/messaging/scheduler/...` | Subordinate messaging primitive |
| `x/messaging/webhook` | `x/messaging` | experimental | medium | integration | Start from `x/messaging` unless webhook transport is direct | docs:primer; coverage:webhook-primitive | `go test -timeout 20s ./x/messaging/webhook/...` | Subordinate messaging/integration primitive |

## Promotion Rule

Do not promote a module from `experimental` to `beta` from this dashboard alone.
Promotion requires:

- complete evidence in `specs/extension-beta-evidence.yaml`
- exported API snapshot comparison
- two consecutive minor release refs with no exported API changes
- owner sign-off
- updated module manifest, primer, and roadmap
