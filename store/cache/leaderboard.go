package cache

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// LeaderboardCache extends Cache with sorted set operations for leaderboards
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

// ZMember represents a member in a sorted set with its score
type ZMember struct {
	Member string  // Member name/identifier
	Score  float64 // Score value
}

// LeaderboardConfig configures the leaderboard cache behavior
type LeaderboardConfig struct {
	MaxLeaderboards  int           // Maximum number of leaderboards (default: 1000)
	MaxMembersPerSet int           // Maximum members per leaderboard (default: 10000)
	DefaultTTL       time.Duration // Default TTL for leaderboards (default: 1 hour)
	CleanupInterval  time.Duration // Cleanup interval for expired leaderboards (default: 5 minutes)
	EnableMetrics    bool          // Enable metrics collection (default: true)
}

// DefaultLeaderboardConfig returns the default configuration
func DefaultLeaderboardConfig() *LeaderboardConfig {
	return &LeaderboardConfig{
		MaxLeaderboards:  1000,
		MaxMembersPerSet: 10000,
		DefaultTTL:       time.Hour,
		CleanupInterval:  5 * time.Minute,
		EnableMetrics:    true,
	}
}

// Validate checks if the configuration is valid
func (c *LeaderboardConfig) Validate() error {
	if c.MaxLeaderboards <= 0 {
		return ErrInvalidConfig
	}
	if c.MaxMembersPerSet <= 0 {
		return ErrInvalidConfig
	}
	if c.CleanupInterval <= 0 {
		c.CleanupInterval = 5 * time.Minute
	}
	if c.DefaultTTL <= 0 {
		c.DefaultTTL = time.Hour
	}
	return nil
}

// Leaderboard errors
var (
	ErrLeaderboardNotFound = errors.New("cache: leaderboard not found")
	ErrMemberNotFound      = errors.New("cache: member not found in sorted set")
	ErrLeaderboardFull     = errors.New("cache: leaderboard member limit reached")
	ErrInvalidScore        = errors.New("cache: invalid score value")
	ErrInvalidRange        = errors.New("cache: invalid range parameters")
)

// sortedSet represents a sorted set with skip list and score map
type sortedSet struct {
	mu         sync.RWMutex
	scores     map[string]float64 // Member -> Score mapping for O(1) lookup
	skipList   *skipList          // Skip list for ordered operations
	expiration time.Time          // Expiration time
}

// newSortedSet creates a new sorted set
func newSortedSet(ttl time.Duration) *sortedSet {
	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	return &sortedSet{
		scores:     make(map[string]float64),
		skipList:   newSkipList(),
		expiration: expiration,
	}
}

// isExpired checks if the sorted set has expired
func (ss *sortedSet) isExpired() bool {
	if ss.expiration.IsZero() {
		return false
	}
	return time.Now().After(ss.expiration)
}

// MemoryLeaderboardCache is an in-memory implementation of LeaderboardCache
type MemoryLeaderboardCache struct {
	*MemoryCache // Embed base cache

	leaderboards sync.Map // key -> *sortedSet
	config       *LeaderboardConfig
	metrics      *LeaderboardMetrics
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// LeaderboardMetrics tracks leaderboard-specific metrics
type LeaderboardMetrics struct {
	mu                sync.RWMutex
	ZAdds             uint64
	ZRems             uint64
	ZRangeQueries     uint64
	ZScoreLookups     uint64
	ZRankCalculations uint64
	TotalLeaderboards int64
	TotalMembers      int64
}

// NewMemoryLeaderboardCache creates a new in-memory leaderboard cache
func NewMemoryLeaderboardCache(cacheConfig Config, lbConfig *LeaderboardConfig) *MemoryLeaderboardCache {
	if lbConfig == nil {
		lbConfig = DefaultLeaderboardConfig()
	}

	if err := lbConfig.Validate(); err != nil {
		panic(fmt.Sprintf("invalid leaderboard config: %v", err))
	}

	baseCache := NewMemoryCacheWithConfig(cacheConfig)

	lbc := &MemoryLeaderboardCache{
		MemoryCache: baseCache,
		config:      lbConfig,
		metrics:     &LeaderboardMetrics{},
		stopChan:    make(chan struct{}),
	}

	// Start cleanup goroutine
	lbc.wg.Add(1)
	go lbc.cleanupLoop()

	return lbc
}

// Close stops the leaderboard cache and cleanup goroutine
func (lbc *MemoryLeaderboardCache) Close() error {
	close(lbc.stopChan)
	lbc.wg.Wait()
	return lbc.MemoryCache.Close()
}

// cleanupLoop periodically removes expired leaderboards
func (lbc *MemoryLeaderboardCache) cleanupLoop() {
	defer lbc.wg.Done()

	ticker := time.NewTicker(lbc.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lbc.cleanupExpiredLeaderboards()
		case <-lbc.stopChan:
			return
		}
	}
}

