package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spcent/plumego/store/cache"
	"github.com/spcent/plumego/store/cache/distributed"
)

func main() {
	fmt.Println("=== Distributed Cache Example ===\n")

	// Create 3 cache nodes
	fmt.Println("1. Setting up 3 cache nodes...")
	nodes := []distributed.CacheNode{
		distributed.NewNode("node-us-east", cache.NewMemoryCache()),
		distributed.NewNode("node-us-west", cache.NewMemoryCache()),
		distributed.NewNode("node-eu", cache.NewMemoryCache()),
	}
	fmt.Println("   ✓ Created node-us-east")
	fmt.Println("   ✓ Created node-us-west")
	fmt.Println("   ✓ Created node-eu")
	fmt.Println()

	// Create distributed cache with replication
	fmt.Println("2. Creating distributed cache with replication...")
	config := distributed.DefaultConfig()
	config.ReplicationFactor = 2 // Each key on 2 nodes
	config.ReplicationMode = distributed.ReplicationAsync
	config.VirtualNodes = 150

	dc := distributed.New(nodes, config)
	defer dc.Close()

	fmt.Printf("   ✓ Virtual nodes per physical node: %d\n", config.VirtualNodes)
	fmt.Printf("   ✓ Replication factor: %d\n", config.ReplicationFactor)
	fmt.Printf("   ✓ Replication mode: Async\n")
	fmt.Println()

	ctx := context.Background()

	// Example 3: Store user sessions
	fmt.Println("3. Storing user sessions...")
	sessions := map[string]string{
		"session:user123": "Alice's session data",
		"session:user456": "Bob's session data",
		"session:user789": "Charlie's session data",
		"session:user101": "Diana's session data",
		"session:user202": "Eve's session data",
	}

	for key, value := range sessions {
		err := dc.Set(ctx, key, []byte(value), 30*time.Minute)
		if err != nil {
			log.Fatalf("Failed to set %s: %v", key, err)
		}
		fmt.Printf("   ✓ Stored %s\n", key)
	}
	fmt.Println()

	// Example 4: Retrieve sessions (automatic routing)
	fmt.Println("4. Retrieving sessions (automatic sharding)...")
	for key := range sessions {
		value, err := dc.Get(ctx, key)
		if err != nil {
			log.Printf("Failed to get %s: %v", key, err)
			continue
		}
		fmt.Printf("   ✓ Retrieved %s: %s\n", key, string(value))
	}
	fmt.Println()

	// Example 5: Show key distribution
	fmt.Println("5. Key Distribution Across Nodes:")
	metrics := dc.GetMetrics()
	fmt.Printf("   Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("   Healthy nodes: %d\n", metrics.HealthyNodes)
	fmt.Printf("   Unhealthy nodes: %d\n", metrics.UnhealthyNodes)
	fmt.Println()

	// Example 6: Add a new node (dynamic scaling)
	fmt.Println("6. Adding a new node (dynamic scaling)...")
	newNode := distributed.NewNode("node-asia", cache.NewMemoryCache())
	err := dc.AddNode(newNode)
	if err != nil {
		log.Fatalf("Failed to add node: %v", err)
	}
	fmt.Println("   ✓ Added node-asia")
	fmt.Printf("   ✓ Total nodes: %d\n", len(dc.Nodes()))

	// Give health checker time to register
	time.Sleep(100 * time.Millisecond)

	metrics = dc.GetMetrics()
	fmt.Printf("   ✓ Rebalance events: %d\n", metrics.RebalanceEvents)
	fmt.Println()

	// Example 7: Simulate node failure and failover
	fmt.Println("7. Simulating node failure...")

	// Mark first node as unhealthy
	nodes[0].UpdateHealth(distributed.HealthStatusUnhealthy)
	fmt.Printf("   ✗ Marked %s as unhealthy\n", nodes[0].ID())

	// Try to get a key that might be on the failed node
	// Failover should automatically use replica
	value, err := dc.Get(ctx, "session:user123")
	if err != nil {
		log.Printf("Failover failed: %v", err)
	} else {
		fmt.Printf("   ✓ Failover successful! Retrieved: %s\n", string(value))
	}

	metrics = dc.GetMetrics()
	fmt.Printf("   ✓ Failover count: %d\n", metrics.FailoverCount)
	fmt.Println()

	// Example 8: Restore node health
	fmt.Println("8. Restoring node health...")
	nodes[0].UpdateHealth(distributed.HealthStatusHealthy)
	fmt.Printf("   ✓ Marked %s as healthy\n", nodes[0].ID())

	time.Sleep(100 * time.Millisecond)
	metrics = dc.GetMetrics()
	fmt.Printf("   ✓ Healthy nodes: %d\n", metrics.HealthyNodes)
	fmt.Println()

	// Example 9: Atomic operations
	fmt.Println("9. Atomic counter operations...")

	// Initialize counter
	dc.Set(ctx, "counter:page-views", []byte("0"), time.Hour)

	// Increment multiple times
	for i := 0; i < 5; i++ {
		count, err := dc.Incr(ctx, "counter:page-views", 1)
		if err != nil {
			log.Printf("Incr failed: %v", err)
			continue
		}
		fmt.Printf("   Page views: %d\n", count)
	}
	fmt.Println()

	// Example 10: Node status
	fmt.Println("10. Node Health Status:")
	for _, node := range dc.Nodes() {
		status := node.HealthStatus()
		statusStr := "✓"
		if status == distributed.HealthStatusUnhealthy {
			statusStr = "✗"
		}
		fmt.Printf("    %s %-15s: %s\n", statusStr, node.ID(), status)
	}
	fmt.Println()

	// Example 11: Remove a node
	fmt.Println("11. Removing a node...")
	err = dc.RemoveNode("node-asia")
	if err != nil {
		log.Printf("Failed to remove node: %v", err)
	} else {
		fmt.Printf("   ✓ Removed node-asia\n")
		fmt.Printf("   ✓ Remaining nodes: %d\n", len(dc.Nodes()))
	}
	fmt.Println()

	// Example 12: Final metrics
	fmt.Println("12. Final Cache Metrics:")
	finalMetrics := dc.GetMetrics()
	fmt.Printf("   Total requests: %d\n", finalMetrics.TotalRequests)
	fmt.Printf("   Failover count: %d\n", finalMetrics.FailoverCount)
	fmt.Printf("   Rebalance events: %d\n", finalMetrics.RebalanceEvents)
	fmt.Printf("   Healthy nodes: %d\n", finalMetrics.HealthyNodes)
	fmt.Printf("   Unhealthy nodes: %d\n", finalMetrics.UnhealthyNodes)

	fmt.Println("\n=== Example Complete ===")
}
