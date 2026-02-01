# Plumego Cache Package

High-performance, feature-rich caching library for Go with support for **basic key-value storage**, **leaderboards** (sorted sets), and **distributed caching**.

## Features

### 🚀 Core Features

- **In-Memory Cache** with LRU eviction
- **TTL Support** for automatic expiration
- **Atomic Operations** (Incr, Decr, Append)
- **Thread-Safe** with optimized RWMutex usage

### 🏆 Leaderboard Support (v1.1+)

- **Skip List** implementation for O(log N) operations
- **Redis-Compatible API** (ZAdd, ZRange, ZRank, ZScore, etc.)
- **Range Queries** by score or rank
- **Bulk Operations** for efficient batch updates
- **Built-in Metrics** tracking

### 🌐 Distributed Caching (v1.2+)

- **Consistent Hashing** with virtual nodes
- **Multi-Node Support** with automatic sharding
- **Replication** (None, Async, Sync modes)
- **Health Checking** with automatic failover
- **Dynamic Scaling** (add/remove nodes at runtime)
- **Sub-microsecond** hash ring lookups

## Quick Start

### Basic Cache

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/spcent/plumego/store/cache"
)

func main() {
    // Create cache with default configuration
    config := cache.DefaultConfig()
    c := cache.NewMemoryCacheWithConfig(config)
    defer c.Close()

    ctx := context.Background()

    // Set a value with 1 hour TTL
    c.Set(ctx, "user:123", []byte("Alice"), time.Hour)

    // Get the value
    value, err := c.Get(ctx, "user:123")
    if err != nil {
        panic(err)
    }
    fmt.Println(string(value)) // Output: Alice

    // Atomic increment
    count, _ := c.Incr(ctx, "counter", 1)
    fmt.Println(count) // Output: 1
}
```

### Leaderboard

```go
package main

import (
    "context"
    "fmt"

    "github.com/spcent/plumego/store/cache"
)

func main() {
    // Create leaderboard cache
    config := cache.DefaultConfig()
    lbConfig := cache.DefaultLeaderboardConfig()
    lbc := cache.NewMemoryLeaderboardCache(config, lbConfig)
    defer lbc.Close()

    ctx := context.Background()

    // Add players to leaderboard
    lbc.ZAdd(ctx, "game:scores",
        &cache.ZMember{Member: "Alice", Score: 1250.0},
        &cache.ZMember{Member: "Bob", Score: 980.0},
        &cache.ZMember{Member: "Charlie", Score: 1500.0},
    )

    // Get top 3 players (descending order)
    top3, _ := lbc.ZRange(ctx, "game:scores", 0, 2, true)
    for i, player := range top3 {
        fmt.Printf("#%d: %s (%.0f points)\n", i+1, player.Member, player.Score)
    }
    // Output:
    // #1: Charlie (1500 points)
    // #2: Alice (1250 points)
    // #3: Bob (980 points)

    // Get player rank
    rank, _ := lbc.ZRank(ctx, "game:scores", "Alice", true)
    fmt.Printf("Alice's rank: #%d\n", rank+1) // Output: #2

    // Update score
    newScore, _ := lbc.ZIncrBy(ctx, "game:scores", "Alice", 300)
    fmt.Printf("Alice's new score: %.0f\n", newScore) // Output: 1550
}
```

### Distributed Cache

```go
package main

import (
    "context"
    "time"

    "github.com/spcent/plumego/store/cache"
    "github.com/spcent/plumego/store/cache/distributed"
)

