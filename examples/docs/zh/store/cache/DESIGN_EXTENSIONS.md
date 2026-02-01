# Cache Package Extensions Design

**Version:** 1.0
**Date:** 2026-02-01
**Status:** Design Phase

---

## Executive Summary

This document outlines the design for extending the `store/cache` package with two major features:

1. **Leaderboard Support** - Sorted set functionality for ranking and scoring
2. **Distributed Key Support** - Consistent hashing and sharding for distributed caching

Both features maintain backward compatibility and follow plumego's design principles: minimal dependencies, standard library first, and composable architecture.

---

## Current State Analysis

### Existing Features
- ✅ Basic key-value operations (Get/Set/Delete)
- ✅ Atomic operations (Incr/Decr/Append)
- ✅ TTL-based expiration
- ✅ Memory tracking and limits
- ✅ Metrics collection
- ✅ HTTP middleware caching
- ✅ Redis adapter

### Current Gaps
- ❌ Sorted sets / leaderboards
- ❌ Distributed caching / sharding
- ❌ LRU/LFU eviction policies
- ❌ Bulk operations
- ❌ Advanced data structures (lists, sets, hashes)

---

## Feature 1: Leaderboard Support

### Overview

Leaderboards require **sorted set** functionality where:
- Each member has a score (float64)
- Members are sorted by score (ascending or descending)
- Fast lookups by rank or score range
- Efficient updates and deletions

### Use Cases

1. **Gaming**: Player rankings, high scores
2. **Social**: User reputation, karma points
3. **Analytics**: Top products, popular content
4. **Real-time**: Live event rankings, trending topics

### Interface Design

```go
// LeaderboardCache extends Cache with sorted set operations
type LeaderboardCache interface {
    Cache // Embeds existing cache interface

    // Sorted Set Operations
    ZAdd(ctx context.Context, key string, members ...*ZMember) error
    ZRem(ctx context.Context, key string, members ...string) error
    ZScore(ctx context.Context, key string, member string) (float64, error)
    ZIncrBy(ctx context.Context, key string, member string, delta float64) (float64, error)

    // Range Queries
    ZRange(ctx context.Context, key string, start, stop int64, desc bool) ([]*ZMember, error)
    ZRangeByScore(ctx context.Context, key string, min, max float64, desc bool) ([]*ZMember, error)
    ZRank(ctx context.Context, key string, member string, desc bool) (int64, error)

    // Cardinality
    ZCard(ctx context.Context, key string) (int64, error)
    ZCount(ctx context.Context, key string, min, max float64) (int64, error)

    // Removal by Range
    ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error)
    ZRemRangeByScore(ctx context.Context, key string, min, max float64) (int64, error)
}

// ZMember represents a member in a sorted set
type ZMember struct {
    Member string
    Score  float64
}

// ZRangeOptions provides fine-grained control over range queries
type ZRangeOptions struct {
    WithScores bool
    Offset     int64
    Count      int64
    Desc       bool
}
```

### Data Structure Design

#### In-Memory Implementation

```go
type sortedSet struct {
    mu      sync.RWMutex

    // Map member -> score for O(1) score lookup
    scores  map[string]float64

    // Skip list or balanced tree for O(log N) range queries
    // We'll use a skip list for simplicity and performance
    skipList *skipList

    // Expiration time
    expiration time.Time
}

// Skip list node
type skipListNode struct {
    member  string
    score   float64
    forward []*skipListNode // Pointers to next nodes at each level
    span    []int64         // Distance to next node at each level (for rank calculation)
}

type skipList struct {
    header *skipListNode
    tail   *skipListNode
    length int64
    level  int32
}
```

**Skip List Advantages:**
- O(log N) insertion, deletion, and search
- O(log N) rank calculation
- O(log N + M) range queries (M = result size)
- Simpler than red-black trees
- Better cache locality than tree structures

#### Alternative: B-Tree

For very large leaderboards, consider using a B-tree:
```go
import "github.com/google/btree" // Optional dependency for extreme cases
```

**Trade-offs:**
- Skip list: Simpler, standard library only, good for most cases
- B-tree: Better for huge datasets (100K+ items), requires dependency

