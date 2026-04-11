# Card 0927: Store And Metrics Primitive Surface Convergence

Priority: P2
State: done
Primary Module: store

## Goal

Prune remaining stable primitive surfaces that carry stale or inconsistent fields: `store/cache`, `store/kv`, and the stable `metrics` collector fan-out helpers.

## Problem

Several stable primitives still show leftover duplication or stale shape after earlier boundary pruning:

`store/cache`:
- `NewMemoryCache()` and `NewMemoryCacheWithConfig(config)` are two constructor paths.
- Invalid config panics in `NewMemoryCacheWithConfig`, while other stable store constructors return errors.
- The `Cache` interface includes atomic/string operations (`Incr`, `Decr`, `Append`) that may exceed the small generic cache primitive.

`store/kv`:
- `Stats` still exposes `WALSize` even though WAL ownership moved to `x/data/kvengine`.
- `Stats.Evictions` exists but the stable primitive does not update it.
- `Close` returns `ErrStoreClosed` on repeated close, which may be inconsistent with idempotent close behavior elsewhere.

`metrics`:
- `NewMultiHTTPObserver` filters nil observers and returns nil/one/multi depending on input.
- `NewMultiCollector` keeps nil collectors and always returns a concrete collector, creating different fan-out nil handling.
- Collector stats shapes should stay consistent across base/noop/multi collectors.

These are related because they are all stable primitive APIs with small, repeatable convergence work and low cross-module ownership ambiguity.

## Scope

- Inspect all callers before removing or changing exported symbols.
- Make cache constructor behavior explicit and non-panicking if feasible.
- Decide whether `Incr`, `Decr`, and `Append` remain in stable `store/cache` or move to an extension-specific cache package.
- Remove stale `store/kv.Stats` fields that no longer represent stable behavior.
- Align `Close` semantics only if tests and existing callers support the change.
- Align metrics fan-out nil handling and stats shape between `MultiCollector`, `NoopCollector`, and `NewMultiHTTPObserver`.
- Update docs/manifests for changed stable primitive surfaces.

## Non-Goals

- Do not remove the small stable `store/kv` concrete API.
- Do not move durable KV-engine behavior back into stable `store`.
- Do not add cache provider implementations to stable `store/cache`.
- Do not add feature-specific metrics methods to stable `metrics`.

## Expected Files

- `store/cache/cache.go`
- `store/cache/cache_test.go`
- `store/kv/kv.go`
- `store/kv/kv_test.go`
- `metrics/*.go`
- `metrics/*_test.go`
- `docs/modules/store/README.md`
- `docs/modules/metrics/README.md`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./store/cache ./store/kv ./metrics
go test -race -timeout 60s ./store/cache ./store/kv ./metrics
go vet ./store/cache ./store/kv ./metrics
```

Then run the required repo-wide gates before committing.

## Done Definition

- Stable store primitive constructors and error behavior are consistent.
- `store/kv.Stats` no longer exposes stale durable-engine fields.
- Metrics fan-out helpers handle nil collectors consistently.
- Removed exported symbols have zero residual references.
- Focused gates and repo-wide gates pass.

## Outcome

- `store/cache` constructors now return validation errors instead of panicking.
- `store/kv.Stats` drops stale durable-engine fields and `Close` is idempotent.
- `metrics.NewMultiCollector` now filters nil collectors and returns nil when no collectors are supplied.
