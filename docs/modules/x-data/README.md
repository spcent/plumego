# x/data

## Purpose

`x/data` contains topology-heavy and fast-evolving data capabilities such as sharding and advanced rw patterns.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen
- Stable readiness review on 2026-05-04: core correctness blockers from the
  follow-up cards are addressed, but the module remains experimental until the
  public API surface, SQL support policy, and operational limits are frozen.

Stable promotion blockers:

- Decide whether the sharding convenience wrapper (`ClusterDB`/`New`) is part
  of the long-term public surface or remains a documented convenience layer over
  `Router`.
- Keep SQL rewrite support intentionally narrow unless a parser-backed strategy
  is approved; current support is simple single-statement identifier
  replacement with fail-closed rejection for complex shapes and
  schema-qualified targets.
- Decide whether `kvengine.Options` should keep both `AutoDetectFormat` and
  `DisableAutoDetect` before compatibility is frozen.
- Define large-object S3 policy beyond standard-library single PUT spooling
  before advertising high-volume object storage guarantees.
- Run repo-wide gates before any status change from experimental to stable.

Second stable-readiness gate on 2026-05-05 passed:

- `go test -timeout 20s ./x/data/...`
- `go test -race -timeout 60s ./x/data/...`
- `go vet ./x/data/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`

Status remains `Experimental`: the second gate confirms the 0744-0750
correctness and lifecycle fixes, but public API freeze decisions and
large-object operational policy are still unresolved.

## Use this module when

- the task is beyond stable `store` primitives
- sharding or topology-aware behavior is involved

## Do not use this module for

- stable store contracts
- application bootstrap

## First files to read

- `x/data/module.yaml`
- the owning subpackage under `x/data/*`
- `specs/repo.yaml`

---

## Submodules

### x/data/file — Tenant-Aware File Storage and Metadata

**Purpose:** Owns tenant-aware local/S3 storage implementations and PostgreSQL
file metadata persistence behind the stable `store/file` contracts.

**When to use:**
- You need tenant-isolated file paths, metadata records, or temporary URL
  generation.
- You need to change database metadata persistence for files.
- The task is storage behavior rather than HTTP multipart parsing or response
  headers.

**Key types:**

| Type / Function | Description |
|---|---|
| `LocalStorage` | Tenant-aware filesystem storage implementation |
| `S3Storage` | S3-compatible storage implementation |
| `DBMetadataManager` | PostgreSQL-backed metadata manager |
| `NewDBMetadataManagerE` | Error-returning metadata manager constructor for dynamic wiring |
| `WithMetadataClock` | Testable clock option for metadata mutation timestamps |

**Boundary rule:**
- Keep HTTP upload/download behavior in `x/fileapi`.
- Keep stable storage contracts and shared errors in `store/file`.
- Keep file metadata SQL behavior here; the DB metadata manager is PostgreSQL-only
  unless a future card adds explicit dialect support.
- Deduplication is tenant-scoped: same-content files in different tenants must
  not return another tenant's metadata record.
- Tenant-facing metadata reads and mutations are tenant-scoped by tenant id;
  global id/path metadata access is not part of the tenant storage contract.
- Local and S3 storage generate object IDs from crypto-random bytes and fail the
  write if secure ID generation fails. S3 URLs preserve object-key hierarchy and
  escape unsafe path segments.
- Local writes go through a temporary file and check sync/close before rename.
- Local writes sync the containing directory after rename so directory metadata
  durability is requested where the platform supports it.
- Local static URLs validate storage paths and escape path segments. Local copy
  and thumbnail writes use the same temp-file, sync, close, rename, and
  directory-sync durability path as uploads.
- S3 `Put` hashes while spooling upload content to a temporary file, then
  streams that file to the object store with a fixed content length instead of
  buffering the whole object in memory. Set `S3Config.TempDir` to control the
  spool directory; S3 error response bodies are read with a small fixed bound
  before being included in returned errors. Presigned URLs include the actual
  request host in the SigV4 canonical request, and S3 status errors are returned
  explicitly instead of being treated as missing objects.

**See:** `x/data/file/module.yaml` for the manifest.

### x/data/idempotency — Durable Idempotency Providers

**Purpose:** Owns durable provider implementations for the stable `store/idempotency` contract.