**Decision:** Start with skip list, provide B-tree option later if needed.

### Memory Storage Layout

```go
type MemoryLeaderboardCache struct {
    *MemoryCache // Embed existing cache

    leaderboards sync.Map // key -> *sortedSet
}
```

### Performance Characteristics

| Operation | Time Complexity | Space Complexity |
|-----------|----------------|------------------|
| ZAdd | O(log N) | O(1) per member |
| ZRem | O(log N) | O(1) |
| ZScore | O(1) | - |
| ZIncrBy | O(log N) | - |
| ZRange | O(log N + M) | O(M) |
| ZRangeByScore | O(log N + M) | O(M) |
| ZRank | O(log N) | - |
| ZCard | O(1) | - |

**N** = number of members in sorted set
**M** = number of returned results

### TTL and Cleanup

```go
// Leaderboard-specific config
type LeaderboardConfig struct {
    MaxLeaderboards int           // Default: 1000
    MaxMembersPerSet int          // Default: 10000
    DefaultTTL      time.Duration // Default: 1 hour
}
```

**Cleanup Strategy:**
1. Periodic cleanup checks leaderboard expiration
2. Lazy expiration on access
3. LRU eviction when `MaxLeaderboards` exceeded

### Redis Adapter

```go
type RedisLeaderboardCache struct {
    *RedisCache // Embed existing adapter
    client RedisZSetClient
}

type RedisZSetClient interface {
    ZAdd(ctx context.Context, key string, members ...redis.Z) error
    ZRem(ctx context.Context, key string, members ...string) error
    ZScore(ctx context.Context, key string, member string) (float64, error)
    ZIncrBy(ctx context.Context, key string, member string, delta float64) (float64, error)
    ZRange(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
    ZRevRange(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
    ZRangeByScore(ctx context.Context, key string, min, max string) ([]redis.Z, error)
    ZRank(ctx context.Context, key string, member string) (int64, error)
    ZRevRank(ctx context.Context, key string, member string) (int64, error)
    ZCard(ctx context.Context, key string) (int64, error)
    ZCount(ctx context.Context, key string, min, max string) (int64, error)
    ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error)
    ZRemRangeByScore(ctx context.Context, key string, min, max string) (int64, error)
}
```

### Error Handling

New error types:
```go
var (
    ErrLeaderboardNotFound = errors.New("cache: leaderboard not found")
    ErrMemberNotFound      = errors.New("cache: member not found in sorted set")
    ErrLeaderboardFull     = errors.New("cache: leaderboard member limit reached")
    ErrInvalidScore        = errors.New("cache: invalid score value")
    ErrInvalidRange        = errors.New("cache: invalid range parameters")
)
```

### Metrics

Extend `MetricsCollector`:
```go
type LeaderboardMetrics struct {
    ZAdds              uint64
    ZRems              uint64
    ZRangeQueries      uint64
    ZScoreLookups      uint64
    ZRankCalculations  uint64
    TotalLeaderboards  int64
    TotalMembers       int64
    AvgMembersPerSet   float64
}
```

---

## Feature 2: Distributed Key Support

### Overview

Distributed caching allows:
- **Horizontal scaling** - Spread cache across multiple nodes
- **Consistent hashing** - Minimize key redistribution on node changes
- **Sharding** - Partition data for better performance
- **Replication** - Improve availability and fault tolerance

### Use Cases

1. **Large-scale applications**: Cache doesn't fit on single node
2. **Multi-region deployment**: Geographic distribution
3. **High availability**: Redundancy and failover
4. **Load balancing**: Distribute read/write load

### Architecture Options

#### Option A: Client-Side Sharding (Recommended)

```
┌──────────┐
│  Client  │
│ (Cache)  │
└────┬─────┘
     │
     ├─── Consistent Hash Ring ───┐
     │                             │
     v                             v
┌─────────┐                   ┌─────────┐
│ Node 1  │                   │ Node 2  │
│ (Cache) │                   │ (Cache) │
└─────────┘                   └─────────┘
```

