# Database Sharding Example

This example demonstrates how to use the database sharding functionality in Plumego.

## Features

- **Multiple Sharding Strategies**: Hash, Mod, Range, and List strategies
- **Read-Write Splitting**: Each shard can have multiple read replicas
- **Cross-Shard Query Policies**: Control how queries without shard keys are handled
- **Health Monitoring**: Built-in health checks and metrics
- **Transparent Interface**: Implements standard `db.DB` interface

## Quick Start

```go
package main

import (
	"context"
	"github.com/spcent/plumego/store/db"
	"github.com/spcent/plumego/store/db/sharding"
)

func main() {
	// Configure sharded cluster
	config := sharding.ClusterConfig{
		Shards: []sharding.ShardConfig{
			{
				Primary: db.Config{
					Driver: "mysql",
					DSN:    "user:pass@tcp(localhost:3306)/shard0",
				},
			},
			// More shards...
		},
		ShardingRules: []sharding.ShardingRuleConfig{
			{
				TableName:      "users",
				ShardKeyColumn: "user_id",
				Strategy:       "mod",
			},
		},
	}

	// Create cluster
	cluster, err := sharding.New(config)
	if err != nil {
		panic(err)
	}
	defer cluster.Close()

	// Use it like a normal database
	ctx := context.Background()
	_, err = cluster.ExecContext(ctx,
		"INSERT INTO users (user_id, name) VALUES (?, ?)",
		100, "Alice")
}
```

## Sharding Strategies

### 1. Mod Strategy

Simple modulo-based sharding. Best for integer keys.

```go
{
	TableName:      "users",
	ShardKeyColumn: "user_id",
	Strategy:       "mod",
}
// User ID 100 -> Shard 0 (100 % 4 = 0)
// User ID 101 -> Shard 1 (101 % 4 = 1)
```

### 2. Hash Strategy

Hash-based sharding. Better distribution for non-integer keys.

```go
{
	TableName:      "users",
	ShardKeyColumn: "username",
	Strategy:       "hash",
}
// Hashes the key and distributes evenly
```

### 3. Range Strategy

Range-based sharding. Good for time-series data.

```go
{
	TableName:      "events",
	ShardKeyColumn: "created_at",
	Strategy:       "range",
	StrategyConfig: sharding.StrategyConfig{
		Ranges: []sharding.RangeDefinition{
			{Start: 0, End: 1000, ShardIndex: 0},
			{Start: 1000, End: 2000, ShardIndex: 1},
		},
	},
}
```

### 4. List Strategy

Explicit mapping of values to shards. Good for geographic data.

```go
{
	TableName:      "geo_data",
	ShardKeyColumn: "region",
	Strategy:       "list",
	StrategyConfig: sharding.StrategyConfig{
		Mapping: map[string]int{
			"US":   0,
			"EU":   1,
			"ASIA": 2,
		},
		DefaultShard: 3,
	},
}
```

## Read-Write Splitting

Each shard can have multiple read replicas:

```go
{
	Primary: db.Config{
		Driver: "mysql",
		DSN:    "user:pass@tcp(localhost:3306)/shard0",
	},
	Replicas: []db.Config{
		{
			Driver: "mysql",
			DSN:    "user:pass@tcp(localhost:3316)/shard0_replica",
		},
	},
	HealthCheck: rw.HealthCheckConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
	},
	FallbackToPrimary: true,
}
```

- Writes always go to primary
- Reads distributed across replicas
- Automatic failover to primary if replicas fail

## Cross-Shard Query Policies

Control how queries without shard keys are handled:

### CrossShardDeny (Default)

```go
CrossShardPolicy: sharding.CrossShardDeny
```
- Safest option
- Queries without shard key are rejected
- Prevents accidental full table scans

### CrossShardFirst

```go
CrossShardPolicy: sharding.CrossShardFirst
```
- Query only the first shard
- Useful for lookup tables

### CrossShardAll

```go
CrossShardPolicy: sharding.CrossShardAll
```
- Query all shards concurrently
- Returns first successful result
- Use with caution (expensive)

## Transactions

For sharded databases, transactions must be shard-specific:

```go
// Begin transaction on specific shard
tx, err := cluster.BeginTxOnShard(ctx, shardIndex, nil)
if err != nil {
	return err
}
defer tx.Rollback()

// Execute queries in transaction
_, err = tx.ExecContext(ctx, "INSERT INTO users ...")
if err != nil {
	return err
}

// Commit
return tx.Commit()
```

## Health Monitoring

```go
// Check health of all shards
health, err := cluster.HealthCheck(ctx, 5*time.Second)
for _, h := range health {
	fmt.Printf("Shard %d: %s (latency: %v)\n",
		h.ShardIndex, h.Status, h.Latency)
}

// Get routing metrics
metrics := cluster.Metrics()
fmt.Printf("Total Queries: %d\n", metrics.TotalQueries)
fmt.Printf("Routing Errors: %d\n", metrics.RoutingErrors)
```

## Performance

- Routing overhead: < 100μs per query
- Supports concurrent queries across shards
- Minimal memory allocation

## Running the Example

```bash
# Run the example
go run main.go

# Note: Requires MySQL to be running on specified ports
# See docker-compose.yml for test environment setup
```

## Best Practices

1. **Choose the Right Strategy**
   - Use `mod` for evenly distributed integer keys
   - Use `hash` for non-integer keys or better distribution
   - Use `range` for time-series or sequential data
   - Use `list` for explicit mappings (e.g., regions)

2. **Handle Shard Keys Properly**
   - Always include shard key in WHERE clauses
   - Avoid cross-shard queries when possible
   - Use transactions only within a single shard

3. **Monitor Health**
   - Enable metrics to track query distribution
   - Set up health checks for replicas
   - Monitor routing errors

4. **Plan for Growth**
   - Start with power-of-2 shard counts (2, 4, 8, 16)
   - Use consistent hashing for easier rebalancing
   - Plan migration strategy before scaling

## Limitations

- Cross-shard transactions are not supported
- Cross-shard JOINs require application-level logic
- Rebalancing requires manual data migration
- Some complex SQL queries may not be supported

## See Also

- [Database Sharding ROADMAP](../../store/db/ROADMAP.md)
- [Read-Write Splitting](../../store/db/rw/)
- [Sharding Strategies](../../store/db/sharding/)