**When to use:**
- You need a SQL-backed idempotency store with explicit dialect and table policy.
- You need a KV-backed idempotency store using the stable `store/kv` primitive.
- You are wiring extension features such as MQ task dedupe to durable idempotency persistence.

**Key types:**

| Type / Function | Description |
|---|---|
| `SQLStore` | SQL-backed implementation of `store/idempotency.Store` |
| `SQLConfig` | SQL dialect, table, and clock configuration |
| `KVStore` | KV-backed implementation of `store/idempotency.Store` |
| `KVConfig` | KV prefix and clock configuration |

**Boundary rule:**
- Keep `store/idempotency` limited to records, statuses, errors, and the minimal store contract.
- Keep SQL dialects, table policy, durable provider behavior, and duplicate-key handling in `x/data/idempotency`.
- Keep domain-specific dedupe rules in the owning application or extension, such as `x/mq`.
- `SQLConfig.DuplicateError` is the explicit duplicate-key classifier hook for
  driver-specific error codes. The built-in string matcher remains only as
  compatibility fallback behavior for existing tests and simple drivers.
- `Complete` is a conditional transition: only an existing, unexpired
  `in_progress` record can become `completed`. Missing, expired, or already
  final records return the existing not-found/expired sentinel behavior instead
  of being overwritten.
- KV-backed idempotency serializes claim/complete/delete sequences across
  wrappers that share the same in-process stable `store/kv` instance.

**See:** `x/data/idempotency/module.yaml` for the manifest.

### x/data/kvengine — Durable Embedded KV Engine

**Purpose:** Owns the topology-heavy and durability-heavy embedded KV engine surface that should not remain in stable `store/kv`.

**When to use:**
- You need WAL-backed persistence with explicit flush and cleanup intervals.
- You need snapshot/replay, serializer selection, compression, or shard-count tuning.
- The stable `store/kv` primitive is too small for the persistence behavior you need.

**Key types:**

| Type / Function | Description |
|---|---|
| `KVStore` | Durable embedded KV engine with WAL + snapshot support |
| `Options` | Engine config including flush cadence, cleanup cadence, compression, serializer format, auto-detect policy, shard count, and read-only mode |
| `SerializationFormat` | Binary/JSON engine format selection |
| `NewKVStore(opts)` | Constructor for the durable engine |
| `Default()` | Convenience constructor that explicitly uses the local `data` directory |

**Boundary rule:**
- Keep the stable `store/kv` package limited to the small embedded primitive.
- Route WAL, snapshots, serializer plumbing, compression, and shard tuning to `x/data/kvengine`.
- `NewKVStore` requires an explicit `Options.DataDir`; it does not silently
  create a relative `data` directory. Use `Default()` only when that local
  convenience path is intentional.
- WAL replay fails closed on decode or CRC corruption. `WALSyncMode` defaults
  to `immediate`, so acknowledged Set/Delete calls flush and fsync the WAL
  before memory state changes; set `WALSyncInterval` only when async durability
  is an explicit performance tradeoff. `AutoDetectFormat` is enabled by
  default; set `DisableAutoDetect` when the configured serializer must be
  enforced.
- `SetMetricsCollector` observes `Set`, `Get`, and `Delete` operations,
  including misses and returned errors. Collector get/set/use is safe under
  concurrent access.
- `Close` is idempotent and repeated calls return the first close result.

**See:** `x/data/kvengine/module.yaml` for the manifest.

### x/data/rw — Read-Write Cluster

**Purpose:** Manages a primary-replica database cluster. Routes read queries to healthy replicas and write queries (and all transaction queries) to the primary.

**When to use:**
- Your service has one primary and one or more read replicas.
- You want automatic replica health checking and failover.
- You need per-query routing hints (`WithForcePrimary`, `WithPreferReplica`).

**Key types:**

| Type / Function | Description |
|---|---|
| `Cluster` | Manages a primary + replica set; implements `ExecContext`, `QueryContext`, `QueryRowContext`, `BeginTx` |
| `New(Config)` | Constructor; returns `(*Cluster, error)` |
| `Config` | Primary `*sql.DB`, `Replicas []*sql.DB`, load balancer, routing policy, health-check config |
| `LoadBalancer` | Interface; `NewRoundRobinBalancer()` and `NewWeightedBalancer(weights)` are built-in |
| `RoutingPolicy` | Interface; `NewSQLTypePolicy()`, `NewTransactionAwarePolicy()`, `NewAlwaysPrimaryPolicy()` |
| `WithForcePrimary(ctx)` | Returns a context that forces the next query to primary |
| `WithPreferReplica(ctx)` | Returns a context that prefers a replica for statements classified as safe reads |

