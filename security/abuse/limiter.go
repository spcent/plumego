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
type Config struct {
	Rate            float64
	Capacity        int
	CleanupInterval time.Duration
	MaxIdle         time.Duration
	Now             func() time.Time
}

// Decision captures the outcome of a limiter check.
type Decision struct {
	Allowed    bool
	Limit      int
	Remaining  int
	Reset      time.Time
	RetryAfter time.Duration
}

// Limiter enforces a token bucket per key.
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
