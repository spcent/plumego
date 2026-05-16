# Extension Boundary

This document defines the structure and governance of Plumego's `x/*` extension
families. It explains what `x/*` is, how packages are organized within it, and
what the path to promotion looks like.

For the promotion criteria in detail, see `docs/EXTENSION_STABILITY_POLICY.md`.
For the current maturity status of each family, see `docs/EXTENSION_MATURITY.md`.
For machine-readable evidence tracking, see `specs/extension-beta-evidence.yaml`.

---

## What `x/*` Is Not

`x/*` is not a plugin market. Plugin markets have these properties:

- Packages are contributed by the community with varying quality standards.
- There is no common ownership or review expectation.
- Maturity is inferred from download count or age, not from explicit criteria.
- Adopting any package implies similar stability expectations.

Plumego's `x/*` does not work this way. Every extension family has an owner,
a defined responsibility boundary, a machine-readable manifest, and a stated
maturity level. Packages in `x/*` are **capability families with explicit
maturity levels and validation paths**.

The practical difference: when evaluating an `x/*` package, the first question
is not "does it exist?" but "what is its maturity status and what evidence
supports that status?"

---

## What `x/*` Is

`x/*` is the layer where product capability, protocol adaptation, and
business-domain features live — kept separate from the stable kernel to prevent
those concerns from stretching the kernel's compatibility promise.

The split serves two purposes:

1. **Stability isolation.** Stable roots carry a long-term compatibility
   promise. Extension families are allowed to evolve faster, which is
   appropriate for product capability and ecosystem-facing integrations.

2. **Scope clarity.** A new engineer or AI agent looking at a change can
   immediately know whether it touches the long-term stable kernel or an
   explicitly-bounded capability area.

---

## Four Categories of Extension Families

### A — App-facing capability families

Direct application entrypoints for common service patterns. These are the
primary extension discovery surface.

| Family | Primary role |
|---|---|
| `x/rest` | Resource route conventions, CRUD patterns, pagination |
| `x/websocket` | WebSocket hub, connection lifecycle, room broadcast |
| `x/gateway` | Reverse proxy, rewrite rules, backend pooling, edge transport |
| `x/tenant` | Tenant resolution, policy, quota, rate limiting, session, tenant-aware stores |
| `x/observability` | Exporter, tracer, collector, and adapter wiring |
| `x/fileapi` | HTTP file upload/download transport |
| `x/ai` | AI provider contracts, session, streaming, tool invocation |
| `x/messaging` | App-facing messaging flows |
| `x/frontend` | Static and embedded asset serving, SPA fallback |
| `x/resilience` | Circuit breaker and rate-limit primitives for extension-layer use |
| `x/observability/ops` | Protected admin routes, runtime diagnostics |

Start from the app-facing family before going deeper into subordinate
primitives.

### B — Supporting primitives

Lower-level building blocks consumed by category A families or by application
code that needs direct access to a specific primitive.

| Package | Parent family | Role |
|---|---|---|
| `x/data/cache` | `x/data` | Cache topology and Redis-backed implementations |
| `x/gateway/discovery` | `x/gateway` | Service discovery backend contracts and adapters |
| `x/messaging/mq` | `x/messaging` | Message queue primitive |
| `x/messaging/pubsub` | `x/messaging` | Pub/sub broker primitive |
| `x/messaging/scheduler` | `x/messaging` | In-process cron, delayed jobs, retryable tasks |
| `x/messaging/webhook` | `x/messaging` | Outbound webhook dispatch |
| `x/gateway/ipc` | `x/gateway` | Inter-process communication transport primitive |
| `x/data` | — | Persistence surface: file, idempotency, advanced topology |

Discovery starts from the owning family entrypoint, not from the subordinate
primitive directly.

### C — Local development and protected operations

Packages that must never be exposed in production without explicit auth gating.

| Package | Role |
|---|---|
| `x/observability/devtools` | Debug endpoints, request inspector, local-only diagnostics |
| `x/observability/ops` | Protected admin and runtime diagnostics routes |

