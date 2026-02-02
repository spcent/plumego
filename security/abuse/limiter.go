// Package abuse provides rate limiting and anti-abuse protection.
//
// This package implements a high-performance token bucket rate limiter with:
//   - Per-key rate limiting (e.g., per IP, per user)
//   - Configurable burst capacity and refill rate
//   - Automatic cleanup of idle entries
//   - Sharded internal storage for reduced lock contention
//   - Memory-efficient design with configurable limits
//
// The token bucket algorithm allows controlled burst traffic while maintaining
// a steady rate over time, making it ideal for API rate limiting.
//
// Example usage:
//
//	import "github.com/spcent/plumego/security/abuse"
//
//	// Create limiter: 100 requests/sec with burst of 200
//	limiter := abuse.NewGuard(abuse.Config{
//		Rate:     100,  // tokens per second
//		Capacity: 200,  // burst capacity
//	})
//
//	// Check if request is allowed
//	clientIP := "192.168.1.1"
//	if !limiter.Allow(clientIP) {
//		// Rate limit exceeded
//		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
//		return
//	}
//
//	// Process request...
package abuse

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultRate            = 100
	defaultCapacity        = 200
	defaultCleanupInterval = time.Minute
	defaultMaxIdle         = 5 * time.Minute
	defaultMaxEntries      = 100000
	defaultShards          = 16
)

// Config controls the per-key limiter behavior.
//
// Config configures a token bucket rate limiter for each key (e.g., client IP).
// The token bucket algorithm allows burst traffic up to the capacity limit,
// then throttles requests to the specified rate.
//
// Example:
//
//	import "github.com/spcent/plumego/security/abuse"
//
//	config := abuse.Config{
//		Rate:            10.0,      // 10 requests per second
//		Capacity:        100,       // Burst capacity of 100 requests
//		CleanupInterval: time.Minute, // Clean up idle entries every minute
//		MaxIdle:         5 * time.Minute, // Remove entries idle for 5 minutes
//	}
//	limiter := abuse.NewLimiter(config)
//
// The limiter automatically cleans up idle entries to prevent memory leaks.
type Config struct {
	// Rate is the number of requests per second allowed per key
	Rate float64

	// Capacity is the maximum burst size (number of tokens in the bucket)
	Capacity int

	// MaxEntries is the maximum number of tracked keys before eviction kicks in
	MaxEntries int

	// CleanupInterval is how often to clean up idle entries
	CleanupInterval time.Duration

	// MaxIdle is the maximum time an entry can be idle before being removed
	MaxIdle time.Duration

	// Shards is the number of lock shards for buckets
	Shards int

	// Now is a function that returns the current time (for testing)
	Now func() time.Time
}

// Decision captures the outcome of a limiter check.
//
// Example:
//
//	import "github.com/spcent/plumego/security/abuse"
//
//	limiter := abuse.NewLimiter(abuse.DefaultConfig())
//	decision := limiter.Allow("192.168.1.1")
//	if !decision.Allowed {
//		fmt.Printf("Rate limited. Retry after: %v\n", decision.RetryAfter)
//	}
type Decision struct {
	// Allowed indicates whether the request is allowed
	Allowed bool

	// Limit is the maximum number of requests allowed
	Limit int

	// Remaining is the number of requests remaining in the current window
	Remaining int

	// Reset is the time when the rate limit resets
	Reset time.Time

	// RetryAfter is the duration to wait before retrying (when not allowed)
	RetryAfter time.Duration
}

// Limiter enforces a token bucket per key.
//
// Limiter provides per-key rate limiting using a token bucket algorithm.
// Each key (e.g., client IP) has its own token bucket that refills at the
// specified rate up to the capacity limit.
//
// Example:
//
//	import "github.com/spcent/plumego/security/abuse"
//
//	config := abuse.Config{
//		Rate:     10.0,  // 10 requests per second
//		Capacity: 100,   // Burst capacity of 100
//	}
//	limiter := abuse.NewLimiter(config)
//	defer limiter.Stop()
//
//	// Check if request is allowed
//	decision := limiter.Allow("192.168.1.1")
//	if !decision.Allowed {
//		// Reject request
//	}
//
// The limiter automatically cleans up idle entries to prevent memory leaks.
// Call Stop() when done to clean up resources.
var bucketPool = sync.Pool{
	New: func() any {
		return &bucket{}
	},
}

type shardCounter struct {
	counters []atomic.Int64
}

func (sc *shardCounter) Add(shardIdx int, delta int64) {
	sc.counters[shardIdx].Add(delta)
}

func (sc *shardCounter) Total() int64 {
	var total int64
	for i := range sc.counters {
		total += sc.counters[i].Load()
	}
	return total
}

func (sc *shardCounter) Load() int64 {
	return sc.Total()
}

type LimiterMetrics struct {
	Allowed   atomic.Int64
	Rejected  atomic.Int64
	Evictions atomic.Int64
	Buckets   atomic.Int64
}

