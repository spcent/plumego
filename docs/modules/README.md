# Plumego Modules Index

This page is the canonical entry for module docs.

- v1 API freeze: `docs/other/V1_CORE_API_FREEZE.md`
- style baseline: `docs/CANONICAL_STYLE_GUIDE.md`
- quick start: `docs/getting-started.md`
- top-level docs index: `docs/README.md`

## Repository Layers

- `Stable`: `core`, `router`, `middleware`, `contract`, `security`, `store`, `health`, `log`, `metrics`
- `Extension`: `x/*`
- `Reference`: `reference/*`
- `Examples`: `examples/*`

## Read Path (Recommended)

1. [core](core/README.md)
2. [router](router/README.md)
3. [middleware](middleware/README.md)
4. [contract](contract/README.md)
5. [security](security/README.md)
6. [store](store/README.md)
7. [health](health/README.md)
8. [log](log/README.md)
9. [metrics](metrics/README.md)

## Stable Modules

- [core](core/README.md): app lifecycle, boot, options, component/runner hooks
- [router](router/README.md): route matching, groups, params, reverse routing
- [middleware](middleware/README.md): transport middleware chain and ordering
- [contract](contract/README.md): request context, response/error contracts
- [security](security/README.md): JWT, headers, abuse/input defenses
- [store](store/README.md): cache/db/file/kv/idempotency abstractions
- [health](health/README.md): readiness and health primitives
- [log](log/README.md): logging contracts and base implementations
- [metrics](metrics/README.md): metrics contracts and collectors

## Extension Modules

- [x/tenant](x-tenant/README.md): target multi-tenant extension boundary
- [x/ai](x-ai/README.md): AI gateway capabilities
- [x/gateway](x-gateway/README.md): gateway transport and protocol extensions
- [x/ops](x-ops/README.md): protected diagnostics and operations HTTP surfaces
- [x/data](x-data/README.md): data topology extensions beyond the stable store layer
- [x/webhook](x-webhook/README.md): webhook ingress/egress extension surface
- [x/websocket](x-websocket/README.md): websocket extension surface

## Legacy Roots Being Retired

- [config](config/README.md): env + runtime config loading
- [validator](validator/README.md): request/field validators
- [frontend](frontend/README.md): static/embedded asset mounting
- [net](net/README.md): networking helpers (webhook/websocket/discovery/mq)
- [scheduler](scheduler/README.md): cron/delayed/retry task scheduling
- [tenant](tenant/README.md): experimental multi-tenant primitives

## Notes

- Prefer explicit root stable package APIs in production code and docs.
- Treat `x/*` as opt-in capability packs.
- Historical planning docs should not override the target layout described in `docs/architecture/`.
- Archived topic writeups, plans, and reports now live under `docs/legacy/`.