Both require the caller to mount them explicitly under an authenticated path.
Neither self-registers or defaults to a public route.

### D — Fast-moving domain packs

Families where the domain itself evolves quickly and compatibility expectations
are correspondingly lighter.

| Package | Notes |
|---|---|
| `x/ai` | Stable-tier subpackages (`provider`, `session`, `streaming`, `tool`) are evaluated separately from the experimental root family |
| `x/frontend` | SPA and asset-serving patterns change with web tooling |
| `x/data` advanced topology | Sub-surface inventory selects specific stable targets |

For `x/ai`, evaluate subpackage maturity individually — do not treat the root
family status as applying to all subpackages.

---

## Maturity Ladder

All `x/*` packages start as `experimental`. Promotion is explicit and requires
meeting the criteria in `docs/EXTENSION_STABILITY_POLICY.md`.

| Status | Meaning | Adoption guidance |
|---|---|---|
| `experimental` | API shape may change; no compatibility expectation | Use for evaluation, prototyping, or when you own the upgrade risk |
| `beta` | API stable within current major version; breaking changes require deprecation | Acceptable for production use in clearly scoped scenarios |
| `ga` | Full v1 compatibility promise; follows `docs/DEPRECATION.md` | Default production choice |

Current status of all families: `docs/EXTENSION_MATURITY.md`.

---

## Promotion Criteria Summary

### `experimental` → `beta`

All of the following must be true:

1. No exported symbol changes for at least two consecutive minor releases.
2. Passes `go run ./internal/checks/dependency-rules` with no violations.
3. Unit tests cover all documented public behavior paths including negative paths,
   and run cleanly with `go test -race ./...`.
4. `module.yaml` is complete and schema-valid.
5. A primer document exists in `docs/modules/<family>/README.md`.
6. A beta evidence document exists in `docs/extension-evidence/<family>.md`.
7. Owner sign-off is recorded in `specs/extension-beta-evidence.yaml`.

### `beta` → `ga`

All of the following must be true:

1. `beta` status has been held across at least two minor releases with no
   exported API changes.
2. Multiple concrete reference uses exist (reference apps or real production
   use).
3. No outstanding known bugs that affect documented behavior.
4. Full compatibility and deprecation policy aligned with `docs/DEPRECATION.md`.
5. Security, concurrency, and boundary tests are comprehensive.

Promotion is recorded in a dedicated promotion card in `tasks/cards/`. It is
never done inline with feature work.

---

## Adoption Sequence

When evaluating an `x/*` extension family:

1. Check `docs/EXTENSION_MATURITY.md` for current status and recommended
   entrypoint.
2. Read the primer at `docs/modules/<family>/README.md`.
3. Read the `module.yaml` for responsibilities and non-goals.
4. Start from `reference/with-<family>` for wiring guidance.
5. Check `specs/extension-beta-evidence.yaml` for known blockers before
   committing to production adoption.

Do not infer production readiness from the existence of a package alone.

---

## Non-goals for `x/*`

Extension families are not allowed to:

- Import stable roots in ways that force stable-root changes to accommodate
  them.
- Self-register routes, middleware, or lifecycle hooks without explicit caller
  mounting.
- Define alternate canonical application layouts.
- Promote themselves by changing their own `module.yaml` status.

Boundary violations are caught automatically:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/extension-maturity
go run ./internal/checks/extension-beta-evidence
```

---

## Adding a New Extension Family

New extension families start from the template in
`specs/change-recipes/new-extension-module.yaml`.

Required before the first commit:

- `module.yaml` with `status: experimental`, `owner`, `responsibilities`,
  `non_goals`, and `test_commands`
- `docs/modules/<family>/README.md` primer
- Entry in `specs/extension-taxonomy.yaml`
- Entry in `specs/extension-maturity.yaml`

Required before any beta promotion attempt: a complete evidence document at
`docs/extension-evidence/<family>.md` using the template in
`docs/extension-evidence/BETA_EVIDENCE_TEMPLATE.md`.
