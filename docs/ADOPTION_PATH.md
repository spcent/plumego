# Adoption Path

This page is the narrow adoption path for new Plumego users.

The first impression should be simple: standard-library HTTP compatibility,
small stable kernel, explicit wiring, and machine-checkable agent workflow.
Optional `x/*` packages are capability additions, not alternate bootstraps.

## 5 Minutes: Run One Standard Service

Read:

1. `docs/getting-started.md`
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
| CRUD/resource conventions | `docs/modules/x-rest/README.md` |
| Tenant resolution and isolation | `docs/modules/x-tenant/README.md` |
| Edge proxy and rewrite | `docs/modules/x-gateway/README.md` |
| WebSocket transport | `docs/modules/x-websocket/README.md` |
| Observability exporters | `docs/modules/x-observability/README.md` |
| AI provider/session/tool primitives | `docs/modules/x-ai/README.md` |

Done when the app still has explicit route wiring and only the selected
capability family is added.

## 1 Day: Understand The Control Plane

Read:

1. `AGENTS.md`
2. `docs/CODEX_WORKFLOW.md`
3. `docs/CANONICAL_STYLE_GUIDE.md`
4. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
5. `specs/repo.yaml`
6. `specs/dependency-rules.yaml`
7. the target module's `module.yaml`

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