**Pros:**
- No external dependencies
- Simple implementation
- No single point of failure
- Works with any cache backend

**Cons:**
- No automatic rebalancing
- Client must track node health
- Manual sharding logic

#### Option B: Proxy-Based Sharding

```
┌──────────┐
│  Client  │
└────┬─────┘
     │
     v
┌──────────┐
│  Proxy   │ ← Twemproxy, Envoy, etc.
└────┬─────┘
     │
     ├──────────┬──────────┐
     v          v          v
 Node 1     Node 2     Node 3
```

**Pros:**
- Transparent to client
- Centralized routing logic
- Advanced features (pooling, routing)

**Cons:**
- External dependency
- Single point of failure (needs HA)
- Added latency

#### Option C: Distributed Hash Table (DHT)

Use Raft/etcd for coordination.

**Pros:**
- Automatic rebalancing
- Strong consistency
- Fault tolerance

**Cons:**
- Complex implementation
- Requires consensus protocol
- Heavy dependency

**Decision:** Implement **Option A (Client-Side Sharding)** first:
- Aligns with plumego's minimal dependencies
- Composable with existing cache interface
- Users can add proxy/DHT if needed

### Interface Design

```go
// DistributedCache provides sharded cache access
type DistributedCache interface {
    Cache // Same interface as single-node cache

    // Node Management
    AddNode(node CacheNode) error
    RemoveNode(nodeID string) error
    Nodes() []CacheNode

    // Health & Monitoring
    NodeHealth(nodeID string) (HealthStatus, error)
    Rebalance() error
}

// CacheNode represents a cache instance in the cluster
type CacheNode interface {
    ID() string
    Cache() Cache
    Weight() int // For weighted distribution
    IsHealthy() bool
}

// HashRing provides consistent hashing
type HashRing interface {
    Add(node CacheNode) error
    Remove(nodeID string) error
    Get(key string) (CacheNode, error)
    GetN(key string, n int) ([]CacheNode, error) // For replication
}
```

### Consistent Hashing Implementation

```go
type consistentHashRing struct {
    mu           sync.RWMutex
    ring         map[uint32]CacheNode // Hash -> Node
    sortedHashes []uint32             // Sorted hash values
    nodes        map[string]CacheNode // Node ID -> Node
    virtualNodes int                  // Default: 150
    hashFunc     hash.Hash32          // Default: FNV-1a
}

// Hash distribution with virtual nodes
func (r *consistentHashRing) hash(key string, vnode int) uint32 {
    r.hashFunc.Reset()
    r.hashFunc.Write([]byte(fmt.Sprintf("%s#%d", key, vnode)))
    return r.hashFunc.Sum32()
}

// Get node for key using binary search
func (r *consistentHashRing) Get(key string) (CacheNode, error) {
    if len(r.sortedHashes) == 0 {
        return nil, ErrNoNodesAvailable
    }

    h := r.hash(key, 0)

    // Binary search for closest hash
    idx := sort.Search(len(r.sortedHashes), func(i int) bool {
        return r.sortedHashes[i] >= h
    })

    // Wrap around if needed
    if idx >= len(r.sortedHashes) {
        idx = 0
    }

    return r.ring[r.sortedHashes[idx]], nil
}
```

### Configuration

```go
type DistributedConfig struct {
    VirtualNodes    int           // Default: 150 (per physical node)
    ReplicationFactor int         // Default: 1 (no replication)
    HashAlgorithm   string        // "fnv1a", "murmur3", "xxhash"
    FailoverEnabled bool          // Auto-failover on node failure
    HealthCheckInterval time.Duration // Default: 10s
    HealthCheckTimeout  time.Duration // Default: 2s
}
```

### Replication Strategy

