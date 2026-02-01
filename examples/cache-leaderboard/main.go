package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spcent/plumego/store/cache"
)

func main() {
	fmt.Println("=== Gaming Leaderboard Example ===")
	fmt.Println()

	// Create leaderboard cache
	config := cache.DefaultConfig()
	lbConfig := cache.DefaultLeaderboardConfig()
	lbConfig.MaxLeaderboards = 10
	lbConfig.MaxMembersPerSet = 1000

	lbc := cache.NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Example 1: Add players to leaderboard
	fmt.Println("1. Adding players to leaderboard...")
	players := []*cache.ZMember{
		{Member: "Alice", Score: 1250.0},
		{Member: "Bob", Score: 980.0},
		{Member: "Charlie", Score: 1500.0},
		{Member: "Diana", Score: 1100.0},
		{Member: "Eve", Score: 890.0},
		{Member: "Frank", Score: 1350.0},
		{Member: "Grace", Score: 1200.0},
		{Member: "Henry", Score: 950.0},
	}

	err := lbc.ZAdd(ctx, "game:weekly:scores", players...)
	if err != nil {
		log.Fatalf("Failed to add players: %v", err)
	}
	fmt.Printf("Added %d players to leaderboard\n\n", len(players))

	// Example 2: Get top 5 players
	fmt.Println("2. Top 5 Players:")
	top5, err := lbc.ZRange(ctx, "game:weekly:scores", 0, 4, true)
	if err != nil {
		log.Fatalf("Failed to get top 5: %v", err)
	}

	for i, player := range top5 {
		fmt.Printf("   #%d: %-10s - %.0f points\n", i+1, player.Member, player.Score)
	}
	fmt.Println()

	// Example 3: Update player score
	fmt.Println("3. Alice wins a match (+150 points)...")
	newScore, err := lbc.ZIncrBy(ctx, "game:weekly:scores", "Alice", 150.0)
	if err != nil {
		log.Fatalf("Failed to increment score: %v", err)
	}
	fmt.Printf("   Alice's new score: %.0f\n\n", newScore)

	// Example 4: Check updated rankings
	fmt.Println("4. Updated Top 5:")
	top5, _ = lbc.ZRange(ctx, "game:weekly:scores", 0, 4, true)
	for i, player := range top5 {
		fmt.Printf("   #%d: %-10s - %.0f points\n", i+1, player.Member, player.Score)
	}
	fmt.Println()

	// Example 5: Get player rank
	fmt.Println("5. Player Rankings:")
	for _, playerName := range []string{"Charlie", "Alice", "Bob"} {
		rank, err := lbc.ZRank(ctx, "game:weekly:scores", playerName, true)
		if err != nil {
			log.Printf("Failed to get rank for %s: %v", playerName, err)
			continue
		}
		score, _ := lbc.ZScore(ctx, "game:weekly:scores", playerName)
		fmt.Printf("   %-10s: Rank #%d (%.0f points)\n", playerName, rank+1, score)
	}
	fmt.Println()

	// Example 6: Get players in score range
	fmt.Println("6. Players with 1000-1400 points:")
	midRange, err := lbc.ZRangeByScore(ctx, "game:weekly:scores", 1000.0, 1400.0, true)
	if err != nil {
		log.Fatalf("Failed to get range: %v", err)
	}

	for _, player := range midRange {
		fmt.Printf("   %-10s - %.0f points\n", player.Member, player.Score)
	}
	fmt.Println()

	// Example 7: Total players and statistics
	fmt.Println("7. Leaderboard Statistics:")
	total, _ := lbc.ZCard(ctx, "game:weekly:scores")
	fmt.Printf("   Total players: %d\n", total)

	count1000Plus, _ := lbc.ZCount(ctx, "game:weekly:scores", 1000.0, 10000.0)
	fmt.Printf("   Players with 1000+ points: %d\n", count1000Plus)

	// Get metrics
	metrics := lbc.GetLeaderboardMetrics()
	fmt.Printf("   Total ZAdd operations: %d\n", metrics.ZAdds)
	fmt.Printf("   Total ZRange queries: %d\n", metrics.ZRangeQueries)
	fmt.Printf("   Total ZRank calculations: %d\n", metrics.ZRankCalculations)
	fmt.Println()

	// Example 8: Multiple leaderboards
	fmt.Println("8. Multiple Game Modes:")

	// Add players to different game modes
	casualPlayers := []*cache.ZMember{
		{Member: "Alice", Score: 500.0},
		{Member: "Bob", Score: 750.0},
		{Member: "Charlie", Score: 650.0},
	}
	lbc.ZAdd(ctx, "game:casual:scores", casualPlayers...)

	rankedPlayers := []*cache.ZMember{
		{Member: "Alice", Score: 2100.0},
		{Member: "Bob", Score: 1850.0},
		{Member: "Charlie", Score: 2500.0},
	}
	lbc.ZAdd(ctx, "game:ranked:scores", rankedPlayers...)

	// Show top player in each mode
	casualTop, _ := lbc.ZRange(ctx, "game:casual:scores", 0, 0, true)
	rankedTop, _ := lbc.ZRange(ctx, "game:ranked:scores", 0, 0, true)

	fmt.Printf("   Casual Mode Champion: %s (%.0f points)\n", casualTop[0].Member, casualTop[0].Score)
	fmt.Printf("   Ranked Mode Champion: %s (%.0f points)\n", rankedTop[0].Member, rankedTop[0].Score)
	fmt.Println()

	// Example 9: Remove inactive players
	fmt.Println("9. Removing players with < 1000 points...")
	removed, err := lbc.ZRemRangeByScore(ctx, "game:weekly:scores", 0, 999.0)
	if err != nil {
		log.Fatalf("Failed to remove players: %v", err)
	}
	fmt.Printf("   Removed %d inactive players\n\n", removed)

	// Example 10: Final leaderboard
	fmt.Println("10. Final Leaderboard (Active Players Only):")
	allActive, _ := lbc.ZRange(ctx, "game:weekly:scores", 0, -1, true)
	for i, player := range allActive {
		fmt.Printf("    #%d: %-10s - %.0f points\n", i+1, player.Member, player.Score)
	}

	fmt.Println("\n=== Example Complete ===")
}
