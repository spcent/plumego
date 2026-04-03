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

### x/data/rw — Read-Write Cluster

**Purpose:** Manages a primary-replica database cluster. Routes read queries to healthy replicas and write queries (and all transaction queries) to the primary.

**When to use:**
- Your service has one primary and one or more read replicas.
- You want automatic replica health checking and failover.
- You need per-query routing hints (`ForcePrimary`, `PreferReplica`).

**Key types:**

| Type / Function | Description |
|---|---|
| `Cluster` | Manages a primary + replica set; implements `ExecContext`, `QueryContext`, `QueryRowContext`, `BeginTx` |
| `New(Config)` | Constructor; returns `(*Cluster, error)` |
| `Config` | Primary `*sql.DB`, `Replicas []*sql.DB`, load balancer, routing policy, health-check config |
| `LoadBalancer` | Interface; `NewRoundRobinBalancer()` and `NewWeightedBalancer(weights)` are built-in |
| `RoutingPolicy` | Interface; `NewSQLTypePolicy()`, `NewTransactionAwarePolicy()`, `NewAlwaysPrimaryPolicy()` |
| `ForcePrimary(ctx)` | Returns a context that forces the next query to primary |
| `PreferReplica(ctx)` | Returns a context that overrides write-detection and prefers a replica |

**Routing rules:**
- `ExecContext` → always primary.
- `BeginTx` → always primary; marks the context so subsequent queries in that transaction also use primary.
- `QueryContext` / `QueryRowContext` → primary if `SQLTypePolicy` detects a write keyword or `SELECT … FOR UPDATE`; replica otherwise.
- If all replicas are unhealthy and `FallbackToPrimary` is `true` (the default), reads fall back to primary.

**Quick start:**

```go
cluster, err := rw.New(rw.Config{
    Primary:           primaryDB,
    Replicas:          []*sql.DB{replica1, replica2},
    FallbackToPrimary: true,
    HealthCheck:       rw.DefaultHealthCheckConfig(),
})
```

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
| `CrossShardAll` | Fan-out to all shards concurrently and return the first successful result |

**Sharding strategies** (all implement `Strategy`):

| Strategy | Constructor | Description |
|---|---|---|
| Hash | `NewHashStrategy()` | MD5/FNV hash of the key, modulo shard count |
| Mod | `NewModStrategy()` | Integer key modulo shard count |
| Range | `NewRangeStrategy(defs)` | Key falls within a defined numeric range |
| List | `NewListStrategy(mapping)` | Key matches a discrete value list |

**Transactions:** Use `BeginTxOnShard(ctx, shardIndex, opts)` when the target shard is known. `BeginTx` without a configured `DefaultShardIndex` returns an error.

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
