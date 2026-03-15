# store

## Purpose

`store` holds persistence primitives and base abstractions.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- defining stable storage contracts
- implementing basic persistence helpers
- working below topology-heavy data features

## Do not use this module for

- tenant-aware storage policy
- tenant-aware adapters
- sharding or heavy topology defaults in stable roots
- app bootstrap

## First files to read

- `store/module.yaml`
- the target package under `store/*`
- `specs/repo.yaml`

## Canonical change shape

- keep interfaces narrow
- keep concurrent behavior testable
- move topology-heavy features to owning extensions

## Extension-layer cache implementations

Topology-heavy and provider-specific cache implementations have been migrated out of the stable root and now live in `x/cache`:

- `x/cache/distributed` — consistent-hashing distributed cache with replication and failover
- `x/cache/redis` — Redis client adapter implementing `store/cache.Cache`

Current rule:

- do not add new topology-heavy or provider-heavy siblings under stable `store/cache`
- do not add tenant-aware adapters or tenant-specific storage policy under stable `store`
- route new topology-heavy cache capabilities to `x/cache`
- route tenant-aware cache adapters to `x/tenant/store/cache`
