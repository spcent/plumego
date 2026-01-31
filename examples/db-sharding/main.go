package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spcent/plumego/store/db"
	"github.com/spcent/plumego/store/db/rw"
	"github.com/spcent/plumego/store/db/sharding"
)

func main() {
	fmt.Println("Database Sharding Example")
	fmt.Println("=========================")

	// Example 1: Basic Sharding with Mod Strategy
	example1Basic()

	// Example 2: Sharding with Read-Write Splitting
	example2ReadWriteSplitting()

	// Example 3: Multiple Tables with Different Strategies
	example3MultipleStrategies()

	// Example 4: Cross-Shard Query Policies
	example4CrossShardPolicies()

	// Example 5: Health Monitoring
	example5HealthMonitoring()
}

// Example 1: Basic sharding with mod strategy
func example1Basic() {
	fmt.Println("\nExample 1: Basic Sharding")
	fmt.Println("-------------------------")

	config := sharding.ClusterConfig{
		Shards: []sharding.ShardConfig{
			{
				Primary: db.Config{
					Driver: "mysql",
					DSN:    "user:pass@tcp(localhost:3306)/shard0",
				},
			},
			{
				Primary: db.Config{
					Driver: "mysql",
					DSN:    "user:pass@tcp(localhost:3307)/shard1",
				},
			},
			{
				Primary: db.Config{
					Driver: "mysql",
					DSN:    "user:pass@tcp(localhost:3308)/shard2",
				},
			},
			{
				Primary: db.Config{
					Driver: "mysql",
					DSN:    "user:pass@tcp(localhost:3309)/shard3",
				},
			},
		},
		ShardingRules: []sharding.ShardingRuleConfig{
			{
				TableName:      "users",
				ShardKeyColumn: "user_id",
				Strategy:       "mod", // user_id % 4 determines the shard
			},
		},
		CrossShardPolicy: sharding.CrossShardDeny,
	}

	// Note: This will fail if MySQL is not running, but demonstrates the API
	cluster, err := sharding.New(config)
	if err != nil {
		fmt.Printf("Failed to create cluster (expected if MySQL not running): %v\n", err)
		return
	}
	defer cluster.Close()

	ctx := context.Background()

	// Insert a user (will go to shard based on user_id % 4)
	_, err = cluster.ExecContext(ctx,
		"INSERT INTO users (user_id, name, email) VALUES (?, ?, ?)",
		100, "Alice", "alice@example.com")
	if err != nil {
		log.Printf("Insert failed: %v", err)
	}

	// Query a user (will automatically route to correct shard)
	row := cluster.QueryRowContext(ctx,
		"SELECT name, email FROM users WHERE user_id = ?", 100)

	var name, email string
	if err := row.Scan(&name, &email); err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		fmt.Printf("Found user: %s <%s>\n", name, email)
	}
}

// Example 2: Sharding with read-write splitting
func example2ReadWriteSplitting() {
	fmt.Println("\nExample 2: Sharding + Read-Write Splitting")
	fmt.Println("-------------------------------------------")

	config := sharding.ClusterConfig{
		Shards: []sharding.ShardConfig{
			{
				Primary: db.Config{
					Driver: "mysql",
					DSN:    "user:pass@tcp(localhost:3306)/shard0",
				},
				Replicas: []db.Config{
					{
						Driver: "mysql",
						DSN:    "user:pass@tcp(localhost:3316)/shard0_replica1",
					},
					{
						Driver: "mysql",
						DSN:    "user:pass@tcp(localhost:3326)/shard0_replica2",
					},
				},
				HealthCheck: rw.HealthCheckConfig{
					Enabled:           true,
					Interval:          30 * time.Second,
					Timeout:           5 * time.Second,
					FailureThreshold:  3,
					RecoveryThreshold: 2,
				},
				FallbackToPrimary: true, // Fallback to primary if all replicas are down
			},
			// Additional shards...
		},
		ShardingRules: []sharding.ShardingRuleConfig{
			{
				TableName:      "users",
				ShardKeyColumn: "user_id",
				Strategy:       "hash", // Use hash strategy for better distribution
			},
		},
	}

	fmt.Printf("Configuration: %d shards with read replicas\n", len(config.Shards))
	fmt.Println("Writes go to primary, reads distributed across replicas")
}

