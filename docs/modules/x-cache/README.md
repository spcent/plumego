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

## Distributed constructor notes

- Prefer `distributed.NewWithConfig` when callers need invalid node or config
  errors.
- `distributed.New` is retained as a compatibility helper and returns `nil`
  when construction fails.
- `Close` is safe to call more than once.

## Distributed behavior

- Node `Weight()` values scale virtual-node placement in the hash ring.
- Virtual-node hash collisions are resolved without overwriting existing ring
  entries and are exposed through `DistributedMetrics.HashCollisions`.
- Pathological virtual-node placement fails with
  `distributed.ErrHashRingSaturated` after a bounded collision probe window and
  rolls back the failed node add.
- `Config.HealthProbe` customizes node health checks. The default probe uses
  the wrapped cache `Exists` operation on an internal health-check key.
- Replica write failures are exposed through
  `DistributedMetrics.ReplicationFailures`, and the latest observed replica
  write duration is exposed through `DistributedMetrics.ReplicationLag`.
- `ReplicationNone` selects only the primary hash-ring node for `Set`,
  `Delete`, `Incr`, `Decr`, and `Append`; it does not require the configured
  secondary replica count to be satisfiable.
- `ReplicationSync` writes selected replicas synchronously and returns an error
  when a selected replica is unhealthy or a replica write fails. It is not a
  strong-consistency or transaction contract: replicas that accepted a mutation
  before another replica failed are not rolled back.
- `ReplicationAsync` writes the primary synchronously and schedules healthy
  secondary replicas in background goroutines.
- `Config.AsyncReplicationTimeout` bounds each async secondary write. The
  default is 2 seconds, including defensive fallback for invalid internal
  timeout state.
- `Config.AsyncReplicationMaxConcurrency` bounds concurrent async secondary
  writes. When the limit is exhausted, the secondary write is dropped and
  `DistributedMetrics.ReplicationFailures` is incremented.
- Operations that require replicas fail with `distributed.ErrInsufficientReplicas`
  when the ring cannot satisfy the configured replica count.
- `Incr`, `Decr`, and `Append` follow the configured replication mode. In
  synchronous mode, the primary mutation happens before secondary mutations; if
  a secondary mutation fails, the returned value/error reports the partial
  outcome and the primary mutation remains visible.
- `Exists` uses the same failover strategy as `Get` when the primary returns an
  infrastructure error or is unhealthy. A primary cache miss remains a miss.
- `FailoverNextNode` reads from the selected replica set. With
  `ReplicationNone`, that selected set contains only the primary, so there is
  no secondary node for next-node failover.
- `FailoverAllNodes` may read from any healthy node in the ring.
- `FailoverRetry` retries the failed primary node when it is still healthy.
  `Config.FailoverRetryAttempts` and `Config.FailoverRetryBackoff` tune this
  path; zero values use the conservative defaults of 3 attempts and 10ms
  backoff.
- Nodes must have non-empty IDs and non-nil `store/cache.Cache` instances.
- `Set`, `Delete`, and `Clear` may return an error after partial side effects
  are already visible on replicas that accepted the mutation.
- `Clear` fails closed when no node can be cleared or any selected node fails.
  It is a best-effort destructive operation and does not roll back nodes that
  were already cleared.

Asynchronous replication is best-effort. It does not report secondary write
errors to the caller and does not currently provide callback, retry, or repair
hooks; inspect `DistributedMetrics.ReplicationFailures` for observable timeout,
drop, and secondary write failure counts.

## Leaderboard behavior

- `leaderboard.MemoryLeaderboardCache` is in-process only.
- `LeaderboardConfig` defaults are normalized on an internal constructor copy;
  caller-owned config values are not mutated.
- Sorted-set operations validate context cancellation and stable `store/cache`
  key rules directly, without probing the underlying memory cache for key
  existence.
- Nil or empty members fail with `leaderboard.ErrInvalidMember`.
- Scores must be finite values; NaN and infinities fail with
  `leaderboard.ErrInvalidScore`.
- Explicitly invalid score ranges and non-negative rank ranges fail with
  `leaderboard.ErrInvalidRange`.
- `MaxLeaderboards` is enforced from tracked leaderboard count in the
  sorted-set creation path without full-map scans or concurrent over-admission.
