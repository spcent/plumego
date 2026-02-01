package cache

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
)

func TestSkipListBasicOperations(t *testing.T) {
	sl := newSkipList()

	// Test insert
	sl.insert("player1", 100.0)
	sl.insert("player2", 95.0)
	sl.insert("player3", 90.0)

	if sl.length != 3 {
		t.Errorf("expected length 3, got %d", sl.length)
	}

	// Test getScore
	score, ok := sl.getScore("player1")
	if !ok || score != 100.0 {
		t.Errorf("expected score 100.0, got %f, ok=%v", score, ok)
	}

	// Test delete
	if !sl.delete("player2", 95.0) {
		t.Error("failed to delete player2")
	}

	if sl.length != 2 {
		t.Errorf("expected length 2 after delete, got %d", sl.length)
	}

	// Verify deleted member is gone
	_, ok = sl.getScore("player2")
	if ok {
		t.Error("player2 should not exist after deletion")
	}
}

func TestSkipListRangeQueries(t *testing.T) {
	sl := newSkipList()

	// Insert members
	members := []struct {
		name  string
		score float64
	}{
		{"player1", 100.0},
		{"player2", 95.0},
		{"player3", 90.0},
		{"player4", 85.0},
		{"player5", 80.0},
	}

	for _, m := range members {
		sl.insert(m.name, m.score)
	}

	// Test ascending range
	nodes := sl.getRange(0, 2, false)
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	expected := []string{"player5", "player4", "player3"}
	for i, node := range nodes {
		if node.member != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], node.member)
		}
	}

	// Test descending range
	nodes = sl.getRange(0, 2, true)
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	expected = []string{"player1", "player2", "player3"}
	for i, node := range nodes {
		if node.member != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], node.member)
		}
	}
}

func TestSkipListRangeByScore(t *testing.T) {
	sl := newSkipList()

	sl.insert("player1", 100.0)
	sl.insert("player2", 95.0)
	sl.insert("player3", 90.0)
	sl.insert("player4", 85.0)
	sl.insert("player5", 80.0)

	// Test score range
	nodes := sl.getRangeByScore(85.0, 95.0, false)
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	expected := []string{"player4", "player3", "player2"}
	for i, node := range nodes {
		if node.member != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], node.member)
		}
	}
}

func TestSkipListGetRank(t *testing.T) {
	sl := newSkipList()

	sl.insert("player1", 100.0)
	sl.insert("player2", 95.0)
	sl.insert("player3", 90.0)

	// Test ascending rank (lowest score = rank 0)
	rank := sl.getRank("player3", 90.0, false)
	if rank != 0 {
		t.Errorf("expected rank 0 for player3, got %d", rank)
	}

	rank = sl.getRank("player1", 100.0, false)
	if rank != 2 {
		t.Errorf("expected rank 2 for player1, got %d", rank)
	}

	// Test descending rank (highest score = rank 0)
	rank = sl.getRank("player1", 100.0, true)
	if rank != 0 {
		t.Errorf("expected rank 0 for player1 (desc), got %d", rank)
	}
}

func TestSkipListCount(t *testing.T) {
	sl := newSkipList()

	sl.insert("player1", 100.0)
	sl.insert("player2", 95.0)
	sl.insert("player3", 90.0)
	sl.insert("player4", 85.0)
	sl.insert("player5", 80.0)

	count := sl.count(85.0, 95.0)
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}

	count = sl.count(100.0, 100.0)
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	count = sl.count(200.0, 300.0)
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestLeaderboardCacheBasicOperations(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Test ZAdd
	err := lbc.ZAdd(ctx, "game:scores", &ZMember{Member: "player1", Score: 100.0})
	if err != nil {
		t.Errorf("ZAdd failed: %v", err)
	}

	// Test ZScore
	score, err := lbc.ZScore(ctx, "game:scores", "player1")
	if err != nil {
		t.Errorf("ZScore failed: %v", err)
	}
	if score != 100.0 {
		t.Errorf("expected score 100.0, got %f", score)
	}

	// Test ZCard
	card, err := lbc.ZCard(ctx, "game:scores")
	if err != nil {
		t.Errorf("ZCard failed: %v", err)
	}
	if card != 1 {
		t.Errorf("expected cardinality 1, got %d", card)
	}

	// Test ZRem
	err = lbc.ZRem(ctx, "game:scores", "player1")
	if err != nil {
		t.Errorf("ZRem failed: %v", err)
	}

	card, err = lbc.ZCard(ctx, "game:scores")
	if err != nil {
		t.Errorf("ZCard failed: %v", err)
	}
	if card != 0 {
		t.Errorf("expected cardinality 0 after ZRem, got %d", card)
	}
}