// cleanupExpiredLeaderboards removes expired leaderboards
func (lbc *MemoryLeaderboardCache) cleanupExpiredLeaderboards() {
	now := time.Now()
	count := 0

	lbc.leaderboards.Range(func(key, value any) bool {
		ss := value.(*sortedSet)
		if !ss.expiration.IsZero() && now.After(ss.expiration) {
			lbc.leaderboards.Delete(key)
			count++
		}
		return true
	})
}

// getOrCreateSortedSet retrieves or creates a sorted set
func (lbc *MemoryLeaderboardCache) getOrCreateSortedSet(key string, ttl time.Duration) (*sortedSet, error) {
	// Try to load existing
	if value, ok := lbc.leaderboards.Load(key); ok {
		ss := value.(*sortedSet)
		if !ss.isExpired() {
			return ss, nil
		}
		// Expired, delete it
		lbc.leaderboards.Delete(key)
	}

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = lbc.config.DefaultTTL
	}

	// Check max leaderboards limit
	count := int64(0)
	lbc.leaderboards.Range(func(_, _ any) bool {
		count++
		return true
	})

	if count >= int64(lbc.config.MaxLeaderboards) {
		return nil, ErrLeaderboardFull
	}

	// Create new sorted set
	ss := newSortedSet(ttl)
	lbc.leaderboards.Store(key, ss)

	return ss, nil
}

// getSortedSet retrieves a sorted set (read-only)
func (lbc *MemoryLeaderboardCache) getSortedSet(key string) (*sortedSet, error) {
	value, ok := lbc.leaderboards.Load(key)
	if !ok {
		return nil, ErrLeaderboardNotFound
	}

	ss := value.(*sortedSet)
	if ss.isExpired() {
		lbc.leaderboards.Delete(key)
		return nil, ErrLeaderboardNotFound
	}

	return ss, nil
}

// validateScore checks if a score is valid (not NaN or Inf)
func validateScore(score float64) error {
	if math.IsNaN(score) || math.IsInf(score, 0) {
		return ErrInvalidScore
	}
	return nil
}