**Routing rules:**
- `ExecContext` → always primary.
- `BeginTx` → always primary; marks the context so subsequent queries in that transaction also use primary.
- `QueryContext` / `QueryRowContext` → primary if `SQLTypePolicy` detects a write keyword, lock-taking read, or unknown statement; replica for known-safe reads.
- `WithForcePrimary(ctx)` is the explicit escape hatch for read-after-write or any other "read from primary now" requirement.
- `WithPreferReplica(ctx)` only affects known-safe reads; it cannot force write-like or unknown statements to replicas.
- If all replicas are unhealthy and `FallbackToPrimary` is explicitly set to `true`, reads fall back to primary; otherwise the query returns a routing error.
- Background replica health checks run only when `HealthCheck.Enabled` is `true`; they use periodic `PingContext` probes and remove replicas from balancing only after the configured failure threshold.
- Set `Config.HealthCheckContext` when health checks should inherit a
  caller-owned shutdown context; otherwise they use `context.Background()` and
  stop through `Cluster.Close`.
- `ReplicaWeights`, when provided, must match the replica count and each weight
  must be positive. Use `NewWeightedBalancerE` for constructor-time validation
  when creating a weighted balancer directly.
- `Cluster.Close` and health checker stop are idempotent; closing the cluster
  owns and closes the configured `*sql.DB` handles once.

**Quick start:**

```go
cluster, err := rw.New(rw.Config{
    Primary:           primaryDB,
    Replicas:          []*sql.DB{replica1, replica2},
    FallbackToPrimary: true,
    HealthCheck:       rw.DefaultHealthCheckConfig(),
})
```

Use `FallbackToPrimary: true` only when serving stale-sensitive reads from the primary during replica outages is acceptable for your service. If you need read-after-write visibility on a per-request basis, wrap the read context with `rw.WithForcePrimary(ctx)` instead of changing the whole cluster policy.

**See:** `x/data/rw/module.yaml` for full manifest.

---

### x/data/sharding — Sharding Router

**Purpose:** Routes SQL queries to the correct shard in a horizontally-partitioned database cluster. Each shard is an `rw.Cluster`, so read-write splitting and replica health management are inherited automatically.

**When to use:**
- Data is horizontally partitioned across multiple database instances.
- You need deterministic, key-based routing before queries reach the database.
- You want pluggable sharding strategies (hash, mod, range, list, or custom).

**Key types:**

| Type / Function | Description |
|---|---|
| `Router` | Routes `ExecContext`, `QueryContext`, `QueryRowContext`, `BeginTxOnShard` across shards |
| `NewRouter(shards, registry, opts...)` | Constructor; returns `(*Router, error)` |
| `ClusterDB` | Convenience wrapper that builds rw clusters from sharding config and delegates routing to `Router` |
| `New(ClusterConfig)` | Convenience constructor for `ClusterDB`; takes ownership of configured shard DB handles |
| `RouterConfig` | `CrossShardPolicy`, `DefaultShardIndex`, `EnableMetrics` |
| `WithCrossShardPolicy(p)` | Option: `CrossShardDeny` (default), `CrossShardFirst`, `CrossShardAll` |
| `Strategy` | Interface for sharding strategies |
| `ShardingRuleRegistry` | Holds per-table sharding rules and strategies |
| `ShardKeyResolver` | Extracts shard key from SQL query arguments |
| `SQLRewriter` | Rewrites logical table names to physical shard-suffixed names |

**Cross-shard policies:**

| Policy | Behaviour |
|---|---|
| `CrossShardDeny` | Reject queries that cannot be resolved to a single shard (safe default) |
| `CrossShardFirst` | Execute on the first resolved shard, or shard 0 for unresolved queries; use for approximate or sampling queries |
| `CrossShardAll` | Fan-out to all shards concurrently, return the first successful result, and cancel remaining shard queries; it does not merge rows across shards |