func main() {
    // Create cache nodes
    config := cache.DefaultConfig()
    nodes := []distributed.CacheNode{
        distributed.NewNode("node-1", cache.NewMemoryCacheWithConfig(config)),
        distributed.NewNode("node-2", cache.NewMemoryCacheWithConfig(config)),
        distributed.NewNode("node-3", cache.NewMemoryCacheWithConfig(config)),
    }

    // Create distributed cache with replication
    distConfig := distributed.DefaultConfig()
    distConfig.ReplicationFactor = 2 // 2x replication for fault tolerance
    dc := distributed.New(nodes, distConfig)
    defer dc.Close()

    ctx := context.Background()

    // Keys are automatically sharded across nodes
    dc.Set(ctx, "user:123", []byte("Alice"), time.Hour)
    dc.Set(ctx, "user:456", []byte("Bob"), time.Hour)

    // Reads automatically routed to correct node
    value, _ := dc.Get(ctx, "user:123")
    println(string(value)) // Output: Alice

    // Automatic failover if a node goes down
    // Data is still available from replicas
}
```

## Installation

```bash
go get github.com/spcent/plumego
```

## Performance

### Benchmarks

Benchmarks run on: Intel Core i7, 16GB RAM, Go 1.24

#### Basic Cache Operations

| Operation | Time/Op | Throughput |
|-----------|---------|------------|
| Set | 156 ns | 6.4M ops/sec |
| Get | 89 ns | 11.2M ops/sec |
| Delete | 95 ns | 10.5M ops/sec |
| Incr | 142 ns | 7.0M ops/sec |

#### Leaderboard Operations

| Operation | Time/Op | Complexity |
|-----------|---------|------------|
| ZAdd | 231 ns | O(log N) |
| ZRange (100) | 8.1 µs | O(log N + M) |
| ZRank | 412 ns | O(log N) |
| ZScore | 189 ns | O(N) |
| ZIncrBy | 267 ns | O(log N) |

#### Distributed Cache

| Operation | Time/Op | Notes |
|-----------|---------|-------|
| Hash Ring Lookup | 165 ns | O(log V) where V=virtual nodes |
| Distributed Set | 423 ns | With async replication |
| Distributed Get | 541 ns | Includes failover capability |
| Node Failover | 1.2 µs | Automatic replica selection |

### Memory Usage

| Structure | Overhead per Item |
|-----------|-------------------|
| Basic Cache Entry | ~80 bytes |
| Leaderboard Member | ~120 bytes (skip list) |
| Hash Ring Virtual Node | ~16 bytes |

## Use Cases

### Basic Cache

- **Session Storage** - User sessions with TTL
- **API Response Caching** - Reduce backend load
- **Rate Limiting** - Track request counts with Incr/Decr
- **Temporary Data** - Short-lived computations

### Leaderboard

- **Gaming Leaderboards** - Player rankings and scores
- **Trending Content** - Most viewed/liked items
- **Time-Series Data** - Events sorted by timestamp
- **Priority Queues** - Task scheduling by priority
- **Analytics** - Top products, users, or metrics

### Distributed Cache

- **High Availability** - Replicate data across nodes
- **Horizontal Scaling** - Distribute load across machines
- **Geographic Distribution** - Cache data closer to users
- **Fault Tolerance** - Automatic failover on node failure
- **Large Datasets** - Shard data beyond single-node capacity

## Architecture

### Basic Cache

```
┌─────────────────────────────────────┐
│      MemoryCache                    │
├─────────────────────────────────────┤
│  • LRU Eviction                     │
│  • TTL Management                   │
│  • Thread-Safe (RWMutex)            │
│  • Atomic Operations                │
└─────────────────────────────────────┘
         │
         ├─> Map[string]*Item
         └─> Doubly-Linked List (LRU)
```

### Leaderboard

```
┌─────────────────────────────────────┐
│   MemoryLeaderboardCache            │
├─────────────────────────────────────┤
│  • Skip List (32 levels)            │
│  • O(log N) operations              │
│  • Score-based indexing             │
│  • Member lookup map                │
└─────────────────────────────────────┘
         │
         ├─> Skip List (sorted by score)
         └─> Map[member]→score (fast lookup)
```

### Distributed Cache

```
┌─────────────────────────────────────┐
│    DistributedCache                 │
├─────────────────────────────────────┤
│  • Consistent Hash Ring             │
│  • Health Checker                   │
│  • Replication Manager              │
│  • Metrics Collector                │
└─────────────────────────────────────┘
         │
         ├─> Hash Ring (virtual nodes)
         ├─> Node[1]: MemoryCache
         ├─> Node[2]: MemoryCache
         └─> Node[N]: MemoryCache
