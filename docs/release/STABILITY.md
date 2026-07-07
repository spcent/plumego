# Stability and Compatibility Guarantees

This document defines what is stable and supported in Plumego v1, and what may change.

## Stable Roots (v1 Guarantee)

The following **9 packages** carry a full v1 compatibility guarantee: their APIs, package names, and behavior are frozen for the `v1.x` release series. Breaking changes require a major version bump to `v2`.

| Package | Role | Status | Guarantee |
|---------|------|--------|-----------|
| `core` | App construction, lifecycle, routes, middleware attachment | **GA** | v1 guarantee |
| `router` | Route matching, path params, groups, metadata, reverse URLs | **GA** | v1 guarantee |
| `contract` | Response writers, error builders, request metadata, transport binding | **GA** | v1 guarantee |
| `middleware` | Transport-layer middleware composition + standard packages | **GA** | v1 guarantee |
| `security` | Auth adapters, JWT, password hashing, abuse guards | **GA** | v1 guarantee |
| `store` | Storage contracts + in-memory primitives (cache, KV, file, idempotency) | **GA** | v1 guarantee |
| `health` | Health/readiness models and status propagation | **GA** | v1 guarantee |
| `log` | Logging interfaces + default logger | **GA** | v1 guarantee |
| `metrics` | Metrics contracts, collectors, and adapters | **GA** | v1 guarantee |

**What this means:**
- ✅ You can depend on these packages in production code
- ✅ Exported types, functions, and interfaces will not change
- ✅ Behavior is stable (no silent breaking changes)
- ❌ Breaking changes require `v2.0.0` or a clear migration path in `v1.x`
- ❌ Symbols will not be removed without deprecation notices

## Beta Extensions (Production-Ready with Caveats)

The following **7 extension families** are **beta**: their APIs are stable across cited release references and suitable for production adoption with the understanding that edge cases may still be rough.

| Family | Status | Release Refs | Note |
|--------|--------|--------------|------|
| `x/frontend` | beta | v1.0.0, v1.1.0 | Asset serving, static + embedded | [Evidence](docs/evidence/extension/x-frontend.md) |
| `x/gateway` | beta | v1.0.0, v1.1.0 | Proxy, rewrite, balancing, edge transport | [Evidence](docs/evidence/extension/x-gateway.md) |
| `x/messaging` | beta | v1.0.0, v1.1.0 | App-facing messaging service (mq/pubsub/scheduler/webhook subordinates are experimental) | [Evidence](docs/evidence/extension/x-messaging.md) |
| `x/observability` | beta | v1.0.0, v1.1.0 | Exporter, tracer, collector, adapter wiring | [Evidence](docs/evidence/extension/x-observability.md) |
| `x/rest` | beta | v1.0.0, v1.1.0 | Resource controller and CRUD route conventions | [Evidence](docs/evidence/extension/x-rest.md) |
| `x/tenant` | beta | v1.0.0, v1.1.0 | Resolution, policy, quota, rate limit, session, tenant-aware stores | [Evidence](docs/evidence/extension/x-tenant.md) |
| `x/websocket` | beta | v1.0.0, v1.1.0 | WebSocket hub and explicit route registration | [Evidence](docs/evidence/extension/x-websocket.md) |

**What this means:**
- ✅ Suitable for production adoption
- ✅ APIs are stable across cited release references
- ✅ Breaking changes require a deprecation notice in the same PR
- ⚠️ Edge cases may still be rough; use with awareness
- ❌ Not guaranteed to be feature-complete
- ❌ May change in `v1.x` if a deprecation notice is provided

**Beta surfaces within experimental families:**

The following subpackages have beta-tier evidence even though their parent families are experimental:

| Subpackage | Parent | Status | Release Refs |
|------------|--------|--------|--------------|
| `x/ai/provider` | `x/ai` | beta surface | v1.0.0, v1.1.0 |
| `x/ai/session` | `x/ai` | beta surface | v1.0.0, v1.1.0 |
| `x/ai/streaming` | `x/ai` | beta surface | v1.0.0, v1.1.0 |
| `x/ai/tool` | `x/ai` | beta surface | v1.0.0, v1.1.0 |
| `x/data/file` | `x/data` | beta surface | v1.0.0, v1.1.0 |
| `x/data/idempotency` | `x/data` | beta surface | v1.0.0, v1.1.0 |

Use these subpackages with the same confidence as beta families; the parent family remains experimental.

## Experimental Extensions (APIs May Change)

The following extension families are **experimental**: their APIs may change in any minor version without notice. Do not use them in production services without explicit project-level stabilization.

