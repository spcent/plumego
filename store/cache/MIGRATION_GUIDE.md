# Cache Package Migration Guide

This guide helps you migrate to the new cache package features added in Phase 2 and Phase 3, including **Leaderboard Support** and **Distributed Caching**.

## Table of Contents

1. [Overview](#overview)
2. [Backward Compatibility](#backward-compatibility)
3. [Upgrading to Leaderboard Cache](#upgrading-to-leaderboard-cache)
4. [Upgrading to Distributed Cache](#upgrading-to-distributed-cache)
5. [Combining Features](#combining-features)
6. [Performance Considerations](#performance-considerations)
7. [Common Migration Patterns](#common-migration-patterns)
8. [Breaking Changes](#breaking-changes)

---

## Overview

### What's New

**Phase 2 - Leaderboard Support:**
- Skip list-based sorted sets
- Redis-compatible leaderboard API (ZAdd, ZRange, ZRank, etc.)
- O(log N) performance for most operations
- Built-in metrics tracking

**Phase 3 - Distributed Caching:**
- Consistent hashing with virtual nodes
- Multi-node support with automatic sharding
- Configurable replication (None, Async, Sync)
- Health checking and automatic failover
- Dynamic node addition/removal

### Release Timeline

- **v1.0** - Basic in-memory cache (LRU eviction, TTL support)
- **v1.1** - Leaderboard support (this release)
- **v1.2** - Distributed caching (this release)

---

## Backward Compatibility

### Existing Code Continues to Work

All existing cache functionality remains **100% compatible**. No breaking changes to the core `Cache` interface.

**Before (v1.0):**
```go
config := cache.DefaultConfig()
c := cache.NewMemoryCacheWithConfig(config)

ctx := context.Background()
c.Set(ctx, "key", []byte("value"), time.Hour)
value, err := c.Get(ctx, "key")
```

**After (v1.1+):**
```go
// Same code works identically
config := cache.DefaultConfig()
c := cache.NewMemoryCacheWithConfig(config)

ctx := context.Background()
c.Set(ctx, "key", []byte("value"), time.Hour)
value, err := c.Get(ctx, "key")
```

### New Interfaces Are Opt-In

The new features are exposed through **separate interfaces**:

- `LeaderboardCache` - For sorted set operations
- `DistributedCache` - For multi-node caching

You only use them if you need the additional functionality.

---

## Upgrading to Leaderboard Cache

### Use Case

You need **sorted sets** for:
- Gaming leaderboards
- Trending content rankings
- Time-series data (sorted by timestamp)
- Priority queues
- Range-based queries

### Basic Migration

**Before - Manual Sorting:**
```go
// Old way: Store scores separately and sort manually
type Player struct {
    Name  string
    Score float64
}

var players []Player
// ... add players ...

// Manual sorting
sort.Slice(players, func(i, j int) bool {
    return players[i].Score > players[j].Score
})

// Get top 5
top5 := players[:5]
```

**After - Using LeaderboardCache:**
```go
import "github.com/spcent/plumego/store/cache"

// Create leaderboard cache
config := cache.DefaultConfig()
lbConfig := cache.DefaultLeaderboardConfig()
lbc := cache.NewMemoryLeaderboardCache(config, lbConfig)
defer lbc.Close()

ctx := context.Background()

// Add players
lbc.ZAdd(ctx, "game:scores",
    &cache.ZMember{Member: "Alice", Score: 1250.0},
    &cache.ZMember{Member: "Bob", Score: 980.0},
    &cache.ZMember{Member: "Charlie", Score: 1500.0},
)

// Get top 5 (automatically sorted by score descending)
top5, err := lbc.ZRange(ctx, "game:scores", 0, 4, true)
for i, player := range top5 {
    fmt.Printf("#%d: %s (%.0f)\n", i+1, player.Member, player.Score)
}
```

### Configuration

**Default Configuration:**
```go
lbConfig := cache.DefaultLeaderboardConfig()
// MaxLeaderboards: 1000
// MaxMembersPerSet: 10000
// EnableMetrics: true
```

**Custom Configuration:**
```go
lbConfig := &cache.LeaderboardConfig{
    MaxLeaderboards:   100,     // Maximum number of leaderboards
    MaxMembersPerSet:  50000,   // Members per leaderboard
    EnableMetrics:     true,    // Track ZAdd, ZRange, ZRank metrics
}

// Validate before use
if err := lbConfig.Validate(); err != nil {
    log.Fatal(err)
}

lbc := cache.NewMemoryLeaderboardCache(config, lbConfig)
```

### Key Operations

**Adding/Updating Scores:**
```go
// Add or update score
lbc.ZAdd(ctx, "leaderboard", &cache.ZMember{Member: "Alice", Score: 100})

// Increment score
newScore, _ := lbc.ZIncrBy(ctx, "leaderboard", "Alice", 50) // now 150
```

**Querying Rankings:**
```go
// Get top N (0-indexed, reverse=true for descending)
top10, _ := lbc.ZRange(ctx, "leaderboard", 0, 9, true)

// Get specific rank
rank, _ := lbc.ZRank(ctx, "leaderboard", "Alice", true) // 0-indexed

// Get score
score, _ := lbc.ZScore(ctx, "leaderboard", "Alice")

// Get by score range
players, _ := lbc.ZRangeByScore(ctx, "leaderboard", 100.0, 200.0, true)
```

**Removal:**
```go
// Remove specific members
lbc.ZRem(ctx, "leaderboard", "Alice", "Bob")

// Remove by rank range
lbc.ZRemRangeByRank(ctx, "leaderboard", 100, -1) // Remove rank 100+

// Remove by score range
lbc.ZRemRangeByScore(ctx, "leaderboard", 0, 50) // Remove scores 0-50
```

### Metrics

```go
metrics := lbc.GetLeaderboardMetrics()
fmt.Printf("ZAdd operations: %d\n", metrics.ZAdds)
fmt.Printf("ZRange queries: %d\n", metrics.ZRangeQueries)
fmt.Printf("ZRank calculations: %d\n", metrics.ZRankCalculations)
```

---

## Upgrading to Distributed Cache

### Use Case

You need **horizontal scaling** for:
- High-traffic applications
- Geographic distribution
- Fault tolerance
- Load balancing
- Sharding large datasets

### Basic Migration

**Before - Single Node:**
```go
// Old way: Single cache instance
config := cache.DefaultConfig()
c := cache.NewMemoryCacheWithConfig(config)

c.Set(ctx, "user:123", userData, time.Hour)
```

**After - Distributed:**
```go
import "github.com/spcent/plumego/store/cache/distributed"

// Create multiple cache nodes
nodes := []distributed.CacheNode{
    distributed.NewNode("node-1", cache.NewMemoryCacheWithConfig(config)),
    distributed.NewNode("node-2", cache.NewMemoryCacheWithConfig(config)),
    distributed.NewNode("node-3", cache.NewMemoryCacheWithConfig(config)),
}

// Create distributed cache
distConfig := distributed.DefaultConfig()
distConfig.ReplicationFactor = 2  // 2x replication
dc := distributed.New(nodes, distConfig)
defer dc.Close()

// Same API, but distributed automatically
dc.Set(ctx, "user:123", userData, time.Hour)
```

**Key changes:**
- Keys are automatically sharded across nodes
- Reads/writes routed to correct node via consistent hashing
- Replication provides fault tolerance
- Health checking with automatic failover

### Configuration

**Default Configuration:**
```go
distConfig := distributed.DefaultConfig()
// VirtualNodes: 150
// ReplicationFactor: 1
// ReplicationMode: Async
// FailoverStrategy: NextNode
// HealthCheckInterval: 10s
// HealthCheckTimeout: 2s
```

**Custom Configuration:**
```go
distConfig := &distributed.Config{
    VirtualNodes:        300,                         // More = better distribution
    ReplicationFactor:   3,                           // 3x replication
    ReplicationMode:     distributed.ReplicationSync, // Sync writes
    FailoverStrategy:    distributed.FailoverNextNode,
    HealthCheckInterval: 5 * time.Second,
    HealthCheckTimeout:  1 * time.Second,
    EnableMetrics:       true,
}

dc := distributed.New(nodes, distConfig)
```

### Replication Modes

**1. No Replication (default for single-node setups):**
```go
distConfig.ReplicationFactor = 1
distConfig.ReplicationMode = distributed.ReplicationNone
```
- Fastest writes
- No redundancy
- Use for development or non-critical data

**2. Async Replication (default for multi-node):**
```go
distConfig.ReplicationFactor = 2
distConfig.ReplicationMode = distributed.ReplicationAsync
```
- Primary write returns immediately
- Replicas updated in background
- Good balance of speed and durability

**3. Sync Replication:**
```go
distConfig.ReplicationFactor = 2
distConfig.ReplicationMode = distributed.ReplicationSync
```
- All replicas written before return
- Slower writes, strong consistency
- Use for critical data

### Dynamic Node Management

**Adding Nodes at Runtime:**
```go
newNode := distributed.NewNode("node-4", cache.NewMemoryCacheWithConfig(config))
dc.AddNode(newNode)
// Hash ring automatically rebalances
```

**Removing Nodes:**
```go
dc.RemoveNode("node-2")
// Keys automatically redistributed
```

**Health Monitoring:**
```go
// Check specific node
status, _ := dc.NodeHealth("node-1")
fmt.Println(status) // "healthy" or "unhealthy"

// Get all nodes
for _, node := range dc.Nodes() {
    fmt.Printf("%s: %s\n", node.ID(), node.HealthStatus())
}
```

### Metrics

```go
metrics := dc.GetMetrics()
fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
fmt.Printf("Failover count: %d\n", metrics.FailoverCount)
fmt.Printf("Healthy nodes: %d\n", metrics.HealthyNodes)
fmt.Printf("Rebalance events: %d\n", metrics.RebalanceEvents)
```

---

## Combining Features

### Distributed Leaderboard

You can combine distributed caching with leaderboard functionality:

**Option 1: Distributed Cache with Leaderboard Nodes**
```go
// Create leaderboard cache instances
lbConfig := cache.DefaultLeaderboardConfig()
nodes := []distributed.CacheNode{
    distributed.NewNode("region-us", cache.NewMemoryLeaderboardCache(config, lbConfig)),
    distributed.NewNode("region-eu", cache.NewMemoryLeaderboardCache(config, lbConfig)),
    distributed.NewNode("region-asia", cache.NewMemoryLeaderboardCache(config, lbConfig)),
}

// Each node supports both Cache and LeaderboardCache interfaces
distConfig := distributed.DefaultConfig()
distConfig.ReplicationFactor = 2
dc := distributed.New(nodes, distConfig)

// Use as distributed cache
dc.Set(ctx, "session:123", sessionData, time.Hour)

// Access leaderboard features on specific nodes
// (In production, you'd cast the node's cache to LeaderboardCache)
```

**Option 2: Separate Leaderboard and Cache Clusters**
```go
// Dedicated leaderboard cluster
lbCluster := cache.NewMemoryLeaderboardCache(config, lbConfig)

// Separate distributed cache cluster for sessions/data
cacheCluster := distributed.New(cacheNodes, distConfig)

// Use each for its purpose
lbCluster.ZAdd(ctx, "scores", ...)
cacheCluster.Set(ctx, "session:123", ...)
```

### Use Case: Global Gaming Platform

```go
// Regional leaderboards, each sharded across multiple nodes
type GamePlatform struct {
    cache       *distributed.DistributedCache
    leaderboard *cache.MemoryLeaderboardCache
}

func (p *GamePlatform) UpdateScore(playerID string, points float64) error {
    // Update distributed session
    session := fmt.Sprintf("player:%s", playerID)
    p.cache.Set(ctx, session, playerData, time.Hour)

    // Update global leaderboard
    p.leaderboard.ZIncrBy(ctx, "global:scores", playerID, points)

    return nil
}

func (p *GamePlatform) GetTopPlayers(n int) ([]*cache.ZMember, error) {
    return p.leaderboard.ZRange(ctx, "global:scores", 0, n-1, true)
}
```

---

## Performance Considerations

### Leaderboard Performance

**Time Complexity:**
- `ZAdd`: O(log N) where N = members in set
- `ZRange`: O(log N + M) where M = range size
- `ZRank`: O(log N)
- `ZScore`: O(N) - linear scan through level 0
- `ZRem`: O(log N)

**Memory Usage:**
- Skip list: ~120 bytes per member (32 levels × ~4 bytes/pointer + member + score)
- Configure `MaxMembersPerSet` based on available memory

**Benchmarks (from tests):**
- ZAdd: ~230 ns/op
- ZRange: ~8 µs/op (100 members)
- ZRank: ~400 ns/op

**Optimization Tips:**
```go
// For large datasets, use ZRangeByScore instead of ZRange when possible
players, _ := lbc.ZRangeByScore(ctx, "scores", 1000, 2000, true)

// Batch operations
members := []*cache.ZMember{
    {Member: "p1", Score: 100},
    {Member: "p2", Score: 200},
    // ... hundreds more
}
lbc.ZAdd(ctx, "scores", members...) // Single operation
```

### Distributed Cache Performance

**Hash Ring Lookups:**
- O(log N) where N = total virtual nodes
- Benchmark: ~165 ns/op

**Virtual Nodes:**
- More nodes = better distribution, slightly slower lookups
- Recommended: 150-300 per physical node
- Trade-off: distribution quality vs. memory

**Replication Impact:**

| Mode | Write Latency | Read Latency | Consistency |
|------|---------------|--------------|-------------|
| None | Fastest | Fastest | None |
| Async | Fast (~10% overhead) | Fast | Eventual |
| Sync | Slower (waits for all) | Fast | Strong |

**Network Considerations:**
```go
// For remote nodes, configure health check intervals
distConfig.HealthCheckInterval = 30 * time.Second // Reduce network overhead
distConfig.HealthCheckTimeout = 5 * time.Second   // Allow for network latency
```

---

## Common Migration Patterns

### Pattern 1: Gradual Rollout

**Phase 1 - Run in Parallel:**
```go
// Keep old cache, add new leaderboard
oldCache := cache.NewMemoryCacheWithConfig(config)
newLeaderboard := cache.NewMemoryLeaderboardCache(config, lbConfig)

// Write to both
oldCache.Set(ctx, "scores:alice", []byte("1250"), time.Hour)
newLeaderboard.ZAdd(ctx, "scores", &cache.ZMember{Member: "alice", Score: 1250})

// Read from new, fallback to old
top, err := newLeaderboard.ZRange(ctx, "scores", 0, 9, true)
if err != nil {
    // Fallback to old cache
}
```

**Phase 2 - Migrate Data:**
```go
// Migrate existing score data
var cursor uint64
for {
    keys, nextCursor, _ := oldCache.Scan(ctx, cursor, "scores:*", 100)
    for _, key := range keys {
        value, _ := oldCache.Get(ctx, key)
        member := strings.TrimPrefix(key, "scores:")
        score, _ := strconv.ParseFloat(string(value), 64)

        newLeaderboard.ZAdd(ctx, "scores", &cache.ZMember{
            Member: member,
            Score:  score,
        })
    }

    if nextCursor == 0 {
        break
    }
    cursor = nextCursor
}
```

**Phase 3 - Switch:**
```go
// Remove old cache, use new leaderboard only
oldCache.Close()
```

### Pattern 2: Feature Flags

```go
type Config struct {
    UseLeaderboard   bool
    UseDistributed   bool
}

func NewCacheService(cfg Config) CacheService {
    if cfg.UseDistributed {
        return &DistributedCacheService{...}
    } else if cfg.UseLeaderboard {
        return &LeaderboardCacheService{...}
    }
    return &BasicCacheService{...}
}
```

### Pattern 3: Adapter Pattern

**Wrap Old Interface:**
```go
// Adapter for legacy code expecting Cache interface
type LeaderboardAdapter struct {
    lbc cache.LeaderboardCache
}

func (a *LeaderboardAdapter) Get(ctx context.Context, key string) ([]byte, error) {
    // Parse key format: "scores:member"
    parts := strings.Split(key, ":")
    score, err := a.lbc.ZScore(ctx, parts[0], parts[1])
    if err != nil {
        return nil, err
    }
    return []byte(fmt.Sprintf("%.0f", score)), nil
}

// Implement other Cache methods...
```

---

## Breaking Changes

### None in v1.1 and v1.2

**Guaranteed Compatibility:**
- All existing `Cache` interface methods unchanged
- Default configurations remain the same
- No changes to method signatures
- No removal of public APIs

### Deprecations

**None planned** for v1.x series.

### Future Considerations (v2.0+)

**Potential changes under discussion:**
- Unified `Cache` interface including leaderboard methods
- Context-based configuration instead of global Config
- Metric interfaces for pluggable monitoring

**We will provide:**
- 6-month deprecation notice
- Detailed migration guide
- Compatibility shim packages

---

## Getting Help

### Documentation

- [API Reference](./API_REFERENCE.md) - Complete API documentation
- [Design Document](./DESIGN_EXTENSIONS.md) - Architecture details
- [Architecture Diagrams](./ARCHITECTURE.md) - Visual reference

### Examples

- **Basic Leaderboard**: `examples/cache-leaderboard/main.go`
- **Distributed Cache**: `examples/cache-distributed/main.go`
- **Combined Features**: `examples/cache-combined/main.go`

### Common Issues

**Q: Can I use LeaderboardCache without changing existing code?**

A: Yes! Create a `LeaderboardCache` instance alongside your existing cache. They can coexist.

**Q: What's the performance overhead of distributed caching?**

A: Minimal. Hash ring lookup is ~165 ns/op. For async replication, write overhead is ~10%.

**Q: Can I mix cache types (basic + leaderboard nodes) in distributed setup?**

A: Yes, but be careful. The distributed cache expects all nodes to support the same interface. Use type assertions when accessing leaderboard methods.

**Q: How do I migrate 1M+ items to leaderboard?**

A: Batch migrations in the background. Process in chunks of 10K-100K items. Monitor memory usage with `MaxMembersPerSet`.

**Q: What happens to data when I add/remove nodes?**

A: Consistent hashing minimizes data movement. Only ~(1/N) keys rebalance when adding the Nth node. Existing replicas remain accessible.

---

## Changelog

### v1.2 (2026-02-01)
- ✅ Added distributed caching with consistent hashing
- ✅ Health checking and automatic failover
- ✅ Dynamic node addition/removal
- ✅ Three replication modes (None, Async, Sync)

### v1.1 (2026-02-01)
- ✅ Added leaderboard support with skip list implementation
- ✅ Redis-compatible sorted set API (ZAdd, ZRange, ZRank, etc.)
- ✅ Leaderboard metrics tracking
- ✅ Configurable leaderboard limits

### v1.0 (Baseline)
- Basic in-memory cache
- LRU eviction
- TTL support
- Atomic operations (Incr/Decr)
- Append operation

---

## Support

For questions or issues, please:
- Check the examples directory
- Review the API reference
- File an issue on GitHub
- Contact the maintainers

Happy caching! 🚀
