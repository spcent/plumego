# Plumego Modules Index (v1 Canonical)

This page is the canonical entry for module docs.

- v1 API freeze: `docs/other/V1_CORE_API_FREEZE.md`
- style baseline: `docs/CANONICAL_STYLE_GUIDE.md`
- quick start: `docs/getting-started.md`

## Stability Tiers

- `GA (v1.x compatibility)`: `core`, `router`, `middleware`, `contract`, `security`, `store`
- `Compatibility`: convenience exports in top-level `plumego` package
- `Experimental (not GA-stable in v1.0)`: `tenant/*`, `net/mq/*`

## Read Path (Recommended)

1. [core](core/README.md)
2. [router](router/README.md)
3. [middleware](middleware/README.md)
4. [contract](contract/README.md)
5. [security](security/README.md)
6. [store](store/README.md)

## GA Modules

- [core](core/README.md): app lifecycle, boot, options, component/runner hooks
- [router](router/README.md): route matching, groups, params, reverse routing
- [middleware](middleware/README.md): transport middleware chain and ordering
- [contract](contract/README.md): request context, response/error contracts
- [security](security/README.md): JWT, headers, abuse/input defenses
- [store](store/README.md): cache/db/file/kv/idempotency abstractions

## Other Modules

- [config](config/README.md): env + runtime config loading
- [health](health/README.md): readiness/build/health primitives
- [metrics](metrics/README.md): metrics/tracing adapters
- [log](log/README.md): structured logging interfaces
- [validator](validator/README.md): request/field validators
- [frontend](frontend/README.md): static/embedded asset mounting
- [net](net/README.md): networking helpers (webhook/websocket/discovery/mq)
- [scheduler](scheduler/README.md): cron/delayed/retry task scheduling
- [ai](ai/README.md): AI gateway capabilities
- [tenant](tenant/README.md): experimental multi-tenant primitives

## Notes

- Prefer explicit package APIs (`core`, `router`, `middleware`, `contract`) in production code and docs.
- Historical planning docs are archived and should not override v1 canonical guidance.