```go
type ReplicationMode int

const (
    ReplicationNone ReplicationMode = iota
    ReplicationAsync  // Write to primary, async replicate
    ReplicationSync   // Write to all replicas (slower, consistent)
)

type DistributedCache struct {
    ring        HashRing
    replication ReplicationMode
    replicaCount int
}

// Set with replication
func (dc *DistributedCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    nodes, err := dc.ring.GetN(key, dc.replicaCount)
    if err != nil {
        return err
    }

    switch dc.replication {
    case ReplicationSync:
        return dc.setSyncReplicas(ctx, nodes, key, value, ttl)
    case ReplicationAsync:
        return dc.setAsyncReplicas(ctx, nodes, key, value, ttl)
    default:
        return nodes[0].Cache().Set(ctx, key, value, ttl)
    }
}
```

### Failover Handling

```go
type FailoverStrategy int

const (
    FailoverNextNode FailoverStrategy = iota // Try next node in ring
    FailoverAllNodes                         // Broadcast to all healthy nodes
    FailoverRetry                            // Retry same node with backoff
)

func (dc *DistributedCache) Get(ctx context.Context, key string) ([]byte, error) {
    node, err := dc.ring.Get(key)
    if err != nil {
        return nil, err
    }

    value, err := node.Cache().Get(ctx, key)
    if err == nil {
        return value, nil
    }

    // Failover if node is unhealthy
    if !node.IsHealthy() {
        return dc.failoverGet(ctx, key)
    }

    return nil, err
}

func (dc *DistributedCache) failoverGet(ctx context.Context, key string) ([]byte, error) {
    nodes, _ := dc.ring.GetN(key, dc.replicaCount)

    for _, node := range nodes[1:] { // Skip primary (already failed)
        if !node.IsHealthy() {
            continue
        }

        value, err := node.Cache().Get(ctx, key)
        if err == nil {
            return value, nil
        }
    }

    return nil, ErrNotFound
}
```

### Metrics & Monitoring

```go
type DistributedMetrics struct {
    // Per-node metrics
    NodeMetrics map[string]*MetricsSnapshot

    // Cluster-wide metrics
    TotalRequests    uint64
    FailoverCount    uint64
    ReplicationLag   time.Duration
    HashCollisions   uint64

    // Health status
    HealthyNodes     int
    UnhealthyNodes   int
    RebalanceEvents  uint64
}
```

### Migration & Rebalancing

```go
// Rebalance redistributes keys when nodes change
func (dc *DistributedCache) Rebalance() error {
    // 1. Track all keys and their current nodes
    keyMap := dc.scanAllKeys()

    // 2. For each key, check if it should move
    migrations := make(map[string][]string) // oldNode -> []keys

    for key, oldNode := range keyMap {
        newNode, _ := dc.ring.Get(key)
        if newNode.ID() != oldNode {
            migrations[oldNode] = append(migrations[oldNode], key)
        }
    }

    // 3. Migrate keys
    for oldNodeID, keys := range migrations {
        if err := dc.migrateKeys(oldNodeID, keys); err != nil {
            return err
        }
    }

    return nil
}
```

---

## Implementation Plan

### Phase 1: Leaderboard Support (Week 1-2)

**Week 1: Core Implementation**
1. ✅ Design skip list data structure
2. ✅ Implement `LeaderboardCache` interface
3. ✅ Implement `MemoryLeaderboardCache`
4. ✅ Add ZAdd, ZRem, ZScore, ZIncrBy
5. ✅ Add range query operations
6. ✅ Add TTL and cleanup

**Week 2: Redis & Testing**
1. ✅ Implement Redis adapter
2. ✅ Write comprehensive tests
3. ✅ Add benchmarks
4. ✅ Update documentation
5. ✅ Add examples

**Deliverables:**
- `store/cache/leaderboard.go` - Main implementation
- `store/cache/skiplist.go` - Skip list data structure
- `store/cache/leaderboard_test.go` - Tests
- `store/cache/redis/leaderboard.go` - Redis adapter
- `examples/leaderboard/` - Example application

### Phase 2: Distributed Cache (Week 3-4)

**Week 3: Consistent Hashing**
1. ✅ Implement hash ring
2. ✅ Implement virtual nodes
3. ✅ Add multiple hash algorithms
4. ✅ Implement node management
5. ✅ Add health checks

**Week 4: Replication & Testing**
1. ✅ Implement replication strategies
2. ✅ Add failover logic
3. ✅ Implement rebalancing
4. ✅ Write comprehensive tests
5. ✅ Add documentation & examples