type Limiter struct {
	shards          []limiterShard
	rate            float64
	capacity        float64
	cleanupInterval time.Duration
	maxIdle         time.Duration
	maxEntries      int
	now             func() time.Time
	stopCh          chan struct{}
	stoppedCh       chan struct{}
	startOnce       sync.Once
	stopOnce        sync.Once
	bucketCount     shardCounter
	// Time caching for performance
	timeCache    time.Time
	timeCacheMu  sync.Mutex
	timeCacheTTL time.Duration
	// Metrics for monitoring
	metrics LimiterMetrics
}

func (l *Limiter) getBucket() *bucket {
	b := bucketPool.Get().(*bucket)
	now := l.now()
	b.tokens = l.capacity
	b.lastRefill = now
	b.lastAccess = now
	return b
}

func (l *Limiter) putBucket(b *bucket) {
	bucketPool.Put(b)
}

// nowWithCache returns the current time with caching for performance
func (l *Limiter) nowWithCache() time.Time {
	l.timeCacheMu.Lock()
	defer l.timeCacheMu.Unlock()

	// Initialize cache TTL if not set
	if l.timeCacheTTL == 0 {
		l.timeCacheTTL = time.Millisecond * 1 // Cache for 1ms
	}

	// Check if cache is still valid
	if !l.timeCache.IsZero() && time.Since(l.timeCache) < l.timeCacheTTL {
		return l.timeCache
	}

	// Update cache
	l.timeCache = l.now()
	return l.timeCache
}

type limiterShard struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
	lastAccess time.Time
}

// DefaultConfig returns baseline limiter settings.
func DefaultConfig() Config {
	return Config{
		Rate:            defaultRate,
		Capacity:        defaultCapacity,
		MaxEntries:      defaultMaxEntries,
		CleanupInterval: defaultCleanupInterval,
		MaxIdle:         defaultMaxIdle,
		Shards:          defaultShards,
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.Rate <= 0 {
		return fmt.Errorf("rate must be positive, got %f", c.Rate)
	}
	if c.Capacity <= 0 {
		return fmt.Errorf("capacity must be positive, got %d", c.Capacity)
	}
	if c.MaxEntries <= 0 {
		return fmt.Errorf("max_entries must be positive, got %d", c.MaxEntries)
	}
	if c.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup_interval must be positive, got %v", c.CleanupInterval)
	}
	if c.MaxIdle <= 0 {
		return fmt.Errorf("max_idle must be positive, got %v", c.MaxIdle)
	}
	if c.Shards <= 0 {
		return fmt.Errorf("shards must be positive, got %d", c.Shards)
	}
	return nil
}

// NewLimiter creates a limiter with the provided configuration.
//
// Example:
//
//	import "github.com/spcent/plumego/security/abuse"
//
//	config := abuse.Config{
//		Rate:     10.0,  // 10 requests per second
//		Capacity: 100,   // Burst capacity of 100
//	}
//	limiter := abuse.NewLimiter(config)
//	defer limiter.Stop()
func NewLimiter(config Config) *Limiter {
	defaults := DefaultConfig()

	if config.Rate <= 0 {
		config.Rate = defaults.Rate
	}
	if config.Capacity <= 0 {
		config.Capacity = defaults.Capacity
	}
	if config.MaxEntries <= 0 {
		config.MaxEntries = defaults.MaxEntries
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = defaults.CleanupInterval
	}
	if config.MaxIdle <= 0 {
		config.MaxIdle = defaults.MaxIdle
	}
	if config.Shards <= 0 {
		config.Shards = defaults.Shards
	}
	if config.Now == nil {
		config.Now = time.Now
	}

	limiter := &Limiter{
		shards:          make([]limiterShard, config.Shards),
		rate:            config.Rate,
		capacity:        float64(config.Capacity),
		cleanupInterval: config.CleanupInterval,
		maxIdle:         config.MaxIdle,
		maxEntries:      config.MaxEntries,
		now:             config.Now,
		stopCh:          make(chan struct{}),
		stoppedCh:       make(chan struct{}),
	}

	// Initialize shard counter
	limiter.bucketCount.counters = make([]atomic.Int64, config.Shards)

	for i := range limiter.shards {
		limiter.shards[i].buckets = make(map[string]*bucket)
	}

	limiter.startCleanup()

	return limiter
}

