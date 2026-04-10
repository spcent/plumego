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
- keep DB analytics, summaries, instrumentation wrappers, pool-stat polling, and slow-query inspection out of `store/db`; route them to `x/observability/dbinsights`
- keep DB health payloads, open-retry loops, and generic timeout policy helpers out of `store/db`; callers own operation deadlines through `context.Context`
- keep HTTP response caching, request-derived cache keys, and cache metrics/introspection ownership out of `store/cache`
- keep signed URLs, metadata-manager ownership, uploader/image metadata, and file path/id helper policy out of `store/file`; route them to `x/data/file` and `x/fileapi`
- keep durable KV-engine concerns such as WAL, snapshots, serializer selection, compression, and shard tuning out of `store/kv`; route them to `x/data/kvengine`
- keep durable idempotency providers, SQL dialect policy, and table schema policy out of `store/idempotency`; route them to `x/data/idempotency`

## File Boundary

- `store/file` is the stable contract layer for file storage interfaces, shared file types, errors, and path/file helpers.
- `x/data/file` is the tenant-aware implementation layer for local/S3 storage backends, provider-specific config, metadata persistence, and thumbnail/image-processing helpers.
- `x/fileapi` is the HTTP transport layer for upload, download, info, delete, list, and temporary URL endpoints.
- Do not move tenant-aware path policy, metadata queries, backend-specific behavior, or image-processing pipelines into stable `store/file`.
- Do not move HTTP handlers or multipart parsing into stable `store`.

## KV Boundary

- `store/kv` is the stable small embedded KV primitive for file-backed key/value persistence, TTL-aware CRUD, key scans, and basic stats.
- `x/data/kvengine` owns durable-engine behavior such as WAL, snapshots, serializer formats, compression, and shard/flush tuning.
- Do not add engine-format plumbing, snapshot APIs, or durability-tuning knobs back into stable `store/kv`.

## Idempotency Boundary

- `store/idempotency` is the stable primitive contract for idempotency records, statuses, errors, and the minimal `Store` interface.
- `x/data/idempotency` owns durable KV/SQL provider implementations, SQL dialect policy, table naming, and duplicate-key handling.
- Do not add provider-specific adapters, table schema policy, or feature-specific dedupe rules back into stable `store/idempotency`.

## DB Boundary

- `store/db` helpers execute with the exact `context.Context` supplied by the caller.
- Query and transaction helpers must not infer deadlines from optional config interfaces.
- Use `context.WithTimeout` or `context.WithDeadline` at the application or owning extension boundary when an operation deadline is required.

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