**Deliverables:**
- `store/cache/distributed/` - New package
  - `hashring.go` - Consistent hashing
  - `distributed.go` - Distributed cache
  - `replication.go` - Replication logic
  - `health.go` - Health checking
  - `*_test.go` - Tests
- `examples/distributed/` - Example application

### Phase 3: Integration & Documentation (Week 5)

1. ✅ Combine leaderboard + distributed
2. ✅ Performance benchmarks
3. ✅ Migration guide
4. ✅ API documentation
5. ✅ Example applications

---

## Testing Strategy

### Unit Tests

```go
// Leaderboard tests
TestZAdd()
TestZRem()
TestZScore()
TestZIncrBy()
TestZRange()
TestZRangeByScore()
TestZRank()
TestZCard()
TestZCount()
TestLeaderboardExpiration()
TestLeaderboardMemoryLimit()
TestConcurrentLeaderboard()

// Distributed tests
TestHashRingAddNode()
TestHashRingRemoveNode()
TestHashRingDistribution()
TestConsistentHashing()
TestFailover()
TestReplication()
TestRebalance()
TestNodeHealth()
TestConcurrentDistributed()
```

### Benchmarks

```go
BenchmarkZAdd
BenchmarkZRange
BenchmarkZRank
BenchmarkDistributedSet
BenchmarkDistributedGet
BenchmarkHashRingLookup
BenchmarkReplication
```

### Integration Tests

1. **Leaderboard + Redis**: Verify parity
2. **Distributed + Multiple Nodes**: Verify sharding
3. **Failover**: Simulate node failures
4. **Rebalancing**: Add/remove nodes dynamically

---

## Performance Targets

### Leaderboard

| Operation | Target | Measurement |
|-----------|--------|-------------|
| ZAdd | < 100µs | p99 latency |
| ZRange(100) | < 500µs | p99 latency |
| ZRank | < 50µs | p99 latency |
| Memory per member | < 100 bytes | Average |

### Distributed Cache

| Metric | Target | Notes |
|--------|--------|-------|
| Hash lookup | < 10µs | With 1000 vnodes |
| Failover time | < 100ms | Single node failure |
| Rebalance time | < 5s | 10K keys, 10 nodes |
| Hash distribution | < 5% variance | Standard deviation |

---

## Backward Compatibility

### API Compatibility

All existing `Cache` interface methods remain unchanged:
```go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Clear(ctx context.Context) error
    Incr(ctx context.Context, key string, delta int64) (int64, error)
    Decr(ctx context.Context, key string, delta int64) (int64, error)
    Append(ctx context.Context, key string, data []byte) error
}
```

### Opt-In Design

New features are opt-in:
```go
// Existing code continues to work
cache := cache.NewMemoryCache(cfg)

// New leaderboard features (opt-in)
lbCache := cache.NewMemoryLeaderboardCache(cfg, lbCfg)

// New distributed features (opt-in)
distCache := distributed.New(nodes, distCfg)
```

---

## Security Considerations

### Leaderboard

1. **Score validation**: Prevent NaN/Inf scores
2. **Member sanitization**: Validate member names
3. **Size limits**: Prevent memory exhaustion
4. **Rate limiting**: Prevent abuse of ZAdd/ZIncrBy

### Distributed Cache

1. **Node authentication**: TLS/mTLS for inter-node communication
2. **Key migration**: Atomic operations during rebalancing
3. **Health checks**: Prevent cascade failures
4. **Replay protection**: Timestamp-based validation

---

## Open Questions

### Leaderboard

1. **Eviction policy**: LRU for entire leaderboards or per-member?
2. **Score precision**: float64 vs decimal for currency/points?
3. **Negative scores**: Support or disallow?
4. **Ties**: How to handle identical scores? (secondary sort by member name?)

### Distributed Cache

1. **Service discovery**: Manual node list or integrate with etcd/consul?
2. **Network protocol**: HTTP/gRPC/TCP for inter-node communication?
3. **Serialization**: GOB/JSON/Protobuf for replication?
4. **Split-brain**: How to handle network partitions?

