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
- keep DB analytics, summaries, and slow-query inspection out of `store/db`; route them to `x/observability/dbinsights`
- keep HTTP response caching and request-derived cache keys out of `store/cache`

## File Boundary

- `store/file` is the stable contract layer for file storage interfaces, shared file types, errors, and path/file helpers.
- `x/data/file` is the tenant-aware implementation layer for local/S3 storage backends, provider-specific config, metadata persistence, and thumbnail/image-processing helpers.
- `x/fileapi` is the HTTP transport layer for upload, download, info, delete, list, and temporary URL endpoints.
- Do not move tenant-aware path policy, metadata queries, backend-specific behavior, or image-processing pipelines into stable `store/file`.
- Do not move HTTP handlers or multipart parsing into stable `store`.

## Extension-layer cache implementations

Topology-heavy and provider-specific cache implementations have been migrated out of the stable root and now live in `x/cache`:

- `x/cache/distributed` — consistent-hashing distributed cache with replication and failover
- `x/cache/leaderboard` — ranked-data cache built on top of stable `store/cache` primitives
- `x/cache/redis` — Redis client adapter implementing `store/cache.Cache`

Current rule:

- do not add new topology-heavy or provider-heavy siblings under stable `store/cache`
- do not add HTTP response caching middleware or request-derived cache helpers under stable `store/cache`
- do not add tenant-aware adapters or tenant-specific storage policy under stable `store`
- route new topology-heavy cache capabilities to `x/cache`
- route HTTP response caching to `x/gateway/cache`
- route tenant-aware cache adapters to `x/tenant/store/cache`
