# Extension Stability Policy

This policy defines the compatibility guarantees for stable root packages and
the status levels used by `x/*` extension families. It is the authoritative
source referenced by `docs/concepts/extension-maturity.md`.

## Stable Roots (General Availability)

The following nine packages are **stable roots**. Their APIs carry a full `v1`
compatibility guarantee: signatures, package names, and behaviour are frozen for
the `v1.x` release series. Breaking changes require a major version bump.

| Package | Role |
|---|---|
| `core` | App construction, route registration, middleware attachment, server lifecycle. |
| `router` | Route matching, path params, groups, metadata, reverse URL generation. |
| `contract` | Response writers, structured error builders, request metadata, transport binding. |
| `middleware` | Transport-only middleware composition and first-party packages. |
| `security` | Auth, JWT, password, security headers, input safety, abuse guards. |
| `store` | Stable storage contracts and in-memory primitives. |
| `health` | Health and readiness models for app and dependency status. |
| `log` | Minimal logging interfaces and a default logger. |
| `metrics` | Minimal metrics contracts (counters, gauges, timings, collectors). |

## Extension Status Levels

### `experimental`

- **API may change in any minor version without notice.**
- Not recommended for production use without explicit project-level stabilization.
- May be promoted to `beta` once release evidence requirements are met (see
  `specs/extension-beta-evidence.yaml`).

**Experimental at v1.x:**
`x/ai`, `x/data`, `x/fileapi`, `x/frontend`, `x/messaging`, `x/messaging/mq`,
`x/messaging/pubsub`, `x/messaging/scheduler`, `x/messaging/webhook`,
`x/openapi`, `x/resilience`, `x/rpc`, `x/tenant`, `x/validate`,
`x/gateway/discovery`, `x/gateway/ipc`, `x/observability/devtools`,
`x/observability/ops`

### `beta`

- API is stable across the release refs cited in the beta evidence doc.
- Breaking changes require a deprecation notice in the same PR.
- Suitable for adoption with the understanding that edge cases may still be rough.
- May be promoted to `ga` once the module has two production release cycles
  with no exported API changes and full test coverage at the stable tier.

**Beta at v1.x:**
`x/gateway`, `x/observability`, `x/rest`, `x/websocket`

### `ga` (General Availability)

- API follows the same compatibility guarantee as stable roots.
- Breaking changes require a major version bump or explicit migration path.
- Only stable roots currently carry `ga` status; extensions reach `ga` via the
  promotion criteria in `specs/extension-maturity.yaml`.

## SemVer Expectations

| Surface | v1.x guarantee | Breaking-change path |
|---|---|---|
| Stable roots | No breaking changes within `v1` | Major version bump (`v2+`) |
| Beta extensions | Stable across cited release refs; deprecation notice required before removal | Deprecation notice → removal in a subsequent minor |
| Experimental extensions | No guarantee; may change in any minor version | None required |
| CLI (`cmd/plumego`) | Command-line interface stable; **not** a Go import surface | Minor version notice |

## What You Can Safely Depend on in v1

**Safe to depend on:**
- All 9 stable root packages — import paths, exported types, functions, and interfaces.
- Beta extension packages (`x/gateway`, `x/observability`, `x/rest`, `x/websocket`) — production-ready with minor caveats; check the beta evidence doc for cited release refs.

**Do not treat as stable:**
- Experimental `x/*` extensions — APIs may change in any minor version.
- Internal packages (`internal/`) — private implementation detail.
- `cmd/plumego` as a Go import surface — it is a CLI tool only.

## Promotion Criteria

| From | To | Requirements |
|---|---|---|
| experimental | beta | Two consecutive release refs with no exported API changes; owner sign-off; beta evidence recorded in `docs/evidence/extension/`. |
| beta | ga | Two production release cycles; full coverage at stable tier; boundary review sign-off; entry in `docs/evidence/stable-api/`. |

## Deprecation Within a Status Level

Deprecated symbols must be removed in the same PR that migrates their last
caller. No dead wrappers may remain at merge time. This applies at all maturity
levels.