| Family | Note |
|--------|------|
| `x/ai` (root) | Stable subpackages exist (provider, session, streaming, tool); root family is experimental |
| `x/data` (root) | Stable subpackages exist (file, idempotency); root and other subpackages are experimental |
| `x/fileapi` | HTTP file transport over `x/data/file` and stable `store/file` contracts |
| `x/openapi` | OpenAPI 3.1 spec generation from route metadata |
| `x/resilience` | Reusable extension-layer circuit breaker and rate-limit primitives |
| `x/rpc` | Optional gRPC server, client, and gateway adapters |
| `x/validate` | Explicit `Bind` or `BindJSON` call sites in handlers |
| `x/gateway/discovery` | Caller-selected discovery backend for gateway |
| `x/gateway/ipc` | IPC transport primitive for gateway |
| `x/observability/devtools` | Local/protected debug routes and profiling |
| `x/observability/ops` | Protected admin and runtime diagnostics routes |
| Subordinate messaging primitives | `x/messaging/mq`, `x/messaging/pubsub`, `x/messaging/scheduler`, `x/messaging/webhook` |
| Subordinate data packages | `x/data/cache`, `x/data/sharding`, `x/data/migrate`, and others |

**What this means:**
- ✅ Safe to use for learning and exploration
- ⚠️ APIs may change in any `v1.x` minor version without notice
- ❌ Do not depend on experimental APIs in production code without accepting risk
- ❌ No deprecation notice required before removal or breaking changes

See `docs/concepts/extension-maturity.md` for the full maturity dashboard and promotion criteria.

## SemVer Expectations

| Surface | v1.x guarantee | Breaking-change path |
|---------|---|---|
| Stable roots (9 packages) | No breaking changes within `v1` | Major version bump (`v2+`) |
| Beta extensions (7 families) | Stable across cited release refs; deprecation notice required before removal | Deprecation notice → removal in a subsequent minor |
| Beta surfaces (within experimental families) | Stable across cited release refs; deprecation notice required before removal | Deprecation notice → removal in a subsequent minor |
| Experimental extensions | No guarantee; may change in any minor version | None required |
| CLI (`cmd/plumego`) | Command-line interface stable; **not** a Go import surface | Minor version notice in release notes |

## What you can safely depend on in v1

### Safe to depend on:

- ✅ All 9 stable root packages and their exported symbols
- ✅ All 7 beta extension families
- ✅ The 6 beta subpackages within experimental families (`x/ai/provider`, `x/ai/session`, `x/ai/streaming`, `x/ai/tool`, `x/data/file`, `x/data/idempotency`)
- ✅ `reference/` applications as architectural examples (not versioned, but stable across patch releases)

### Do not treat as stable:

- ❌ Experimental `x/*` families (APIs may change in any minor version)
- ❌ Internal packages (`internal/`) — private implementation detail
- ❌ `cmd/plumego` as a Go import surface — it is a CLI tool only
- ❌ Test helpers or examples as part of the public API

## Deprecation Policy

When a stable root or beta extension API needs to change, we follow this process:

1. **PR includes deprecation notice** — The breaking change is noted in the PR description and commit message
2. **Symbol is marked as deprecated** — If possible, the old symbol receives a `// Deprecated:` doc comment
3. **Migration path is documented** — A guide explaining how to update code is added to `COMPATIBILITY.md`
4. **One minor version with deprecation** — The old symbol remains in the code through at least one full minor release
5. **Removal in a subsequent minor** — The symbol is removed in a later `v1.x` or in `v2.0`

Example: If `v1.3.0` deprecates `core.NewWithoutDefaults()`, it will remain in `v1.3.x` and `v1.4.x`, and be removed in `v1.5.0` or later.

## Supported versions

| Version | Status | Security updates |
|---------|--------|-------------------|
| v1.2.x+ | Active | ✅ Yes |
| v1.1.x | Maintenance | ✅ Yes |
| v1.0.x | Maintenance | ✅ Yes |
| < v1.0 | Unsupported | ❌ No |

## How to report issues

- **Bug in stable root or beta extension:** Open an issue. Include the version, minimal reproduction, and expected behavior. Bugs in stable surfaces are high priority.
- **API request or extension idea:** Open a discussion or issue. We review these carefully before committing to stable roots.
- **Security vulnerability:** See `SECURITY.md` for responsible disclosure.

## Release notes

Every release includes:
- List of stable root and beta extension changes
- Clear separation of stable/beta/experimental changes
- Migration notes for any deprecations
- New experimental features noted as "may change"

See `docs/release/` for version history and roadmap.

---

For more details on extension maturity levels, see `docs/concepts/extension-maturity.md`.  
For upgrade paths and forward compatibility, see `COMPATIBILITY.md`.