func TestLeaderboardCacheZIncrBy(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Increment non-existent member (should create with delta)
	score, err := lbc.ZIncrBy(ctx, "game:scores", "player1", 10.0)
	if err != nil {
		t.Errorf("ZIncrBy failed: %v", err)
	}
	if score != 10.0 {
		t.Errorf("expected score 10.0, got %f", score)
	}

	// Increment existing member
	score, err = lbc.ZIncrBy(ctx, "game:scores", "player1", 5.0)
	if err != nil {
		t.Errorf("ZIncrBy failed: %v", err)
	}
	if score != 15.0 {
		t.Errorf("expected score 15.0, got %f", score)
	}

	// Decrement (negative delta)
	score, err = lbc.ZIncrBy(ctx, "game:scores", "player1", -3.0)
	if err != nil {
		t.Errorf("ZIncrBy failed: %v", err)
	}
	if score != 12.0 {
		t.Errorf("expected score 12.0, got %f", score)
	}
}

func TestLeaderboardCacheZRange(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Add multiple members
	members := []*ZMember{
		{Member: "player1", Score: 100.0},
		{Member: "player2", Score: 95.0},
		{Member: "player3", Score: 90.0},
		{Member: "player4", Score: 85.0},
		{Member: "player5", Score: 80.0},
	}

	err := lbc.ZAdd(ctx, "game:scores", members...)
	if err != nil {
		t.Fatalf("ZAdd failed: %v", err)
	}

	// Get top 3 (descending order)
	results, err := lbc.ZRange(ctx, "game:scores", 0, 2, true)
	if err != nil {
		t.Errorf("ZRange failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	expected := []string{"player1", "player2", "player3"}
	for i, r := range results {
		if r.Member != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], r.Member)
		}
	}

	// Get bottom 2 (ascending order)
	results, err = lbc.ZRange(ctx, "game:scores", 0, 1, false)
	if err != nil {
		t.Errorf("ZRange failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	expected = []string{"player5", "player4"}
	for i, r := range results {
		if r.Member != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], r.Member)
		}
	}
}

func TestLeaderboardCacheZRangeByScore(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Add multiple members
	members := []*ZMember{
		{Member: "player1", Score: 100.0},
		{Member: "player2", Score: 95.0},
		{Member: "player3", Score: 90.0},
		{Member: "player4", Score: 85.0},
		{Member: "player5", Score: 80.0},
	}

	err := lbc.ZAdd(ctx, "game:scores", members...)
	if err != nil {
		t.Fatalf("ZAdd failed: %v", err)
	}

	// Get members with score between 85 and 95
	results, err := lbc.ZRangeByScore(ctx, "game:scores", 85.0, 95.0, false)
	if err != nil {
		t.Errorf("ZRangeByScore failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	expected := []string{"player4", "player3", "player2"}
	for i, r := range results {
		if r.Member != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], r.Member)
		}
	}
}

func TestLeaderboardCacheZRank(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Add multiple members
	members := []*ZMember{
		{Member: "player1", Score: 100.0},
		{Member: "player2", Score: 95.0},
		{Member: "player3", Score: 90.0},
	}

	err := lbc.ZAdd(ctx, "game:scores", members...)
	if err != nil {
		t.Fatalf("ZAdd failed: %v", err)
	}

	// Get descending rank (0 = highest score)
	rank, err := lbc.ZRank(ctx, "game:scores", "player1", true)
	if err != nil {
		t.Errorf("ZRank failed: %v", err)
	}
	if rank != 0 {
		t.Errorf("expected rank 0 for player1, got %d", rank)
	}

	rank, err = lbc.ZRank(ctx, "game:scores", "player2", true)
	if err != nil {
		t.Errorf("ZRank failed: %v", err)
	}
	if rank != 1 {
		t.Errorf("expected rank 1 for player2, got %d", rank)
	}

	// Test non-existent member
	_, err = lbc.ZRank(ctx, "game:scores", "nonexistent", true)
	if err == nil {
		t.Error("expected error for non-existent member")
	}
}

