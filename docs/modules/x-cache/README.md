# x/cache

## Purpose

`x/cache` provides extension-layer cache adapters and topology-heavy cache implementations. It builds on the stable `store/cache` abstractions.

## Status

- `experimental` in the Plumego extension layer
- Migrated from `store/cache/distributed` and `store/cache/redis` (Phase 4 stable-root debt reduction)

## Use this module when

- implementing distributed caching with consistent hashing
- building ranked-data or leaderboard cache features on top of `store/cache`
- adapting a Redis client to the `store/cache.Cache` interface
- building topology-heavy or provider-specific cache backends

## Do not use this module for

- core in-memory caching (use `store/cache.MemoryCache`)
- tenant-aware cache scoping (use `x/tenant/store/cache`)
- adding tenant-specific logic to cache keys

## Sub-packages

- `x/cache/distributed` — consistent-hashing distributed cache with replication modes and failover strategies
- `x/cache/leaderboard` — in-memory ranked-data cache on top of stable `store/cache` primitives
- `x/cache/redis` — minimal Redis client adapter implementing `store/cache.Cache`

## First files to read

- `x/cache/module.yaml`
- `x/cache/distributed/distributed.go`
- `x/cache/leaderboard/leaderboard.go`
- `x/cache/redis/redis.go`

## Canonical change shape

- implement `store/cache.Cache` interface
- keep topology decisions in this layer, not in stable store
- keep feature-specific cache behavior in this layer, not in stable store
- keep provider-specific logic isolated to sub-packages
