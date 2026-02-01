# Cache Extensions API Reference

Quick reference for the proposed leaderboard and distributed cache APIs.

---

## Leaderboard API

### Basic Operations

```go
// Add or update members in leaderboard
err := cache.ZAdd(ctx, "game:scores",
    &cache.ZMember{Member: "player1", Score: 100.0},
    &cache.ZMember{Member: "player2", Score: 95.5},
)

// Remove members
err := cache.ZRem(ctx, "game:scores", "player1", "player2")

// Get score for a member
score, err := cache.ZScore(ctx, "game:scores", "player1")

// Increment score
newScore, err := cache.ZIncrBy(ctx, "game:scores", "player1", 10.0)
```

### Queries

```go
// Get top 10 players (highest scores)
top10, err := cache.ZRange(ctx, "game:scores", 0, 9, true)
// Returns: []*ZMember with Member and Score

// Get players ranked 10-20
players, err := cache.ZRange(ctx, "game:scores", 10, 20, true)

// Get all players with score between 90 and 100
players, err := cache.ZRangeByScore(ctx, "game:scores", 90.0, 100.0, true)

// Get rank of a player (0-based index)
rank, err := cache.ZRank(ctx, "game:scores", "player1", true) // desc=true for leaderboard
// rank=0 means first place
```

### Statistics

```go
// Count total members
count, err := cache.ZCard(ctx, "game:scores")

// Count members in score range
count, err := cache.ZCount(ctx, "game:scores", 90.0, 100.0)
```

### Bulk Removal

```go
// Remove bottom 10 players
removed, err := cache.ZRemRangeByRank(ctx, "game:scores", 0, 9)

// Remove players with score < 50
removed, err := cache.ZRemRangeByScore(ctx, "game:scores", 0, 50.0)
```

---

## Distributed Cache API

### Setup

```go
import "github.com/spcent/plumego/store/cache/distributed"

// Create cache nodes
node1 := cache.NewMemoryCache(cache.DefaultConfig())
node2 := cache.NewMemoryCache(cache.DefaultConfig())
node3 := cache.NewMemoryCache(cache.DefaultConfig())

// Create distributed cache
distCache := distributed.New(
    distributed.WithNodes(
        distributed.NewNode("node1", node1, distributed.WithWeight(1)),
        distributed.NewNode("node2", node2, distributed.WithWeight(1)),
        distributed.NewNode("node3", node3, distributed.WithWeight(2)), // 2x capacity
    ),
    distributed.WithVirtualNodes(150),
    distributed.WithReplication(2), // Store each key on 2 nodes
    distributed.WithFailover(distributed.FailoverNextNode),
)
```

### Basic Operations (Same as Cache interface)

```go
// Set - automatically sharded to correct node
err := distCache.Set(ctx, "user:123", userData, 10*time.Minute)

// Get - automatically routes to correct node
data, err := distCache.Get(ctx, "user:123")

// Delete
err := distCache.Delete(ctx, "user:123")

// Exists
exists, err := distCache.Exists(ctx, "user:123")
```

### Node Management

```go
// Add new node (triggers rebalancing)
newNode := cache.NewMemoryCache(cache.DefaultConfig())
err := distCache.AddNode(distributed.NewNode("node4", newNode))

// Remove node (triggers rebalancing)
err := distCache.RemoveNode("node2")

// Get all nodes
nodes := distCache.Nodes()
for _, node := range nodes {
    fmt.Printf("Node %s: healthy=%v\n", node.ID(), node.IsHealthy())
}
```

### Health & Monitoring

```go
// Check node health
health, err := distCache.NodeHealth("node1")
// Returns: HealthStatusHealthy, HealthStatusDegraded, HealthStatusUnhealthy

// Manual rebalance (normally automatic)
err := distCache.Rebalance()

// Get metrics
metrics := distCache.GetMetrics()
fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
fmt.Printf("Failover count: %d\n", metrics.FailoverCount)
fmt.Printf("Healthy nodes: %d\n", metrics.HealthyNodes)
```

---

## Combined: Distributed Leaderboard