**Sharding strategies** (all implement `Strategy`):

| Strategy | Constructor | Description |
|---|---|---|
| Hash | `NewHashStrategy()` | FNV hash of the key, modulo shard count |
| Mod | `NewModStrategy()` | Integer key modulo shard count |
| Range | `NewRangeStrategy(defs)` | Key falls within a defined numeric range |
| List | `NewListStrategy(mapping)` | Key matches a discrete value list |

**Strategy selection guidance:**

- `mod` for stable integer IDs and evenly distributed numeric keys.
- `hash` for arbitrary strings or other non-numeric keys.
- `range` for ordered domains where operators need predictable shard spans.
- `list` for a small, explicit set of values such as region codes.

**Routing limits and transactions:**

- Keep `CrossShardDeny` unless a specific read path can tolerate approximate or first-success semantics.
- `CrossShardAll` returns the first successful `*sql.Rows`; callers must not
  expect merged result sets, and late successful rows are closed by the router.
- `IN` and bounded range predicates can resolve to multiple shards; they still follow the configured cross-shard policy.
- `QueryRowContext` follows the same fail-closed routing rules as `QueryContext`;
  route resolution errors and invalid shard indexes are returned from `Scan`
  instead of falling back to `DefaultShardIndex`.
- SQL rewriting supports simple single-statement table replacement. Nested
  `SELECT`, CTE, `UNION`, and multiple-statement queries fail closed instead of
  using broad string replacement.
- SQL rewriting only changes table identifiers in SQL code regions. String
  literals and comments are preserved, and schema-qualified target tables such
  as `public.users` fail closed until parser-backed schema support is added.
- Use `BeginTxOnShard(ctx, shardIndex, opts)` when the target shard is known. `BeginTx` without a configured `DefaultShardIndex` returns an error.
- Keep `DefaultShardIndex` at `-1` by default so unresolved routing stays visible instead of silently pinning traffic to one shard.
- `ShardingRuleRegistry` protects its map under concurrent access. Returned
  `*ShardingRule` values should still be treated as immutable once a router is
  built.

**Observability boundary:**

- `x/data/sharding` may expose local topology metrics and lightweight trace
  helpers for shard decisions, SQL classification, and rewrite/cache counters.
- Sharding logging and trace attributes must not record raw SQL text, query
  arguments, or shard-key values by default. Record safe metadata such as
  operation, shard index, table name, argument count, and redaction markers
  instead.
- Generic tracing infrastructure, exporters, collectors, and sampling policy
  belong in `x/observability`; do not import `x/observability` into `x/data`
  just to wire a backend.
- Prometheus text emitted by the sharding metrics helper is local topology
  output. Broader export orchestration belongs in `x/observability`. The helper
  exposes latency min/avg/max and histogram buckets, but not percentiles.
- Router `SingleShardQueries` counts queries planned for exactly one shard.
  Cross-shard fan-out updates `CrossShardQueries` and per-shard execution
  counts without inflating the single-shard total.

**Quick start:**

```go
// Build one rw.Cluster per shard
shards := []*rw.Cluster{shard0, shard1, shard2, shard3}

// Register sharding rules
registry := sharding.NewShardingRuleRegistry()
rule, err := sharding.NewShardingRule("orders", "user_id", sharding.NewHashStrategy(), len(shards))
if err != nil {
    return err
}
if err := registry.Register(rule); err != nil {
    return err
}

// Create router
router, err := sharding.NewRouter(shards, registry,
    sharding.WithCrossShardPolicy(sharding.CrossShardDeny),
)
```

**Dynamic config:** Live rule updates are available via `x/data/sharding/config` — see `x/data/sharding/config/README.md`. `ConfigWatcher.Start` is a single-use lifecycle method; repeated starts return an error, while `Stop` is idempotent.

**See:** `x/data/sharding/module.yaml` for full manifest.

---

## Composition pattern

`x/data/rw` and `x/data/sharding` are designed to compose:

```
Application
    └── sharding.Router          (horizontal partitioning)
            └── rw.Cluster × N   (primary/replica per shard)
                    └── *sql.DB  (physical connections)
```

Build each shard as an `rw.Cluster`, then pass the slice to `sharding.NewRouter`.