// Allow checks and consumes a token for the given key.
func (l *Limiter) Allow(key string) Decision {
	if key == "" {
		key = "unknown"
	}

	now := l.now()
	if l.maxEntries > 0 && l.bucketCount.Load() >= int64(l.maxEntries) {
		l.evictOldestGlobal()
	}
	shard := l.shardFor(key)
	shardIdx := fnv32a(key) % uint32(len(l.shards))
	shard.mu.Lock()
	b, ok := shard.buckets[key]
	if !ok {
		if l.maxEntries > 0 && l.bucketCount.Load() >= int64(l.maxEntries) {
			l.evictOldestLocked(shard)
		}
		b = &bucket{
			tokens:     l.capacity,
			lastRefill: now,
			lastAccess: now,
		}
		shard.buckets[key] = b
		l.bucketCount.Add(int(shardIdx), 1)
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}
	b.tokens = math.Min(l.capacity, b.tokens+elapsed*l.rate)
	b.lastRefill = now
	b.lastAccess = now

	decision := Decision{
		Allowed:   b.tokens >= 1,
		Limit:     int(l.capacity),
		Remaining: int(math.Floor(b.tokens)),
		Reset:     now,
	}

	if decision.Allowed {
		b.tokens -= 1
		decision.Remaining = int(math.Floor(b.tokens))
	} else if l.rate > 0 {
		need := 1 - b.tokens
		if need < 0 {
			need = 0
		}
		decision.RetryAfter = durationFromSeconds(need / l.rate)
	}

	if l.rate > 0 {
		missing := l.capacity - b.tokens
		if missing < 0 {
			missing = 0
		}
		decision.Reset = now.Add(durationFromSeconds(missing / l.rate))
	}

	shard.mu.Unlock()

	// Record metrics
	l.RecordAllow(decision.Allowed)

	return decision
}

// Stop stops the cleanup loop.
func (l *Limiter) Stop() {
	l.stopOnce.Do(func() {
		close(l.stopCh)
		<-l.stoppedCh
	})
}

func (l *Limiter) startCleanup() {
	l.startOnce.Do(func() {
		go l.cleanupLoop()
	})
}

func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanupInterval)
	defer ticker.Stop()
	defer close(l.stoppedCh)

	for {
		select {
		case <-ticker.C:
			l.cleanup(l.now())
		case <-l.stopCh:
			return
		}
	}
}

func (l *Limiter) cleanup(now time.Time) {
	for i := range l.shards {
		shard := &l.shards[i]
		shard.mu.Lock()
		for key, b := range shard.buckets {
			if now.Sub(b.lastAccess) > l.maxIdle {
				delete(shard.buckets, key)
				l.bucketCount.Add(i, -1)
			}
		}
		shard.mu.Unlock()
	}
}

func (l *Limiter) shardFor(key string) *limiterShard {
	if len(l.shards) == 1 {
		return &l.shards[0]
	}
	idx := fnv32a(key) % uint32(len(l.shards))
	return &l.shards[idx]
}

func (l *Limiter) evictOldestLocked(shard *limiterShard) {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for key, b := range shard.buckets {
		if first || b.lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = b.lastAccess
			first = false
		}
	}
	if oldestKey != "" {
		delete(shard.buckets, oldestKey)
		// Note: We can't easily determine the shard index here without the key
		// For simplicity in this method, we'll use the global counter approach
		// In a production system, you might want to pass the shard index
		l.bucketCount.Add(0, -1) // This is not ideal but works for now
	}
}

func (l *Limiter) evictOldestGlobal() {
	var oldestShard *limiterShard
	var oldestKey string
	var oldestTime time.Time
	first := true

	// Phase 1: Find the oldest entry across all shards
	for i := range l.shards {
		shard := &l.shards[i]
		shard.mu.Lock()
		for key, b := range shard.buckets {
			if first || b.lastAccess.Before(oldestTime) {
				oldestKey = key
				oldestTime = b.lastAccess
				oldestShard = shard
				first = false
			}
		}
		shard.mu.Unlock()
	}

	if oldestShard == nil || oldestKey == "" {
		return
	}

	// Phase 2: Double-check and remove the oldest entry
	// This prevents TOCTOU race where the entry might have been modified
	oldestShard.mu.Lock()
	if bucket, exists := oldestShard.buckets[oldestKey]; exists && bucket.lastAccess.Equal(oldestTime) {
		delete(oldestShard.buckets, oldestKey)
		// Calculate shard index for the deleted key
		shardIdx := fnv32a(oldestKey) % uint32(len(l.shards))
		l.bucketCount.Add(int(shardIdx), -1)
		l.metrics.Evictions.Add(1) // Record eviction metric
	}
	oldestShard.mu.Unlock()
}

// Metrics returns the current metrics for monitoring
func (l *Limiter) Metrics() *LimiterMetrics {
	return &l.metrics
}

// RecordAllow records the outcome of an Allow check for metrics
func (l *Limiter) RecordAllow(allowed bool) {
	if allowed {
		l.metrics.Allowed.Add(1)
	} else {
		l.metrics.Rejected.Add(1)
	}
}

func durationFromSeconds(seconds float64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	const maxDuration = time.Duration(1<<63 - 1)
	maxSeconds := float64(maxDuration) / float64(time.Second)
	if seconds > maxSeconds {
		return maxDuration
	}
	return time.Duration(seconds * float64(time.Second))
}

func fnv32a(key string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	hash := uint32(offset32)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= prime32
	}
	return hash
}
