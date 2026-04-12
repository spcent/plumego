# Card 0960: Store Cache Doc API Resync

Priority: P2
State: done
Primary Module: store

## Goal

Resync `store/cache` package documentation with the real stable cache API and
behavior so the package comment teaches the surface that actually exists.

## Problem

- `store/cache/cache.go` package comments still describe an API that no longer
  exists:
  - `cache.New(...)`
  - `Config{MaxSize, TTL}`
- The live stable constructors are:
  - `NewMemoryCache()`
  - `NewMemoryCacheWithConfig(Config)`
- The package comment also advertises "LRU eviction", but the stable
  implementation is documented elsewhere as an in-memory primitive with TTL and
  memory-bounded eviction, not a canonical LRU cache surface.

This makes the stable cache package look broader and different from its real
API, which is exactly the kind of learning-surface drift we have been pruning
from other stable roots.

## Scope

- Rewrite `store/cache` package-level docs/examples to use the live constructor
  and config shape.
- Remove or tighten stale capability claims such as `LRU` when they do not
  match the intended stable surface description.
- Keep the package comment aligned with `docs/modules/store/README.md` and the
  current `MemoryCache` behavior.

## Non-Goals

- Do not add a `cache.New(...)` compatibility wrapper.
- Do not change cache runtime behavior in this card unless the doc mismatch
  reveals a real stable-surface bug.
- Do not widen the card into `x/cache` topology work or HTTP response caching.

## Files

- `store/cache/cache.go`
- `docs/modules/store/README.md`
- any nearby stable docs/examples that still teach the stale cache constructor
  or config names

## Tests

- `go test -timeout 20s ./store/cache`
- `go test -race -timeout 60s ./store/cache`
- `go vet ./store/cache`
- `rg -n 'cache\\.New\\(|MaxSize|\\bTTL\\b|LRU eviction' store/cache docs/modules/store/README.md`

## Docs Sync

- Keep touched stable store docs aligned with the final cache wording.

## Done Definition

- `store/cache` package docs no longer advertise removed constructor/config
  names.
- Capability wording for the stable cache surface matches the implemented
  primitive.
- Grep for the stale example symbols in touched docs is empty.

## Outcome

- Rewrote the `store/cache` package comment to use the live
  `NewMemoryCache` / `NewMemoryCacheWithConfig` constructor path.
- Removed the stale `cache.New(...)` and `Config{MaxSize, TTL}` example shape.
- Tightened the capability wording from "LRU eviction" to the actual stable
  TTL-aware, memory-bounded primitive description.

## Validation Run

```bash
go test -timeout 20s ./store/cache
go test -race -timeout 60s ./store/cache
go vet ./store/cache
rg -n 'cache\.New\(|LRU eviction' store/cache docs/modules/store/README.md
```
