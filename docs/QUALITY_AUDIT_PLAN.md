# Quality Audit Plan

This plan covers the medium-term quality work needed after stable-root freeze
and pre-v1 release evidence are in place.

## Benchmark Evidence

Add benchmark evidence without changing runtime behavior:

- `router`: static route, param route, route group, reverse URL generation
- `middleware`: chain build, chain execution, timeout/recovery overhead
- `core`: prepare/server handler construction and route registration

Compare against a minimal `net/http` baseline where useful. The goal is not to
claim fastest performance; it is to publish credible expectations.

## Negative-Path Audit

Prioritize fail-closed behavior:

| Area | Cases |
| --- | --- |
| `security` | invalid token, expired token, malformed signature, timing-safe comparison |
| `x/tenant` | missing tenant, denied policy, quota exceeded, cross-tenant isolation |
| `x/gateway` | bad backend, rewrite mismatch, timeout, context cancellation |
| `middleware` | panic recovery, timeout write path, body limit, auth short-circuit |

Each audit should produce tests next to the changed or verified behavior.

## AI Surface Split

Keep `x/ai` experimental at the family level. Evaluate stable-tier subpackages
individually:

- `x/ai/provider`
- `x/ai/session`
- `x/ai/streaming`
- `x/ai/tool`

Do not let orchestration, marketplace, semantic cache, distributed execution,
or resilience wrappers imply stability for the whole family.

## Adoption Review

Before a release candidate, review README and website content against
`docs/ADOPTION_PATH.md`. The first user path should be:

1. run the standard service
2. add one capability
3. understand the control plane

Feature catalogs should come after the adoption path.