---

## Alternative Approaches

### Leaderboard Alternatives

**1. B-Tree instead of Skip List**
- Pros: Better for very large sets (100K+ members)
- Cons: Requires external dependency

**2. Heap + Hash Map**
- Pros: Simpler implementation
- Cons: No efficient range queries

**3. Radix Tree**
- Pros: Memory efficient for string keys
- Cons: Complex implementation

**Decision**: Skip list is the best balance of performance, simplicity, and no dependencies.

### Distributed Alternatives

**1. Redis Cluster Protocol**
- Pros: Standard protocol, many client libraries
- Cons: Requires Redis as dependency

**2. Raft Consensus**
- Pros: Strong consistency guarantees
- Cons: Complex, requires consensus protocol

**3. Rendezvous Hashing**
- Pros: Minimal data movement on rebalance
- Cons: Higher CPU cost per lookup

**Decision**: Consistent hashing with virtual nodes is proven and well-understood.

---

## Migration Path

### For Existing Users

**No changes required** - Existing code continues to work:
```go
// Existing code (no changes)
cache := cache.NewMemoryCache(cache.DefaultConfig())
cache.Set(ctx, "key", []byte("value"), 5*time.Minute)
```

### To Use Leaderboards

```go
// Option 1: New cache instance
lbCache := cache.NewMemoryLeaderboardCache(
    cache.DefaultConfig(),
    cache.DefaultLeaderboardConfig(),
)

// Option 2: Type assertion
if lb, ok := cache.(cache.LeaderboardCache); ok {
    lb.ZAdd(ctx, "scores", &cache.ZMember{
        Member: "player1",
        Score:  100.0,
    })
}
```

### To Use Distributed Cache

```go
// Create nodes
node1 := cache.NewMemoryCache(cfg)
node2 := cache.NewMemoryCache(cfg)

// Create distributed cache
distCache := distributed.New(
    distributed.WithNodes(
        distributed.NewNode("node1", node1),
        distributed.NewNode("node2", node2),
    ),
    distributed.WithReplication(2),
)

// Use same interface
distCache.Set(ctx, "key", value, ttl)
```

---

## Success Metrics

### Technical Metrics

- ✅ All tests pass
- ✅ Benchmarks meet performance targets
- ✅ Zero breaking changes
- ✅ Test coverage > 90%
- ✅ Documentation complete

### User Metrics

- ✅ Easy to adopt (< 10 lines of code to enable)
- ✅ Clear examples provided
- ✅ Migration guide available
- ✅ Performance characteristics documented

---

## Timeline Summary

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| Design | 1 week | This document |
| Leaderboard Implementation | 2 weeks | Working leaderboard cache |
| Distributed Implementation | 2 weeks | Working distributed cache |
| Testing & Documentation | 1 week | Production-ready release |
| **Total** | **6 weeks** | v1.0 release |

---

## Next Steps

1. ✅ Review and approve this design
2. ⏳ Create GitHub issues for each phase
3. ⏳ Set up project board
4. ⏳ Begin Phase 1 implementation
5. ⏳ Schedule weekly design reviews

---

## Appendix A: Skip List Detailed Design

### Structure

```go
const maxLevel = 32 // Maximum skip list level

type skipList struct {
    header *skipListNode
    tail   *skipListNode
    length int64
    level  int32 // Current max level
}

type skipListNode struct {
    member  string
    score   float64
    backward *skipListNode           // For reverse traversal
    forward  [maxLevel]*skipListNode // Forward pointers
    span     [maxLevel]int64         // Span for rank calculation
}
```

### Level Generation

```go
func randomLevel() int32 {
    level := int32(1)
    for rand.Float64() < 0.25 && level < maxLevel {
        level++
    }
    return level
}
```

**Probability**: P = 0.25 (matches Redis implementation)
- Level 1: 100% of nodes
- Level 2: 25% of nodes
- Level 3: 6.25% of nodes
- Level 4: 1.56% of nodes

### Insert Algorithm