```go
// Create distributed leaderboard cache
node1 := cache.NewMemoryLeaderboardCache(cfg, lbCfg)
node2 := cache.NewMemoryLeaderboardCache(cfg, lbCfg)

distLB := distributed.New(
    distributed.WithNodes(
        distributed.NewNode("node1", node1),
        distributed.NewNode("node2", node2),
    ),
)

// Use leaderboard features (automatically sharded)
distLB.ZAdd(ctx, "global:scores",
    &cache.ZMember{Member: "player1", Score: 100.0},
)

// Each leaderboard key is stored on a single node (for consistency)
// Different leaderboards can be on different nodes (for distribution)
```

---

## Configuration Reference

### LeaderboardConfig

```go
type LeaderboardConfig struct {
    MaxLeaderboards  int           // Max number of leaderboards (default: 1000)
    MaxMembersPerSet int           // Max members per leaderboard (default: 10000)
    DefaultTTL       time.Duration // Default expiration (default: 1 hour)
    EvictionPolicy   EvictionPolicy // LRU, LFU, or None (default: LRU)
}

// Create with defaults
cfg := cache.DefaultLeaderboardConfig()

// Create with custom values
cfg := &cache.LeaderboardConfig{
    MaxLeaderboards:  5000,
    MaxMembersPerSet: 50000,
    DefaultTTL:       24 * time.Hour,
    EvictionPolicy:   cache.EvictionLRU,
}
```

### DistributedConfig

```go
type DistributedConfig struct {
    VirtualNodes        int                 // Virtual nodes per physical node (default: 150)
    ReplicationFactor   int                 // Number of replicas (default: 1)
    HashAlgorithm       string              // "fnv1a", "murmur3", "xxhash" (default: "fnv1a")
    FailoverStrategy    FailoverStrategy    // NextNode, AllNodes, Retry (default: NextNode)
    ReplicationMode     ReplicationMode     // None, Async, Sync (default: Async)
    HealthCheckInterval time.Duration       // Health check frequency (default: 10s)
    HealthCheckTimeout  time.Duration       // Health check timeout (default: 2s)
    AutoRebalance       bool                // Auto-rebalance on topology changes (default: true)
}

// Create with defaults
cfg := distributed.DefaultConfig()

// Create with custom values
cfg := &distributed.Config{
    VirtualNodes:      300,  // Better distribution
    ReplicationFactor: 3,    // High availability
    ReplicationMode:   distributed.ReplicationSync, // Strong consistency
}
```

---

## Usage Examples

### Example 1: Gaming Leaderboard

```go
func setupGameLeaderboard() cache.LeaderboardCache {
    cfg := &cache.LeaderboardConfig{
        MaxLeaderboards:  100,   // 100 different game modes
        MaxMembersPerSet: 10000, // 10K players per mode
        DefaultTTL:       24 * time.Hour,
    }

    return cache.NewMemoryLeaderboardCache(cache.DefaultConfig(), cfg)
}

func updatePlayerScore(lb cache.LeaderboardCache, gameMode, playerID string, score float64) error {
    key := fmt.Sprintf("leaderboard:%s", gameMode)
    return lb.ZAdd(context.Background(), key, &cache.ZMember{
        Member: playerID,
        Score:  score,
    })
}

func getTopPlayers(lb cache.LeaderboardCache, gameMode string, limit int) ([]*cache.ZMember, error) {
    key := fmt.Sprintf("leaderboard:%s", gameMode)
    return lb.ZRange(context.Background(), key, 0, int64(limit-1), true)
}

func getPlayerRank(lb cache.LeaderboardCache, gameMode, playerID string) (int64, error) {
    key := fmt.Sprintf("leaderboard:%s", gameMode)
    rank, err := lb.ZRank(context.Background(), key, playerID, true)
    if err != nil {
        return -1, err
    }
    return rank + 1, nil // Convert 0-based to 1-based
}
```

### Example 2: Distributed Session Cache

```go
func setupDistributedSessions() cache.Cache {
    // Create 3 cache nodes
    nodes := make([]distributed.CacheNode, 3)
    for i := 0; i < 3; i++ {
        nodeCache := cache.NewMemoryCache(cache.DefaultConfig())
        nodes[i] = distributed.NewNode(
            fmt.Sprintf("node%d", i),
            nodeCache,
        )
    }

    // Create distributed cache with replication
    return distributed.New(
        distributed.WithNodes(nodes...),
        distributed.WithReplication(2), // Store each session on 2 nodes
        distributed.WithFailover(distributed.FailoverNextNode),
    )
}

func storeSession(cache cache.Cache, sessionID string, data []byte) error {
    key := fmt.Sprintf("session:%s", sessionID)
    return cache.Set(context.Background(), key, data, 30*time.Minute)
}

func getSession(cache cache.Cache, sessionID string) ([]byte, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    return cache.Get(context.Background(), key)
}
```

