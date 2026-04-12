# x/data

## Purpose

`x/data` contains topology-heavy and fast-evolving data capabilities such as sharding and advanced rw patterns.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

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
| `Options` | Engine config including flush cadence, cleanup cadence, compression, serializer format, shard count, and read-only mode |
| `SerializationFormat` | Binary/JSON engine format selection |
| `NewKVStore(opts)` | Constructor for the durable engine |

**Boundary rule:**
- Keep the stable `store/kv` package limited to the small embedded primitive.
- Route WAL, snapshots, serializer plumbing, compression, and shard tuning to `x/data/kvengine`.

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
| `WithPreferReplica(ctx)` | Returns a context that overrides write-detection and prefers a replica |

**Routing rules:**
- `ExecContext` → always primary.
- `BeginTx` → always primary; marks the context so subsequent queries in that transaction also use primary.
- `QueryContext` / `QueryRowContext` → primary if `SQLTypePolicy` detects a write keyword or `SELECT … FOR UPDATE`; replica otherwise.
- `WithForcePrimary(ctx)` is the explicit escape hatch for read-after-write or any other "read from primary now" requirement.
- `WithPreferReplica(ctx)` overrides the SQL-type heuristic and should only be used for known-safe reads.
- If all replicas are unhealthy and `FallbackToPrimary` is explicitly set to `true`, reads fall back to primary; otherwise the query returns a routing error.
- Background replica health checks run only when `HealthCheck.Enabled` is `true`; they use periodic `PingContext` probes and remove replicas from balancing only after the configured failure threshold.

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
| `CrossShardFirst` | Execute on shard 0 only; use for approximate or sampling queries |
| `CrossShardAll` | Fan-out to all shards concurrently and return the first successful result; it does not merge rows across shards |

**Sharding strategies** (all implement `Strategy`):

| Strategy | Constructor | Description |
|---|---|---|
| Hash | `NewHashStrategy()` | MD5/FNV hash of the key, modulo shard count |
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
- Use `BeginTxOnShard(ctx, shardIndex, opts)` when the target shard is known. `BeginTx` without a configured `DefaultShardIndex` returns an error.
- Keep `DefaultShardIndex` at `-1` by default so unresolved routing stays visible instead of silently pinning traffic to one shard.

**Quick start:**

```go
// Build one rw.Cluster per shard
shards := []*rw.Cluster{shard0, shard1, shard2, shard3}

// Register sharding rules
registry := sharding.NewShardingRuleRegistry()
registry.Register(sharding.ShardingRule{
    Table:    "orders",
    KeyCol:   "user_id",
    Strategy: sharding.NewHashStrategy(),
})

// Create router
router, err := sharding.NewRouter(shards, registry,
    sharding.WithCrossShardPolicy(sharding.CrossShardDeny),
)
```

**Dynamic config:** Live rule updates are available via `x/data/sharding/config` — see `x/data/sharding/config/README.md`.

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