```go
func (sl *skipList) insert(member string, score float64) *skipListNode {
    update := make([]*skipListNode, maxLevel)
    rank := make([]int64, maxLevel)

    x := sl.header

    // Find insertion position
    for i := sl.level - 1; i >= 0; i-- {
        if i == sl.level-1 {
            rank[i] = 0
        } else {
            rank[i] = rank[i+1]
        }

        for x.forward[i] != nil &&
            (x.forward[i].score < score ||
                (x.forward[i].score == score && x.forward[i].member < member)) {
            rank[i] += x.span[i]
            x = x.forward[i]
        }
        update[i] = x
    }

    // Create new node
    level := randomLevel()
    if level > sl.level {
        for i := sl.level; i < level; i++ {
            rank[i] = 0
            update[i] = sl.header
            update[i].span[i] = sl.length
        }
        sl.level = level
    }

    x = &skipListNode{
        member:  member,
        score:   score,
        forward: make([]*skipListNode, maxLevel),
        span:    make([]int64, maxLevel),
    }

    // Update forward pointers
    for i := int32(0); i < level; i++ {
        x.forward[i] = update[i].forward[i]
        update[i].forward[i] = x

        x.span[i] = update[i].span[i] - (rank[0] - rank[i])
        update[i].span[i] = (rank[0] - rank[i]) + 1
    }

    // Update spans for untouched levels
    for i := level; i < sl.level; i++ {
        update[i].span[i]++
    }

    // Set backward pointer
    if update[0] == sl.header {
        x.backward = nil
    } else {
        x.backward = update[0]
    }

    if x.forward[0] != nil {
        x.forward[0].backward = x
    } else {
        sl.tail = x
    }

    sl.length++
    return x
}
```

### Range Query Algorithm

```go
func (sl *skipList) getRange(start, stop int64, desc bool) []*ZMember {
    if start < 0 {
        start = sl.length + start
    }
    if stop < 0 {
        stop = sl.length + stop
    }

    if start < 0 {
        start = 0
    }
    if stop >= sl.length {
        stop = sl.length - 1
    }

    rangeLen := stop - start + 1
    if rangeLen <= 0 {
        return nil
    }

    // Traverse to start position
    var x *skipListNode
    if desc {
        x = sl.tail
        if start > 0 {
            x = sl.getNodeByRank(sl.length - start)
        }
    } else {
        x = sl.header.forward[0]
        if start > 0 {
            x = sl.getNodeByRank(start + 1)
        }
    }

    // Collect results
    results := make([]*ZMember, 0, rangeLen)
    for i := int64(0); i < rangeLen && x != nil; i++ {
        results = append(results, &ZMember{
            Member: x.member,
            Score:  x.score,
        })
        if desc {
            x = x.backward
        } else {
            x = x.forward[0]
        }
    }

    return results
}
```

---

## Appendix B: Hash Function Comparison

| Algorithm | Speed | Quality | Collision Rate |
|-----------|-------|---------|----------------|
| FNV-1a | Fast | Good | Low |
| MurmurHash3 | Very Fast | Excellent | Very Low |
| xxHash | Fastest | Excellent | Very Low |
| CRC32 | Fast | Poor | Medium |
| MD5 | Slow | Excellent | Very Low |

**Recommendation**: FNV-1a for default (stdlib), xxHash as option for high-performance.

---

## Appendix C: Virtual Nodes Calculation

**Formula**: `vnodes = physical_nodes × virtual_nodes_per_node`

**Distribution variance** decreases with more virtual nodes:
- 10 vnodes: ~15% variance
- 50 vnodes: ~7% variance
- 150 vnodes: ~3% variance
- 500 vnodes: ~1% variance

**Trade-off**: More vnodes = better distribution but higher memory and lookup cost.

**Recommended**: 150 virtual nodes per physical node (matches Ketama).

---

## References

1. Skip Lists: A Probabilistic Alternative to Balanced Trees (Pugh, 1990)
2. Consistent Hashing and Random Trees (Karger et al., 1997)
3. Redis Sorted Sets Implementation
4. Cassandra Partitioning & Replication
5. libketama - Consistent Hashing Library