### Example 3: Real-time Analytics Dashboard

```go
type AnalyticsDashboard struct {
    lb cache.LeaderboardCache
}

func (d *AnalyticsDashboard) TrackPageView(url string) error {
    return d.lb.ZIncrBy(context.Background(), "analytics:pageviews", url, 1.0)
}

func (d *AnalyticsDashboard) GetTopPages(limit int) ([]*cache.ZMember, error) {
    return d.lb.ZRange(context.Background(), "analytics:pageviews", 0, int64(limit-1), true)
}

func (d *AnalyticsDashboard) GetPageViews(url string) (int64, error) {
    score, err := d.lb.ZScore(context.Background(), "analytics:pageviews", url)
    if err != nil {
        return 0, err
    }
    return int64(score), nil
}
```

### Example 4: Social Media Trending Topics

```go
type TrendingTopics struct {
    lb cache.LeaderboardCache
}

func (t *TrendingTopics) IncrementTopic(topic string, mentions int) error {
    return t.lb.ZIncrBy(
        context.Background(),
        "trending:topics",
        topic,
        float64(mentions),
    )
}

func (t *TrendingTopics) GetTrending(timeWindow string, limit int) ([]*cache.ZMember, error) {
    key := fmt.Sprintf("trending:topics:%s", timeWindow)
    return t.lb.ZRange(context.Background(), key, 0, int64(limit-1), true)
}

func (t *TrendingTopics) DecayScores() error {
    // Reduce all scores by 10% every hour (simulate decay)
    ctx := context.Background()
    members, err := t.lb.ZRange(ctx, "trending:topics", 0, -1, false)
    if err != nil {
        return err
    }

    for _, member := range members {
        newScore := member.Score * 0.9
        t.lb.ZAdd(ctx, "trending:topics", &cache.ZMember{
            Member: member.Member,
            Score:  newScore,
        })
    }
    return nil
}
```

---

## Performance Expectations

### Leaderboard Operations

| Operation | Complexity | Typical Latency | Max Throughput |
|-----------|-----------|----------------|----------------|
| ZAdd | O(log N) | < 100µs | ~100K ops/sec |
| ZRem | O(log N) | < 100µs | ~100K ops/sec |
| ZScore | O(1) | < 10µs | ~1M ops/sec |
| ZIncrBy | O(log N) | < 100µs | ~100K ops/sec |
| ZRange(100) | O(log N + M) | < 500µs | ~50K ops/sec |
| ZRank | O(log N) | < 50µs | ~200K ops/sec |
| ZCard | O(1) | < 5µs | ~2M ops/sec |

*N = total members, M = result size*

### Distributed Cache

| Metric | Expected Value | Notes |
|--------|---------------|-------|
| Hash lookup | < 10µs | With 150 vnodes per node |
| Network overhead | ~1-5ms | Depends on network latency |
| Failover time | < 100ms | Single node failure |
| Rebalance (10K keys) | < 5s | 10 nodes |
| Distribution variance | < 5% | With 150 vnodes |

---

## Error Handling

### Leaderboard Errors

```go
// Handle common errors
score, err := lb.ZScore(ctx, "scores", "player1")
switch {
case errors.Is(err, cache.ErrNotFound):
    // Player not in leaderboard
case errors.Is(err, cache.ErrLeaderboardNotFound):
    // Leaderboard doesn't exist
case errors.Is(err, cache.ErrInvalidScore):
    // Score is NaN or Inf
default:
    // Other error
}
```

### Distributed Cache Errors

```go
// Handle distributed cache errors
data, err := distCache.Get(ctx, "key")
switch {
case errors.Is(err, cache.ErrNotFound):
    // Key not found (checked all replicas)
case errors.Is(err, distributed.ErrNoNodesAvailable):
    // All nodes are down
case errors.Is(err, distributed.ErrNodeUnhealthy):
    // Primary node unhealthy, failover failed
default:
    // Other error
}
```

---

## Migration Guide

### From Basic Cache to Leaderboard

