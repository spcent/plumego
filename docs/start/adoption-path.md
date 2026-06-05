# Adoption Path

This page is the narrow adoption path for new Plumego users.

The first impression should be simple: standard-library HTTP compatibility,
small stable kernel, explicit wiring, and machine-checkable agent workflow.
Optional `x/*` packages are capability additions, not alternate bootstraps.

## Migrating From Another Framework

Read the focused migration guide for your current stack, then return to the
standard service path:

| Current stack | First read |
| --- | --- |
| Gin | `docs/guides/migration/from-gin.md` |
| Echo | `docs/guides/migration/from-echo.md` |
| Chi | `docs/guides/migration/from-chi.md` |
| Existing middleware stack | `docs/guides/migration/middleware-compat.md` |

## 5 Minutes: Run One Standard Service

Read:

1. `docs/start/getting-started.md`
2. `reference/standard-service/README.md`

Use:

- `core.DefaultConfig`
- `core.New`
- `app.Use`
- `app.Get`
- `contract.WriteResponse`
- `contract.WriteError`

Avoid:

- `x/*` packages
- custom response envelopes
- hidden route registration

Done when a local service responds from one explicit route using
`http.HandlerFunc`.

## 30 Minutes: Add One Capability

Keep `reference/standard-service` as the app shape. Add one capability family:

| Need | First read |
| --- | --- |
| CRUD/resource conventions | `docs/modules/x/rest/README.md` |
| Tenant resolution and isolation | `docs/modules/x/tenant/README.md` |
| Edge proxy and rewrite | `docs/modules/x/gateway/README.md` |
| Dynamic service discovery behind gateway or clients | `docs/modules/x/gateway/README.md` (`x/gateway/discovery` section) |
| WebSocket transport | `docs/modules/x/websocket/README.md` |
| Messaging flows, queues, and webhook delivery | `docs/modules/x/messaging/README.md` |
| Inbound webhook verification or outbound webhook transport | `docs/modules/x/messaging/README.md` (`x/messaging/webhook` section) |
| File upload, download, and temporary URL transport | `docs/modules/x/fileapi/README.md` |
| Observability exporters | `docs/modules/x/observability/README.md` |
| Local debug routes, config snapshots, or pprof | `docs/modules/x/observability/devtools/README.md` |
| Reusable circuit breaker or rate-limit primitives | `docs/modules/x/resilience/README.md` |
| AI provider/session/tool primitives | `docs/modules/x/ai/README.md` |
| gRPC service hosting, client pooling, or HTTP-RPC gateway | `docs/modules/x/rpc/README.md` |
| OpenAPI 3.1 spec generation from route metadata | `docs/modules/x/openapi/README.md` |

Done when the app still has explicit route wiring and only the selected
capability family is added.

## 1 Day: Understand The Control Plane

Read:

1. `AGENTS.md` (repository structure and agent operating model)
2. `docs/operations/agent-context-budget.md`
3. `docs/operations/codex-workflow.md`
4. `docs/reference/canonical-style-guide.md`
5. `specs/task-routing.yaml`
6. `specs/repo.yaml`
7. `specs/dependency-rules.yaml`
8. the target module's `module.yaml`

Done when a developer or agent can answer:

- which module owns the change
- which paths are out of scope
- whether stable public APIs are affected
- which validation commands must pass

## Messaging Rule

When updating README, website, or onboarding docs, present Plumego in this
order:

1. standard-library compatibility
2. small stable kernel
3. explicit reference application path
4. machine-checkable agent workflow
5. optional experimental capability catalog