// ZAdd adds or updates members in a sorted set
func (lbc *MemoryLeaderboardCache) ZAdd(ctx context.Context, key string, members ...*ZMember) error {
	if len(members) == 0 {
		return nil
	}

	// Validate scores
	for _, m := range members {
		if err := validateScore(m.Score); err != nil {
			return err
		}
	}

	ss, err := lbc.getOrCreateSortedSet(key, 0)
	if err != nil {
		return err
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Check member limit
	newMembers := 0
	for _, m := range members {
		if _, exists := ss.scores[m.Member]; !exists {
			newMembers++
		}
	}

	if int64(len(ss.scores)+newMembers) > int64(lbc.config.MaxMembersPerSet) {
		return ErrLeaderboardFull
	}

	// Add/update members
	for _, m := range members {
		// If member exists with different score, remove from skip list first
		if oldScore, exists := ss.scores[m.Member]; exists && oldScore != m.Score {
			ss.skipList.delete(m.Member, oldScore)
		}

		// Update score map
		ss.scores[m.Member] = m.Score

		// Insert into skip list (will replace if exists)
		ss.skipList.insert(m.Member, m.Score)
	}

	// Update metrics
	if lbc.config.EnableMetrics {
		lbc.metrics.mu.Lock()
		lbc.metrics.ZAdds += uint64(len(members))
		lbc.metrics.mu.Unlock()
	}

	return nil
}

// ZRem removes members from a sorted set
func (lbc *MemoryLeaderboardCache) ZRem(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}

	ss, err := lbc.getSortedSet(key)
	if err != nil {
		return err
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	for _, member := range members {
		if score, exists := ss.scores[member]; exists {
			ss.skipList.delete(member, score)
			delete(ss.scores, member)
		}
	}

	// Update metrics
	if lbc.config.EnableMetrics {
		lbc.metrics.mu.Lock()
		lbc.metrics.ZRems += uint64(len(members))
		lbc.metrics.mu.Unlock()
	}

	return nil
}

// ZScore returns the score of a member
func (lbc *MemoryLeaderboardCache) ZScore(ctx context.Context, key string, member string) (float64, error) {
	ss, err := lbc.getSortedSet(key)
	if err != nil {
		return 0, err
	}

	ss.mu.RLock()
	defer ss.mu.RUnlock()

	score, exists := ss.scores[member]
	if !exists {
		return 0, ErrMemberNotFound
	}

	// Update metrics
	if lbc.config.EnableMetrics {
		lbc.metrics.mu.Lock()
		lbc.metrics.ZScoreLookups++
		lbc.metrics.mu.Unlock()
	}

	return score, nil
}

// ZIncrBy increments the score of a member by delta
func (lbc *MemoryLeaderboardCache) ZIncrBy(ctx context.Context, key string, member string, delta float64) (float64, error) {
	if err := validateScore(delta); err != nil {
		return 0, err
	}

	ss, err := lbc.getOrCreateSortedSet(key, 0)
	if err != nil {
		return 0, err
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Get current score or start from 0
	oldScore := float64(0)
	if score, exists := ss.scores[member]; exists {
		oldScore = score
		// Remove from skip list
		ss.skipList.delete(member, oldScore)
	} else {
		// Check member limit for new member
		if int64(len(ss.scores)+1) > int64(lbc.config.MaxMembersPerSet) {
			return 0, ErrLeaderboardFull
		}
	}

	// Calculate new score
	newScore := oldScore + delta
	if err := validateScore(newScore); err != nil {
		// Restore old score if new score is invalid
		if oldScore != 0 {
			ss.skipList.insert(member, oldScore)
		}
		return 0, err
	}

	// Update score map
	ss.scores[member] = newScore

	// Insert into skip list with new score
	ss.skipList.insert(member, newScore)

	return newScore, nil
}

// ZRange returns members in the specified rank range
func (lbc *MemoryLeaderboardCache) ZRange(ctx context.Context, key string, start, stop int64, desc bool) ([]*ZMember, error) {
	ss, err := lbc.getSortedSet(key)
	if err != nil {
		return nil, err
	}

	ss.mu.RLock()
	defer ss.mu.RUnlock()

	nodes := ss.skipList.getRange(start, stop, desc)
	if nodes == nil {
		return []*ZMember{}, nil
	}

	results := make([]*ZMember, len(nodes))
	for i, node := range nodes {
		results[i] = &ZMember{
			Member: node.member,
			Score:  node.score,
		}
	}

	// Update metrics
	if lbc.config.EnableMetrics {
		lbc.metrics.mu.Lock()
		lbc.metrics.ZRangeQueries++
		lbc.metrics.mu.Unlock()
	}

	return results, nil
}

// ZRangeByScore returns members with scores in the specified range
func (lbc *MemoryLeaderboardCache) ZRangeByScore(ctx context.Context, key string, min, max float64, desc bool) ([]*ZMember, error) {
	if err := validateScore(min); err != nil {
		return nil, err
	}
	if err := validateScore(max); err != nil {
		return nil, err
	}

	ss, err := lbc.getSortedSet(key)
	if err != nil {
		return nil, err
	}

	ss.mu.RLock()
	defer ss.mu.RUnlock()

	nodes := ss.skipList.getRangeByScore(min, max, desc)
	if nodes == nil {
		return []*ZMember{}, nil
	}

	results := make([]*ZMember, len(nodes))
	for i, node := range nodes {
		results[i] = &ZMember{
			Member: node.member,
			Score:  node.score,
		}
	}

	// Update metrics
	if lbc.config.EnableMetrics {
		lbc.metrics.mu.Lock()
		lbc.metrics.ZRangeQueries++
		lbc.metrics.mu.Unlock()
	}

	return results, nil
}

// ZRank returns the rank of a member (0-based)
func (lbc *MemoryLeaderboardCache) ZRank(ctx context.Context, key string, member string, desc bool) (int64, error) {
	ss, err := lbc.getSortedSet(key)
	if err != nil {
		return -1, err
	}

	ss.mu.RLock()
	defer ss.mu.RUnlock()

	score, exists := ss.scores[member]
	if !exists {
		return -1, ErrMemberNotFound
	}

	rank := ss.skipList.getRank(member, score, desc)
	if rank < 0 {
		return -1, ErrMemberNotFound
	}

	// Update metrics
	if lbc.config.EnableMetrics {
		lbc.metrics.mu.Lock()
		lbc.metrics.ZRankCalculations++
		lbc.metrics.mu.Unlock()
	}

	return rank, nil
}

// ZCard returns the number of members in a sorted set
func (lbc *MemoryLeaderboardCache) ZCard(ctx context.Context, key string) (int64, error) {
	ss, err := lbc.getSortedSet(key)
	if err != nil {
		if errors.Is(err, ErrLeaderboardNotFound) {
			return 0, nil
		}
		return 0, err
	}

	ss.mu.RLock()
	defer ss.mu.RUnlock()

	return ss.skipList.length, nil
}

// ZCount returns the number of members with scores in the specified range
func (lbc *MemoryLeaderboardCache) ZCount(ctx context.Context, key string, min, max float64) (int64, error) {
	if err := validateScore(min); err != nil {
		return 0, err
	}
	if err := validateScore(max); err != nil {
		return 0, err
	}

	ss, err := lbc.getSortedSet(key)
	if err != nil {
		if errors.Is(err, ErrLeaderboardNotFound) {
			return 0, nil
		}
		return 0, err
	}

	ss.mu.RLock()
	defer ss.mu.RUnlock()

	return ss.skipList.count(min, max), nil
}

// ZRemRangeByRank removes members in the specified rank range
func (lbc *MemoryLeaderboardCache) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	ss, err := lbc.getSortedSet(key)
	if err != nil {
		if errors.Is(err, ErrLeaderboardNotFound) {
			return 0, nil
		}
		return 0, err
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Get nodes to delete
	nodes := ss.skipList.getRange(start, stop, false)
	if nodes == nil {
		return 0, nil
	}

	// Remove from skip list and score map
	for _, node := range nodes {
		ss.skipList.delete(node.member, node.score)
		delete(ss.scores, node.member)
	}

	// Update metrics
	if lbc.config.EnableMetrics {
		lbc.metrics.mu.Lock()
		lbc.metrics.ZRems += uint64(len(nodes))
		lbc.metrics.mu.Unlock()
	}

	return int64(len(nodes)), nil
}

// ZRemRangeByScore removes members with scores in the specified range
func (lbc *MemoryLeaderboardCache) ZRemRangeByScore(ctx context.Context, key string, min, max float64) (int64, error) {
	if err := validateScore(min); err != nil {
		return 0, err
	}
	if err := validateScore(max); err != nil {
		return 0, err
	}

	ss, err := lbc.getSortedSet(key)
	if err != nil {
		if errors.Is(err, ErrLeaderboardNotFound) {
			return 0, nil
		}
		return 0, err
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Get nodes to delete
	nodes := ss.skipList.getRangeByScore(min, max, false)
	if nodes == nil {
		return 0, nil
	}

	// Remove from skip list and score map
	for _, node := range nodes {
		ss.skipList.delete(node.member, node.score)
		delete(ss.scores, node.member)
	}

	// Update metrics
	if lbc.config.EnableMetrics {
		lbc.metrics.mu.Lock()
		lbc.metrics.ZRems += uint64(len(nodes))
		lbc.metrics.mu.Unlock()
	}

	return int64(len(nodes)), nil
}

// GetLeaderboardMetrics returns a snapshot of leaderboard metrics
func (lbc *MemoryLeaderboardCache) GetLeaderboardMetrics() *LeaderboardMetrics {
	lbc.metrics.mu.RLock()
	defer lbc.metrics.mu.RUnlock()

	// Count current leaderboards and members
	totalLeaderboards := int64(0)
	totalMembers := int64(0)

	lbc.leaderboards.Range(func(key, value any) bool {
		ss := value.(*sortedSet)
		ss.mu.RLock()
		totalMembers += ss.skipList.length
		ss.mu.RUnlock()
		totalLeaderboards++
		return true
	})

	return &LeaderboardMetrics{
		ZAdds:             lbc.metrics.ZAdds,
		ZRems:             lbc.metrics.ZRems,
		ZRangeQueries:     lbc.metrics.ZRangeQueries,
		ZScoreLookups:     lbc.metrics.ZScoreLookups,
		ZRankCalculations: lbc.metrics.ZRankCalculations,
		TotalLeaderboards: totalLeaderboards,
		TotalMembers:      totalMembers,
	}
}