```go
// Before: Track scores with separate keys
cache.Set(ctx, "score:player1", []byte("100"), ttl)
cache.Set(ctx, "score:player2", []byte("95"), ttl)
// Problem: Can't efficiently get rankings

// After: Use leaderboard
lb.ZAdd(ctx, "scores",
    &cache.ZMember{Member: "player1", Score: 100},
    &cache.ZMember{Member: "player2", Score: 95},
)
// Benefits: Efficient rankings and range queries
```

### From Single Node to Distributed

```go
// Before: Single node cache
cache := cache.NewMemoryCache(cfg)

// After: Distributed cache
node1 := cache.NewMemoryCache(cfg)
node2 := cache.NewMemoryCache(cfg)

distCache := distributed.New(
    distributed.WithNodes(
        distributed.NewNode("node1", node1),
        distributed.NewNode("node2", node2),
    ),
)

// Same interface, automatic sharding
distCache.Set(ctx, key, value, ttl)
```

---

## Best Practices

### Leaderboard

1. **Use descriptive keys**: `game:mode:casual:leaderboard` instead of `lb1`
2. **Set appropriate TTL**: Daily/weekly/monthly leaderboards with matching TTL
3. **Limit member count**: Use `ZRemRangeByRank` to keep only top N
4. **Handle ties**: Consider secondary sort criteria (timestamp, user ID)
5. **Batch operations**: Add multiple members in one `ZAdd` call

### Distributed Cache

1. **Plan capacity**: Size nodes based on expected data distribution
2. **Use replication**: At least 2 replicas for production
3. **Monitor health**: Track failover count and node health
4. **Graceful rebalancing**: Do during low-traffic periods
5. **Key design**: Design keys to distribute evenly across nodes

---

## Troubleshooting

### Leaderboard Issues

**Problem**: Leaderboard queries are slow
- **Solution**: Check `MaxMembersPerSet`, consider pagination
- **Check**: Use `ZCard` to verify member count

**Problem**: Memory usage too high
- **Solution**: Reduce `MaxLeaderboards` or `MaxMembersPerSet`
- **Check**: Monitor `MetricsCollector.CurrentMemoryUsage`

**Problem**: Scores not updating
- **Solution**: Verify `ZIncrBy` delta is not zero
- **Check**: Use `ZScore` to verify current value

### Distributed Cache Issues

**Problem**: Uneven key distribution
- **Solution**: Increase `VirtualNodes` (try 300)
- **Check**: Monitor per-node metrics

**Problem**: High failover count
- **Solution**: Check node health, increase health check timeout
- **Check**: Use `NodeHealth()` to diagnose

**Problem**: Rebalancing takes too long
- **Solution**: Reduce keys per node or increase node capacity
- **Check**: Monitor `RebalanceEvents` metric

---

## Comparison with Redis

| Feature | Plumego | Redis | Notes |
|---------|---------|-------|-------|
| ZAdd/ZRem | ✅ | ✅ | Compatible API |
| ZRange | ✅ | ✅ | Compatible API |
| ZRank | ✅ | ✅ | Compatible API |
| ZIncrBy | ✅ | ✅ | Compatible API |
| ZUnionStore | ❌ | ✅ | Future enhancement |
| ZInterStore | ❌ | ✅ | Future enhancement |
| ZScan | ❌ | ✅ | Use ZRange instead |
| ZLexRange | ❌ | ✅ | Future enhancement |
| Persistence | Memory only | Disk + Memory | Use Redis adapter for persistence |
| Clustering | Client-side | Server-side | Different approach |

---

## Future Enhancements (Post-v1.0)

1. **ZUnionStore / ZInterStore**: Combine multiple leaderboards
2. **ZLexRange**: Range by lexicographical order
3. **ZScan**: Incremental iteration
4. **Persistent leaderboards**: Disk-backed storage
5. **Server-side clustering**: Built-in cluster protocol
6. **Geo-distributed caching**: Multi-region support
7. **Advanced eviction**: LFU, TTL-based eviction
8. **Compression**: Automatic value compression

---

## Support & Feedback

For questions or issues:
- 📖 See full design: `/store/cache/DESIGN_EXTENSIONS.md`
- 🐛 Report bugs: GitHub issues
- 💡 Feature requests: GitHub discussions
- 📧 Security issues: See `SECURITY.md`
