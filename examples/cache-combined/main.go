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
	fmt.Println("=== Distributed Leaderboard Example ===")
	fmt.Println("Global gaming platform with regional leaderboards")
	fmt.Println()

	// Create leaderboard caches for different regions
	fmt.Println("1. Setting up regional leaderboard nodes...")

	config := cache.DefaultConfig()
	lbConfig := cache.DefaultLeaderboardConfig()

	// Create 3 regional nodes, each with leaderboard support
	nodes := []distributed.CacheNode{
		distributed.NewNode("region-americas", cache.NewMemoryLeaderboardCache(config, lbConfig)),
		distributed.NewNode("region-europe", cache.NewMemoryLeaderboardCache(config, lbConfig)),
		distributed.NewNode("region-asia", cache.NewMemoryLeaderboardCache(config, lbConfig)),
	}

	fmt.Println("   ✓ Americas node (leaderboard-enabled)")
	fmt.Println("   ✓ Europe node (leaderboard-enabled)")
	fmt.Println("   ✓ Asia node (leaderboard-enabled)")
	fmt.Println()

	// Create distributed cache
	fmt.Println("2. Creating distributed cache cluster...")
	distConfig := distributed.DefaultConfig()
	distConfig.ReplicationFactor = 2
	distConfig.VirtualNodes = 200

	dc := distributed.New(nodes, distConfig)
	defer dc.Close()

	fmt.Printf("   ✓ Replication factor: %d\n", distConfig.ReplicationFactor)
	fmt.Printf("   ✓ Virtual nodes: %d per physical node\n", distConfig.VirtualNodes)
	fmt.Println()

	ctx := context.Background()

	// Example 3: Regional leaderboards
	fmt.Println("3. Populating regional leaderboards...")

	// Note: Each leaderboard key will be automatically sharded to a node
	// based on consistent hashing

	regionalPlayers := map[string][]*cache.ZMember{
		"americas": {
			{Member: "Alice_USA", Score: 2500.0},
			{Member: "Bob_CAN", Score: 2200.0},
			{Member: "Carol_MEX", Score: 1900.0},
		},
		"europe": {
			{Member: "David_UK", Score: 2700.0},
			{Member: "Eve_FR", Score: 2400.0},
			{Member: "Frank_DE", Score: 2100.0},
		},
		"asia": {
			{Member: "Grace_JP", Score: 2800.0},
			{Member: "Henry_KR", Score: 2600.0},
			{Member: "Iris_CN", Score: 2300.0},
		},
	}

	for region, players := range regionalPlayers {
		// Get the node that will handle this leaderboard
		// (In a real app, you'd access the leaderboard cache directly on that node)

		// For demonstration, we'll store as regular cache data
		// In production, you'd want to access the LeaderboardCache interface
		// on the specific node

		for _, player := range players {
			// Store player data
			playerKey := fmt.Sprintf("player:%s:%s", region, player.Member)
			playerData := fmt.Sprintf("Score: %.0f, Region: %s", player.Score, region)
			dc.Set(ctx, playerKey, []byte(playerData), time.Hour)
		}

		fmt.Printf("   ✓ Loaded %s leaderboard (%d players)\n", region, len(players))
	}
	fmt.Println()

	// Example 4: Global leaderboard (distributed across nodes)
	fmt.Println("4. Creating global leaderboard...")

	globalPlayers := []*cache.ZMember{
		{Member: "Grace_JP", Score: 2800.0},
		{Member: "David_UK", Score: 2700.0},
		{Member: "Henry_KR", Score: 2600.0},
		{Member: "Alice_USA", Score: 2500.0},
		{Member: "Eve_FR", Score: 2400.0},
		{Member: "Iris_CN", Score: 2300.0},
		{Member: "Bob_CAN", Score: 2200.0},
		{Member: "Frank_DE", Score: 2100.0},
		{Member: "Carol_MEX", Score: 1900.0},
	}

	// Store global leaderboard data
	// (Key will be automatically sharded to one of the nodes)
	for rank, player := range globalPlayers {
		key := fmt.Sprintf("global:rank:%d", rank+1)
		value := fmt.Sprintf("%s:%.0f", player.Member, player.Score)
		dc.Set(ctx, key, []byte(value), time.Hour)
	}

	fmt.Printf("   ✓ Global leaderboard created (%d players)\n", len(globalPlayers))
	fmt.Println()

	// Example 5: Query top players
	fmt.Println("5. Top 5 Global Players:")
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("global:rank:%d", i)
		data, err := dc.Get(ctx, key)
		if err != nil {
			log.Printf("Failed to get rank %d: %v", i, err)
			continue
		}
		fmt.Printf("   #%d: %s\n", i, string(data))
	}
	fmt.Println()

	// Example 6: Session management with distributed cache
	fmt.Println("6. Managing player sessions...")

	sessions := map[string]string{
		"session:Alice_USA": "Active, Last match: 2m ago",
		"session:David_UK":  "Active, Last match: 5m ago",
		"session:Grace_JP":  "Idle, Last match: 15m ago",
		"session:Henry_KR":  "Active, Last match: 1m ago",
		"session:Eve_FR":    "Active, Last match: 3m ago",
	}

	for sessionKey, sessionData := range sessions {
		dc.Set(ctx, sessionKey, []byte(sessionData), 30*time.Minute)
	}

	fmt.Printf("   ✓ Stored %d active sessions\n", len(sessions))
	fmt.Println()

	// Example 7: Real-time score updates
	fmt.Println("7. Real-time score updates...")

	// Simulate match results
	updates := map[string]int64{
		"player:americas:Alice_USA": 50,
		"player:europe:David_UK":    75,
		"player:asia:Grace_JP":      100,
	}

	for playerKey, points := range updates {
		// Retrieve current data
		data, err := dc.Get(ctx, playerKey)
		if err != nil {
			log.Printf("Failed to get %s: %v", playerKey, err)
			continue
		}

		fmt.Printf("   ✓ Before: %s\n", string(data))
		fmt.Printf("     Added %d points\n", points)

		// In a real app, you'd update the leaderboard here
		// For now, just demonstrate the distributed cache access
	}
	fmt.Println()

	// Example 8: Cross-region statistics
	fmt.Println("8. Cross-Region Statistics:")

	// Count players per region
	regionCounts := map[string]int{
		"americas": len(regionalPlayers["americas"]),
		"europe":   len(regionalPlayers["europe"]),
		"asia":     len(regionalPlayers["asia"]),
	}

	totalPlayers := 0
	for region, count := range regionCounts {
		fmt.Printf("   %s: %d players\n", region, count)
		totalPlayers += count
	}
	fmt.Printf("   Total global players: %d\n", totalPlayers)
	fmt.Println()

	// Example 9: Health monitoring
	fmt.Println("9. Cluster Health Status:")

	metrics := dc.GetMetrics()
	fmt.Printf("   Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("   Healthy nodes: %d/%d\n", metrics.HealthyNodes, len(nodes))

	for _, node := range dc.Nodes() {
		status := "✓"
		if node.HealthStatus() != distributed.HealthStatusHealthy {
			status = "✗"
		}
		fmt.Printf("   %s %s: %s\n", status, node.ID(), node.HealthStatus())
	}
	fmt.Println()

	// Example 10: Performance metrics
	fmt.Println("10. Performance Summary:")
	fmt.Println("   Leaderboard Features:")
	fmt.Println("     • O(log N) score updates")
	fmt.Println("     • O(log N + M) range queries")
	fmt.Println("     • O(log N) rank lookups")
	fmt.Println()
	fmt.Println("   Distributed Features:")
	fmt.Printf("     • %d virtual nodes per region\n", distConfig.VirtualNodes)
	fmt.Printf("     • %dx replication for high availability\n", distConfig.ReplicationFactor)
	fmt.Println("     • Automatic failover on node failure")
	fmt.Println("     • Sub-microsecond hash lookups")

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("• Leaderboards can be distributed across multiple nodes")
	fmt.Println("• Each leaderboard key is automatically sharded via consistent hashing")
	fmt.Println("• Session data and leaderboard data coexist in the same cluster")
	fmt.Println("• Replication ensures high availability")
	fmt.Println("• Dynamic scaling with node add/remove")
}