// Example 3: Multiple tables with different strategies
func example3MultipleStrategies() {
	fmt.Println("\nExample 3: Multiple Tables with Different Strategies")
	fmt.Println("----------------------------------------------------")

	config := sharding.ClusterConfig{
		Shards: []sharding.ShardConfig{
			// 4 shards configuration...
		},
		ShardingRules: []sharding.ShardingRuleConfig{
			{
				// Users table: sharded by user_id using mod strategy
				TableName:      "users",
				ShardKeyColumn: "user_id",
				Strategy:       "mod",
			},
			{
				// Orders table: sharded by order_id using hash strategy
				TableName:      "orders",
				ShardKeyColumn: "order_id",
				Strategy:       "hash",
			},
			{
				// Events table: sharded by timestamp using range strategy
				TableName:      "events",
				ShardKeyColumn: "created_at",
				Strategy:       "range",
				StrategyConfig: sharding.StrategyConfig{
					Ranges: []sharding.RangeDefinition{
						{Start: 0, End: 1640995200, Shard: 0},          // 2022
						{Start: 1640995200, End: 1672531200, Shard: 1}, // 2023
						{Start: 1672531200, End: 1704067200, Shard: 2}, // 2024
						{Start: 1704067200, End: 1735689600, Shard: 3}, // 2025
					},
				},
			},
			{
				// Geo table: sharded by region using list strategy
				TableName:      "geo_data",
				ShardKeyColumn: "region",
				Strategy:       "list",
				StrategyConfig: sharding.StrategyConfig{
					Mapping: map[string]int{
						"US":   0,
						"EU":   1,
						"ASIA": 2,
					},
					DefaultShard: 3, // Default for unmapped regions
				},
			},
		},
	}

	fmt.Println("Configured 4 tables with different sharding strategies:")
	for _, rule := range config.ShardingRules {
		fmt.Printf("- %s: sharded by %s using %s strategy\n",
			rule.TableName, rule.ShardKeyColumn, rule.Strategy)
	}
}

// Example 4: Cross-shard query policies
func example4CrossShardPolicies() {
	fmt.Println("\nExample 4: Cross-Shard Query Policies")
	fmt.Println("--------------------------------------")

	// Policy 1: Deny (default, safest)
	configDeny := sharding.ClusterConfig{
		CrossShardPolicy: sharding.CrossShardDeny,
		// ... shard configuration
	}
	fmt.Println("Policy: CrossShardDeny")
	fmt.Println("  Queries without shard key will be rejected")
	fmt.Println("  Safest option, prevents accidental full table scans")

	// Policy 2: First (query first shard only)
	configFirst := sharding.ClusterConfig{
		CrossShardPolicy: sharding.CrossShardFirst,
		// ... shard configuration
	}
	fmt.Println("\nPolicy: CrossShardFirst")
	fmt.Println("  Queries without shard key go to first shard only")
	fmt.Println("  Useful for lookup tables or metadata")

	// Policy 3: All (query all shards)
	configAll := sharding.ClusterConfig{
		CrossShardPolicy: sharding.CrossShardAll,
		// ... shard configuration
	}
	fmt.Println("\nPolicy: CrossShardAll")
	fmt.Println("  Queries without shard key go to all shards concurrently")
	fmt.Println("  Use with caution, can be expensive")

	_ = configDeny
	_ = configFirst
	_ = configAll
}

// Example 5: Health monitoring
func example5HealthMonitoring() {
	fmt.Println("\nExample 5: Health Monitoring")
	fmt.Println("-----------------------------")

	config := sharding.ClusterConfig{
		Shards: []sharding.ShardConfig{
			// ... shard configuration
		},
		ShardingRules: []sharding.ShardingRuleConfig{
			{
				TableName:      "users",
				ShardKeyColumn: "user_id",
				Strategy:       "mod",
			},
		},
		EnableMetrics: true, // Enable metrics collection
	}

	cluster, err := sharding.New(config)
	if err != nil {
		fmt.Printf("Failed to create cluster: %v\n", err)
		return
	}
	defer cluster.Close()

	// Check health of all shards
	ctx := context.Background()
	health, err := cluster.HealthCheck(ctx, 5*time.Second)
	if err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Println("Shard Health:")
		for _, h := range health {
			fmt.Printf("  Shard %d: %s (latency: %v)\n",
				h.ShardIndex, h.Status, h.Latency)
		}
	}

	// Get routing metrics
	metrics := cluster.Metrics()
	fmt.Printf("\nRouting Metrics:\n")
	fmt.Printf("  Total Queries: %d\n", metrics.TotalQueries)
	fmt.Printf("  Single-Shard: %d\n", metrics.SingleShardQueries)
	fmt.Printf("  Cross-Shard: %d\n", metrics.CrossShardQueries)
	fmt.Printf("  Routing Errors: %d\n", metrics.RoutingErrors)
	fmt.Printf("  Per-Shard Counts: %v\n", metrics.ShardQueryCounts)
}