func TestLeaderboardCacheZCount(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Add multiple members
	members := []*ZMember{
		{Member: "player1", Score: 100.0},
		{Member: "player2", Score: 95.0},
		{Member: "player3", Score: 90.0},
		{Member: "player4", Score: 85.0},
		{Member: "player5", Score: 80.0},
	}

	err := lbc.ZAdd(ctx, "game:scores", members...)
	if err != nil {
		t.Fatalf("ZAdd failed: %v", err)
	}

	// Count members in range
	count, err := lbc.ZCount(ctx, "game:scores", 85.0, 95.0)
	if err != nil {
		t.Errorf("ZCount failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}

	// Count all
	count, err = lbc.ZCount(ctx, "game:scores", 0, 200.0)
	if err != nil {
		t.Errorf("ZCount failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestLeaderboardCacheZRemRange(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Add multiple members
	members := []*ZMember{
		{Member: "player1", Score: 100.0},
		{Member: "player2", Score: 95.0},
		{Member: "player3", Score: 90.0},
		{Member: "player4", Score: 85.0},
		{Member: "player5", Score: 80.0},
	}

	err := lbc.ZAdd(ctx, "game:scores", members...)
	if err != nil {
		t.Fatalf("ZAdd failed: %v", err)
	}

	// Remove by rank
	removed, err := lbc.ZRemRangeByRank(ctx, "game:scores", 0, 1)
	if err != nil {
		t.Errorf("ZRemRangeByRank failed: %v", err)
	}
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}

	// Verify remaining count
	card, err := lbc.ZCard(ctx, "game:scores")
	if err != nil {
		t.Errorf("ZCard failed: %v", err)
	}
	if card != 3 {
		t.Errorf("expected cardinality 3, got %d", card)
	}
}

func TestLeaderboardCacheZRemRangeByScore(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Add multiple members
	members := []*ZMember{
		{Member: "player1", Score: 100.0},
		{Member: "player2", Score: 95.0},
		{Member: "player3", Score: 90.0},
		{Member: "player4", Score: 85.0},
		{Member: "player5", Score: 80.0},
	}

	err := lbc.ZAdd(ctx, "game:scores", members...)
	if err != nil {
		t.Fatalf("ZAdd failed: %v", err)
	}

	// Remove by score range
	removed, err := lbc.ZRemRangeByScore(ctx, "game:scores", 80.0, 90.0)
	if err != nil {
		t.Errorf("ZRemRangeByScore failed: %v", err)
	}
	if removed != 3 {
		t.Errorf("expected 3 removed, got %d", removed)
	}

	// Verify remaining members
	card, err := lbc.ZCard(ctx, "game:scores")
	if err != nil {
		t.Errorf("ZCard failed: %v", err)
	}
	if card != 2 {
		t.Errorf("expected cardinality 2, got %d", card)
	}
}

func TestLeaderboardCacheTTL(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()
	lbConfig.DefaultTTL = 100 * time.Millisecond
	lbConfig.CleanupInterval = 50 * time.Millisecond

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Add member
	err := lbc.ZAdd(ctx, "game:scores", &ZMember{Member: "player1", Score: 100.0})
	if err != nil {
		t.Fatalf("ZAdd failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Try to access expired leaderboard
	_, err = lbc.ZScore(ctx, "game:scores", "player1")
	if err == nil {
		t.Error("expected error for expired leaderboard")
	}
}

func TestLeaderboardCacheInvalidScore(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Test NaN
	err := lbc.ZAdd(ctx, "game:scores", &ZMember{Member: "player1", Score: math.NaN()})
	if err != ErrInvalidScore {
		t.Errorf("expected ErrInvalidScore for NaN, got %v", err)
	}

	// Test Inf
	err = lbc.ZAdd(ctx, "game:scores", &ZMember{Member: "player1", Score: math.Inf(1)})
	if err != ErrInvalidScore {
		t.Errorf("expected ErrInvalidScore for Inf, got %v", err)
	}

	// Test -Inf
	err = lbc.ZAdd(ctx, "game:scores", &ZMember{Member: "player1", Score: math.Inf(-1)})
	if err != ErrInvalidScore {
		t.Errorf("expected ErrInvalidScore for -Inf, got %v", err)
	}
}

func TestLeaderboardCacheMemberLimit(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()
	lbConfig.MaxMembersPerSet = 5

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Add up to limit
	var err error
	for i := 0; i < 5; i++ {
		err = lbc.ZAdd(ctx, "game:scores", &ZMember{
			Member: fmt.Sprintf("player%d", i),
			Score:  float64(i),
		})
		if err != nil {
			t.Fatalf("ZAdd failed at %d: %v", i, err)
		}
	}

	// Try to add one more (should fail)
	err = lbc.ZAdd(ctx, "game:scores", &ZMember{Member: "player5", Score: 5.0})
	if err != ErrLeaderboardFull {
		t.Errorf("expected ErrLeaderboardFull, got %v", err)
	}
}

func TestLeaderboardCacheConcurrency(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()
	const numGoroutines = 100
	const numOpsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOpsPerGoroutine; j++ {
				member := fmt.Sprintf("player%d", id)

				// ZIncrBy
				_, _ = lbc.ZIncrBy(ctx, "game:scores", member, 1.0)

				// ZScore
				_, _ = lbc.ZScore(ctx, "game:scores", member)

				// ZRank
				_, _ = lbc.ZRank(ctx, "game:scores", member, true)
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	card, err := lbc.ZCard(ctx, "game:scores")
	if err != nil {
		t.Errorf("ZCard failed: %v", err)
	}

	if card != numGoroutines {
		t.Errorf("expected cardinality %d, got %d", numGoroutines, card)
	}

	// Verify each player has correct score
	for i := 0; i < numGoroutines; i++ {
		member := fmt.Sprintf("player%d", i)
		score, err := lbc.ZScore(ctx, "game:scores", member)
		if err != nil {
			t.Errorf("ZScore failed for %s: %v", member, err)
		}
		if score != float64(numOpsPerGoroutine) {
			t.Errorf("expected score %d for %s, got %f", numOpsPerGoroutine, member, score)
		}
	}
}

func TestLeaderboardCacheMetrics(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Perform operations
	_ = lbc.ZAdd(ctx, "game:scores", &ZMember{Member: "player1", Score: 100.0})
	_, _ = lbc.ZScore(ctx, "game:scores", "player1")
	_, _ = lbc.ZRank(ctx, "game:scores", "player1", true)
	_, _ = lbc.ZRange(ctx, "game:scores", 0, 10, true)
	_ = lbc.ZRem(ctx, "game:scores", "player1")

	// Get metrics
	metrics := lbc.GetLeaderboardMetrics()

	if metrics.ZAdds != 1 {
		t.Errorf("expected 1 ZAdd, got %d", metrics.ZAdds)
	}
	if metrics.ZScoreLookups != 1 {
		t.Errorf("expected 1 ZScoreLookup, got %d", metrics.ZScoreLookups)
	}
	if metrics.ZRankCalculations != 1 {
		t.Errorf("expected 1 ZRankCalculation, got %d", metrics.ZRankCalculations)
	}
	if metrics.ZRangeQueries != 1 {
		t.Errorf("expected 1 ZRangeQuery, got %d", metrics.ZRangeQueries)
	}
	if metrics.ZRems != 1 {
		t.Errorf("expected 1 ZRem, got %d", metrics.ZRems)
	}
}

func TestLeaderboardCacheUpdateScore(t *testing.T) {
	config := DefaultConfig()
	lbConfig := DefaultLeaderboardConfig()

	lbc := NewMemoryLeaderboardCache(config, lbConfig)
	defer lbc.Close()

	ctx := context.Background()

	// Add initial score
	err := lbc.ZAdd(ctx, "game:scores", &ZMember{Member: "player1", Score: 100.0})
	if err != nil {
		t.Fatalf("ZAdd failed: %v", err)
	}

	// Update with different score
	err = lbc.ZAdd(ctx, "game:scores", &ZMember{Member: "player1", Score: 150.0})
	if err != nil {
		t.Fatalf("ZAdd update failed: %v", err)
	}

	// Verify updated score
	score, err := lbc.ZScore(ctx, "game:scores", "player1")
	if err != nil {
		t.Errorf("ZScore failed: %v", err)
	}
	if score != 150.0 {
		t.Errorf("expected updated score 150.0, got %f", score)
	}

	// Verify cardinality is still 1
	card, err := lbc.ZCard(ctx, "game:scores")
	if err != nil {
		t.Errorf("ZCard failed: %v", err)
	}
	if card != 1 {
		t.Errorf("expected cardinality 1, got %d", card)
	}
}