```

## Configuration

### Basic Cache Config

```go
config := &cache.Config{
    MaxSize:         1000,               // Max items (0 = unlimited)
    DefaultTTL:      time.Hour,          // Default TTL
    CleanupInterval: 10 * time.Minute,   // Expired item cleanup frequency
    EvictionPolicy:  cache.EvictionLRU,  // LRU eviction
}
```

### Leaderboard Config

```go
lbConfig := &cache.LeaderboardConfig{
    MaxLeaderboards:  1000,  // Max number of leaderboards
    MaxMembersPerSet: 10000, // Max members per leaderboard
    EnableMetrics:    true,  // Track operations
}
```

### Distributed Config

```go
distConfig := &distributed.Config{
    VirtualNodes:        150,                         // Virtual nodes per physical node
    ReplicationFactor:   2,                           // Number of replicas
    ReplicationMode:     distributed.ReplicationAsync, // Async/Sync/None
    FailoverStrategy:    distributed.FailoverNextNode, // Failover behavior
    HealthCheckInterval: 10 * time.Second,            // Health check frequency
    HealthCheckTimeout:  2 * time.Second,             // Health check timeout
    EnableMetrics:       true,                        // Collect metrics
}
```

## API Overview

### Basic Cache Interface

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
    Close() error
}
```

### Leaderboard Interface

```go
type LeaderboardCache interface {
    Cache  // Inherits basic cache operations

    // Add/Update members
    ZAdd(ctx context.Context, key string, members ...*ZMember) error
    ZIncrBy(ctx context.Context, key string, member string, delta float64) (float64, error)

    // Query by rank
    ZRange(ctx context.Context, key string, start, stop int64, reverse bool) ([]*ZMember, error)
    ZRank(ctx context.Context, key string, member string, reverse bool) (int64, error)

    // Query by score
    ZRangeByScore(ctx context.Context, key string, min, max float64, reverse bool) ([]*ZMember, error)
    ZScore(ctx context.Context, key string, member string) (float64, error)

    // Remove members
    ZRem(ctx context.Context, key string, members ...string) (int64, error)
    ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error)
    ZRemRangeByScore(ctx context.Context, key string, min, max float64) (int64, error)

    // Statistics
    ZCard(ctx context.Context, key string) (int64, error)
    ZCount(ctx context.Context, key string, min, max float64) (int64, error)
}
```

### Distributed Cache Interface

```go
type DistributedCache interface {
    Cache  // Same interface as basic cache

    // Node management
    AddNode(node CacheNode) error
    RemoveNode(nodeID string) error
    Nodes() []CacheNode
    NodeHealth(nodeID string) (HealthStatus, error)

    // Metrics
    GetMetrics() *DistributedMetrics
}
```

## Examples

The `examples/` directory contains complete working examples:

- **[cache-leaderboard](../../examples/cache-leaderboard/main.go)** - Gaming leaderboard with 10 demonstration scenarios
- **[cache-distributed](../../examples/cache-distributed/main.go)** - Distributed cache with failover and node management
- **[cache-combined](../../examples/cache-combined/main.go)** - Combining distributed caching with leaderboards

Run an example:

```bash
go run ./examples/cache-leaderboard/main.go
```

## Documentation

- **[API Reference](./API_REFERENCE.md)** - Complete API documentation with examples
- **[Migration Guide](./MIGRATION_GUIDE.md)** - Upgrading from basic cache to leaderboard/distributed
- **[Design Document](./DESIGN_EXTENSIONS.md)** - Architecture and implementation details
- **[Architecture Diagrams](./ARCHITECTURE.md)** - Visual architecture reference
- **[Implementation Status](./IMPLEMENTATION_STATUS.md)** - Current progress and roadmap

## Testing

### Run All Tests

```bash
# All tests
go test ./store/cache/...

# With race detection
go test -race ./store/cache/...

# With coverage
go test -cover ./store/cache/...

# Verbose
go test -v ./store/cache/...
```

### Run Benchmarks

