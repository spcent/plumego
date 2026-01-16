package abuse

import (
	"math"
	"sync"
	"time"
)

const (
	defaultRate            = 100
	defaultCapacity        = 200
	defaultCleanupInterval = time.Minute
	defaultMaxIdle         = 5 * time.Minute
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

	// CleanupInterval is how often to clean up idle entries
	CleanupInterval time.Duration

	// MaxIdle is the maximum time an entry can be idle before being removed
	MaxIdle time.Duration

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
type Limiter struct {
	mu              sync.Mutex
	buckets         map[string]*bucket
	rate            float64
	capacity        float64
	cleanupInterval time.Duration
	maxIdle         time.Duration
	now             func() time.Time
	stopCh          chan struct{}
	stoppedCh       chan struct{}
	startOnce       sync.Once
	stopOnce        sync.Once
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
		CleanupInterval: defaultCleanupInterval,
		MaxIdle:         defaultMaxIdle,
	}
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
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = defaults.CleanupInterval
	}
	if config.MaxIdle <= 0 {
		config.MaxIdle = defaults.MaxIdle
	}
	if config.Now == nil {
		config.Now = time.Now
	}

	limiter := &Limiter{
		buckets:         make(map[string]*bucket),
		rate:            config.Rate,
		capacity:        float64(config.Capacity),
		cleanupInterval: config.CleanupInterval,
		maxIdle:         config.MaxIdle,
		now:             config.Now,
		stopCh:          make(chan struct{}),
		stoppedCh:       make(chan struct{}),
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
	l.mu.Lock()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{
			tokens:     l.capacity,
			lastRefill: now,
			lastAccess: now,
		}
		l.buckets[key] = b
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
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
		decision.RetryAfter = time.Duration(need / l.rate * float64(time.Second))
	}

	if l.rate > 0 {
		missing := l.capacity - b.tokens
		if missing < 0 {
			missing = 0
		}
		decision.Reset = now.Add(time.Duration(missing / l.rate * float64(time.Second)))
	}

	l.mu.Unlock()
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
	l.mu.Lock()
	defer l.mu.Unlock()

	for key, b := range l.buckets {
		if now.Sub(b.lastAccess) > l.maxIdle {
			delete(l.buckets, key)
		}
	}
}