- Failed first writes do not leave empty leaderboards behind.
- Missing leaderboards return zero for cardinality and range-removal count
  operations, and `ErrLeaderboardNotFound` for member/range read and direct
  member removal operations.
- `Close` is nil-safe and idempotent.
- After `Close`, leaderboard-specific operations and leaderboard `Clear` return
  `leaderboard.ErrClosed`.
- `LeaderboardMetrics.ZIncrements` counts successful `ZIncrBy` mutations.
- `LeaderboardMetrics.ZRems` counts actual removed members, not requested
  member names.
- Leaderboards use `DefaultTTL` when created by sorted-set writes.

## Redis adapter behavior

- `redis.Adapter` adapts caller-provided clients; it does not import a concrete
  Redis driver.
- Prefer `redis.NewValidatedAdapterWithOptions` for new call sites that need
  construction-time option validation; this is the canonical constructor for new
  adapter wiring. `redis.NewAdapterWithOptions` and `redis.NewAdapter` remain
  compatibility constructors.
- Options passed to `redis.NewAdapterWithOptions` are copied into
  constructor-owned behavior; exported fields remain for compatibility with
  older `redis.NewAdapter` call sites.
- `NewValidatedAdapterWithOptions` rejects nil clients, negative
  `MaxKeyLength`, and invalid explicit `ClearPrefix` values during
  construction.
- Redis key validation wraps stable `store/cache` key errors.
- Adapter `Get`, `Set`, and `Append` copy byte slices at the adapter boundary
  so caller mutation and client-owned buffers do not leak through the adapter
  contract.
- `Adapter.Capabilities` reports optional atomic, append, prefix-clear, and
  FlushDB behavior supported by the wrapped client and selected options.
- The minimal `redis.Client` interface supports get, set, delete, and exists.
- `Incr` and `Decr` require the wrapped client to implement
  `redis.Incrementer`; otherwise they return `redis.ErrAtomicUnsupported`.
- `Append` requires the wrapped client to implement `redis.Appender`; otherwise
  it returns `redis.ErrAtomicUnsupported`.
- `Clear` fails closed by default. When `ClearPrefix` is configured it uses
  `redis.PrefixFlusher` and does not fall back to DB-wide flushing. Without a
  prefix it calls `FlushDB` only when the client implements `redis.Flusher` and
  `Adapter.AllowFlushDB` is explicitly enabled.

## Stable-readiness blockers

- No two-release exported API stability evidence has been recorded for
  `x/cache`; promotion should select one child surface rather than the whole
  module root.
- Distributed cache async replication remains best-effort and surfaces
  secondary failures through metrics only. Async writes are bounded by timeout
  and concurrency limit, but no caller callback, retry, queue, or repair
  contract has been selected.
- Leaderboard exported API snapshots and Redis sorted-set compatibility scope
  have not been recorded. Current behavior is explicitly Plumego-local
  ranked-data behavior, not a Redis compatibility promise.
- Redis adapter behavior depends on caller-provided client implementations; no
  concrete Redis driver contract or integration matrix is part of this module,
  even though adapter option validation, byte-slice ownership, and capability
  reporting are now explicit.
- `Clear` can be namespaced through `PrefixFlusher`, but DB-wide `FlushDB`
  remains available when explicitly enabled and still needs production guidance
  before stable promotion.
- Owner sign-off and API snapshots are still missing.

Fifth-pass stabilization evidence is recorded in
`docs/extension-evidence/x-cache.md`. `x/cache/module.yaml` remains
`experimental` until the extension stability policy is satisfied.

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

## Boundary rules

- `x/cache` extends stable `store/cache` with topology-heavy or provider-specific backends; do not duplicate these in stable `store/cache`
- keep consistent-hashing and replication logic inside `x/cache/distributed`; do not push topology decisions into stable roots
- keep provider-specific client logic (Redis, future backends) isolated to their sub-packages; do not let provider details leak through the `store/cache.Cache` interface
- tenant-aware cache scoping belongs in `x/tenant/store/cache`; do not add tenant logic to `x/cache`
- do not add feature-specific ranked-data or leaderboard logic to stable `store/cache`; keep it in `x/cache/leaderboard`