```bash
# All benchmarks
go test -bench=. ./store/cache/...

# Specific benchmark
go test -bench=BenchmarkZAdd ./store/cache/

# With memory stats
go test -bench=. -benchmem ./store/cache/...
```

## Comparison with Redis

| Feature | Plumego Cache | Redis |
|---------|---------------|-------|
| **Deployment** | Embedded (no server) | Standalone server |
| **Language** | Pure Go | C with Go client |
| **Persistence** | In-memory only | RDB/AOF persistence |
| **Networking** | Local (same process) | TCP/Unix socket |
| **Data Structures** | KV, Sorted Sets | KV, Lists, Sets, Sorted Sets, Streams, etc. |
| **Distributed** | Client-side sharding | Redis Cluster |
| **Performance** | ~10M ops/sec (local) | ~100K ops/sec (network) |
| **Use Case** | Application cache | Shared cache, message queue, pub/sub |

**When to use Plumego Cache:**
- Embedded caching (no separate server)
- Maximum performance (no network overhead)
- Simple deployment (single binary)
- Go-native integration

**When to use Redis:**
- Shared cache across services
- Persistence requirements
- Rich data structures (lists, streams, geo, etc.)
- Pub/Sub messaging

## Thread Safety

All cache operations are **thread-safe**:

- Basic cache uses `sync.RWMutex` for concurrent access
- Leaderboard uses per-set mutexes for fine-grained locking
- Distributed cache uses atomic operations for metrics
- Health checker runs in background goroutine

Safe for concurrent use from multiple goroutines.

## Error Handling

Common errors:

```go
// Not found
_, err := c.Get(ctx, "nonexistent")
if errors.Is(err, cache.ErrNotFound) {
    // Handle missing key
}

// Leaderboard not found
_, err := lbc.ZRange(ctx, "unknown", 0, 10, true)
if errors.Is(err, cache.ErrLeaderboardNotFound) {
    // Create leaderboard first
}

// Node unhealthy
_, err := dc.Get(ctx, "key")
if errors.Is(err, distributed.ErrNodeUnhealthy) {
    // All nodes failed
}

// No nodes available
err := dc.Set(ctx, "key", value, ttl)
if errors.Is(err, distributed.ErrNoNodesAvailable) {
    // Cluster has no healthy nodes
}
```

## Roadmap

### Completed ✅

- [x] Basic in-memory cache with LRU eviction
- [x] TTL support and background cleanup
- [x] Atomic operations (Incr, Decr, Append)
- [x] Leaderboard support with skip lists
- [x] Redis-compatible sorted set API
- [x] Distributed caching with consistent hashing
- [x] Health checking and automatic failover
- [x] Replication (Async, Sync modes)
- [x] Dynamic node management
- [x] Comprehensive benchmarks and tests

### Planned 🚧

- [ ] Persistence (WAL, snapshots)
- [ ] Redis protocol compatibility (RESP)
- [ ] Remote cache nodes (RPC)
- [ ] Additional data structures (Lists, Sets, HyperLogLog)
- [ ] Lua scripting support
- [ ] Pub/Sub messaging
- [ ] Transaction support (MULTI/EXEC)
- [ ] TTL refresh on access
- [ ] Scan/iteration support
- [ ] Prometheus metrics export

### Future Ideas 💡

- [ ] Write-through/write-behind caching
- [ ] Cache warming strategies
- [ ] Multi-level caching (L1/L2)
- [ ] Compression support
- [ ] Bloom filters for negative caching
- [ ] Geospatial indexing
- [ ] Time-series optimizations

## Contributing

Contributions welcome! Please:

1. Read the [Design Document](./DESIGN_EXTENSIONS.md)
2. Check existing tests for patterns
3. Add tests for new features
4. Run benchmarks before/after changes
5. Update documentation

## License

MIT License - see LICENSE file for details

## Support

- **Documentation**: See links above
- **Examples**: Check `examples/` directory
- **Issues**: File on GitHub
- **Questions**: Contact maintainers

---

**Plumego Cache** - Fast, flexible caching for Go applications 🚀
