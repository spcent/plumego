# Module Matrix

This file is the high-level ownership matrix for the target repository.

## Stable Layer

| Module | Responsibility | Must Not Own |
|---|---|---|
| `core` | App lifecycle, bootstrap, route/middleware assembly | Feature packs, tenant logic, AI, webhook, websocket |
| `router` | Matching, params, groups, reverse routing | Response helpers, auth, business validation |
| `contract` | Errors, response helpers, request context contracts | Routing, persistence, app bootstrap |
| `middleware` | Transport-only HTTP middleware | Business policy, service injection, tenant config |
| `security` | JWT, headers, input safety, abuse primitives | App wiring, tenant policy orchestration |
| `store` | Stable persistence primitives and implementations | Tenant-aware adapters, business repositories |
| `health` | Health models and helpers | App features |
| `log` | Logging contracts and base implementations | Feature behavior |
| `metrics` | Metrics contracts and base collectors | App feature orchestration |

## Extension Layer

| Module | Responsibility | Status |
|---|---|---|
| `x/tenant` | Tenant resolution, quota, policy, rate limiting, tenant-aware adapters | Experimental |
| `x/ai` | AI gateway capabilities | Experimental |
| `x/websocket` | WebSocket capability pack | Experimental |
| `x/webhook` | Webhook capability pack | Experimental |
| `x/scheduler` | Scheduling capability pack | Experimental |
| `x/frontend` | Frontend/static asset capability pack | Experimental |
| `x/ops` | Operations capability pack | Experimental |
| `x/devtools` | Local development tools | Experimental |
| `x/messaging` | Canonical messaging capability pack; first discovery entrypoint for queue/pubsub workflows | Experimental |
| `x/discovery` | Service discovery capability pack | Experimental |
| `x/gateway` | Gateway and proxy capability pack | Experimental |

## Governance Notes

- Stable packages are the default target for routine HTTP toolkit iteration.
- Extension packages may depend on stable packages, but stable packages must not depend on extensions.
- Every stable package and every `x/*` package must define a `module.yaml`.
