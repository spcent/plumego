# Card 0241

Priority: P1
State: done
Primary Module: store
Owned Files:
- `store/cache/cache.go`
- `store/cache/cache_test.go`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `x/cache`
- `x/observability`

Goal:
- Keep stable `store/cache` focused on cache primitives, not cache metrics tracking or introspection payload ownership.
- Move cache metrics and inspection ownership to an extension package that actually consumes or exports them.

Problem:
- Stable `store/cache` still exports `MetricsTracker`, `MetricsSnapshot`, `GetMetrics()`, and `EnableMetrics`.
- That makes the stable cache primitive own observability state and stats payload shape, which conflicts with the store boundary that keeps analytics and instrumentation outside stable roots.
- The cache implementation is currently both a persistence primitive and its own metrics/reporting package.

Scope:
- Remove exported cache metrics tracker and snapshot ownership from stable `store/cache`.
- Keep stable cache behavior limited to cache operations and lifecycle needed by the in-process primitive itself.
- Move any remaining cache metrics/introspection surface to an owning extension package and update downstream callers/tests in the same change.
- Sync store docs and manifest language to the reduced cache boundary.

Non-goals:
- Do not redesign cache eviction, TTL, or atomic value helpers.
- Do not move the in-memory cache implementation itself out of stable `store/cache`.
- Do not introduce compatibility aliases for removed metrics helpers.

Files:
- `store/cache/cache.go`
- `store/cache/cache_test.go`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `x/cache`
- `x/observability`

Tests:
- `go test -timeout 20s ./store/cache ./x/cache/... ./x/observability/...`
- `go test -race -timeout 60s ./store/cache ./x/cache/... ./x/observability/...`
- `go vet ./store/cache ./x/cache/... ./x/observability/...`

Docs Sync:
- Keep store docs aligned on the rule that stable `store/cache` owns cache primitives only, while metrics and inspection/export helpers live in owning extensions.

Done Definition:
- Stable `store/cache` no longer exports metrics tracker or metrics snapshot ownership.
- Cache metrics and introspection are owned outside the stable store root.
- Store docs and manifest describe the same narrowed cache surface implemented in code.
