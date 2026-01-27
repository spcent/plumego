package middleware

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
)

// RateLimiter provides intelligent concurrency control with backpressure, monitoring, and dynamic adjustment.
//
// RateLimiter is an advanced concurrency limiter that uses a semaphore-based approach to limit
// the number of concurrent requests. It also supports queueing with timeout, dynamic adjustment
// based on system pressure, and comprehensive metrics collection.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	config := middleware.RateLimiterConfig{
//		MaxConcurrent:  100,          // Maximum 100 concurrent requests
//		QueueDepth:     200,          // Queue depth of 200
//		QueueTimeout:   1 * time.Second, // 1 second timeout for queue
//		AdjustInterval: 10 * time.Second, // Adjust concurrency every 10 seconds
//		MinConcurrent:  50,           // Minimum 50 concurrent requests
//		MaxConcurrentLimit: 200,      // Maximum 200 concurrent requests (for dynamic adjustment)
//		WindowSize:     10,           // Pressure sampling window size
//		Logger:         middleware.NewGLogger(),
//	}
//
//	limiter := middleware.NewRateLimiter(config)
//	handler := limiter.Middleware()(myHandler)
//
// The limiter provides the following features:
//   - Concurrency limiting using a semaphore
//   - Request queueing with timeout
//   - Dynamic adjustment based on system pressure
//   - Comprehensive metrics collection
//   - Backpressure mechanism
//
// Pressure levels:
//   - PressureLow: System is underutilized, can increase concurrency
//   - PressureNormal: System is operating normally
//   - PressureHigh: System is under high load, should decrease concurrency
//   - PressureCritical: System is critically overloaded, must decrease concurrency
//
// Metrics:
//   - CurrentConcurrent: Number of currently processing requests
//   - CurrentQueue: Number of requests waiting in queue
//   - TotalRequests: Total number of requests received
//   - AcceptedRequests: Number of requests accepted and processed
//   - RejectedRequests: Number of requests rejected (queue full)
//   - TimeoutRequests: Number of requests that timed out waiting for worker
//   - PressureLevel: Current system pressure level
type RateLimiter struct {
	// Configuration parameters
	maxConcurrent      int64         // Maximum concurrent requests
	queueDepth         int64         // Queue depth
	queueTimeout       time.Duration // Queue timeout
	adjustInterval     time.Duration // Dynamic adjustment interval
	minConcurrent      int64         // Minimum concurrent (for dynamic adjustment)
	maxConcurrentLimit int64         // Maximum concurrent limit (for dynamic adjustment)

	// Runtime state
	currentConcurrent int64 // Current concurrent requests
	currentQueue      int64 // Current queue length
	totalRequests     int64 // Total requests
	rejectedRequests  int64 // Rejected requests
	timeoutRequests   int64 // Timeout requests
	acceptedRequests  int64 // Successfully processed requests

	// Backpressure mechanism
	pressureWindow []pressureSample // Pressure sampling window
	windowSize     int              // Window size
	highPressure   int64            // High pressure counter
	lowPressure    int64            // Low pressure counter

	// Synchronization primitives
	sem       chan struct{} // Concurrency control semaphore
	queue     chan struct{} // Queue semaphore
	mutex     sync.RWMutex  // Protects shared state
	adjustMux sync.Mutex    // Protects adjustment logic

	// Monitoring and logging
	logger      log.StructuredLogger
	metricsHook func(metrics ConcurrencyMetrics) // Metrics callback

	// Dynamic adjustment
	lastAdjustTime time.Time
	adjustTicker   *time.Ticker
	stopChan       chan struct{}
}

// ConcurrencyMetrics contains runtime metrics for concurrency control
type ConcurrencyMetrics struct {
	Timestamp         time.Time
	CurrentConcurrent int64
	CurrentQueue      int64
	MaxConcurrent     int64
	QueueDepth        int64
	TotalRequests     int64
	AcceptedRequests  int64
	RejectedRequests  int64
	TimeoutRequests   int64
	PressureLevel     PressureLevel
	AdjustmentCount   int64
}

// PressureLevel represents system pressure level
type PressureLevel int

const (
	PressureLow      PressureLevel = 0
	PressureNormal   PressureLevel = 1
	PressureHigh     PressureLevel = 2
	PressureCritical PressureLevel = 3
)

// pressureSample contains sampling data for pressure analysis
type pressureSample struct {
	timestamp   time.Time
	concurrent  int64
	queueLength int64
	accepted    int64
	rejected    int64
	timeout     int64
}

// RateLimiterConfig configuration parameters
type RateLimiterConfig struct {
	MaxConcurrent      int64                            // Maximum concurrent requests
	QueueDepth         int64                            // Queue depth
	QueueTimeout       time.Duration                    // Queue timeout
	AdjustInterval     time.Duration                    // Dynamic adjustment interval (0 = no adjustment)
	MinConcurrent      int64                            // Minimum concurrent (for dynamic adjustment)
	MaxConcurrentLimit int64                            // Maximum concurrent limit (for dynamic adjustment)
	WindowSize         int                              // Pressure sampling window size
	MetricsHook        func(metrics ConcurrencyMetrics) // Metrics callback
	Logger             log.StructuredLogger
}

// NewRateLimiter creates an advanced concurrency limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	// Set default values
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = 100
	}
	if config.QueueDepth < config.MaxConcurrent {
		config.QueueDepth = config.MaxConcurrent * 2
	}
	if config.QueueTimeout <= 0 {
		config.QueueTimeout = 100 * time.Millisecond
	}
	if config.MinConcurrent <= 0 {
		config.MinConcurrent = config.MaxConcurrent / 2
	}
	if config.MaxConcurrentLimit <= 0 {
		config.MaxConcurrentLimit = config.MaxConcurrent * 3
	}
	if config.WindowSize <= 0 {
		config.WindowSize = 10
	}

	rl := &RateLimiter{
		maxConcurrent:      config.MaxConcurrent,
		queueDepth:         config.QueueDepth,
		queueTimeout:       config.QueueTimeout,
		adjustInterval:     config.AdjustInterval,
		minConcurrent:      config.MinConcurrent,
		maxConcurrentLimit: config.MaxConcurrentLimit,
		windowSize:         config.WindowSize,
		logger:             config.Logger,
		metricsHook:        config.MetricsHook,
		sem:                make(chan struct{}, config.MaxConcurrent),
		queue:              make(chan struct{}, config.QueueDepth),
		pressureWindow:     make([]pressureSample, 0, config.WindowSize),
		stopChan:           make(chan struct{}),
		lastAdjustTime:     time.Now(),
	}

	// Start dynamic adjuster (if enabled)
	if config.AdjustInterval > 0 {
		rl.adjustTicker = time.NewTicker(config.AdjustInterval)
		go rl.adjustmentLoop()
	}

	return rl
}

// Middleware creates the middleware
func (rl *RateLimiter) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rl.handleRequest(w, r, next)
		})
	}
}

// handleRequest processes a single request
func (rl *RateLimiter) handleRequest(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// Atomically increment total requests
	atomic.AddInt64(&rl.totalRequests, 1)

	// Step 1: Try to enter queue (non-blocking)
	if !rl.tryEnterQueue() {
		rl.rejectRequest(w, r, "server_busy", "Server is busy, queue full")
		atomic.AddInt64(&rl.rejectedRequests, 1)
		return
	}
	defer rl.leaveQueue()

	// Step 2: Wait for concurrency slot (with timeout)
	if !rl.waitForConcurrencySlot() {
		rl.rejectRequest(w, r, "server_queue_timeout", "Request timed out waiting for worker")
		atomic.AddInt64(&rl.timeoutRequests, 1)
		return
	}
	defer rl.leaveConcurrency()

	// Record successfully accepted request
	atomic.AddInt64(&rl.acceptedRequests, 1)

	// Execute next handler
	next.ServeHTTP(w, r)
}

// tryEnterQueue attempts to enter the queue
func (rl *RateLimiter) tryEnterQueue() bool {
	// Non-blocking attempt to send to queue
	select {
	case rl.queue <- struct{}{}:
		atomic.AddInt64(&rl.currentQueue, 1)
		return true
	default:
		return false
	}
}

// leaveQueue exits the queue
func (rl *RateLimiter) leaveQueue() {
	// Non-blocking queue exit
	select {
	case <-rl.queue:
		atomic.AddInt64(&rl.currentQueue, -1)
	default:
	}
}

// waitForConcurrencySlot waits for a concurrency slot
func (rl *RateLimiter) waitForConcurrencySlot() bool {
	timer := time.NewTimer(rl.queueTimeout)
	defer timer.Stop()

	// Wait for concurrency slot or timeout
	select {
	case rl.sem <- struct{}{}:
		atomic.AddInt64(&rl.currentConcurrent, 1)
		return true
	case <-timer.C:
		return false
	}
}

// leaveConcurrency exits the concurrency slot
func (rl *RateLimiter) leaveConcurrency() {
	// Non-blocking concurrency slot exit
	select {
	case <-rl.sem:
		atomic.AddInt64(&rl.currentConcurrent, -1)
	default:
	}
}

// rejectRequest rejects a request
func (rl *RateLimiter) rejectRequest(w http.ResponseWriter, r *http.Request, code, message string) {
	contract.WriteError(w, r, contract.APIError{
		Status:   http.StatusServiceUnavailable,
		Code:     code,
		Category: contract.CategoryServer,
		Message:  message,
		Details: map[string]any{
			"current_concurrent": atomic.LoadInt64(&rl.currentConcurrent),
			"current_queue":      atomic.LoadInt64(&rl.currentQueue),
			"max_concurrent":     atomic.LoadInt64(&rl.maxConcurrent),
			"queue_depth":        atomic.LoadInt64(&rl.queueDepth),
		},
	})

	if rl.logger != nil {
		rl.logger.WithFields(log.Fields{
			"code":               code,
			"current_concurrent": atomic.LoadInt64(&rl.currentConcurrent),
			"current_queue":      atomic.LoadInt64(&rl.currentQueue),
		}).Warn("request rejected due to concurrency limit", nil)
	}
}

// adjustmentLoop handles dynamic adjustment loop
func (rl *RateLimiter) adjustmentLoop() {
	if rl.adjustTicker == nil {
		return
	}

	for {
		select {
		case <-rl.adjustTicker.C:
			rl.adjustConcurrency()
		case <-rl.stopChan:
			return
		}
	}
}

// adjustConcurrency dynamically adjusts concurrency
func (rl *RateLimiter) adjustConcurrency() {
	rl.adjustMux.Lock()
	defer rl.adjustMux.Unlock()

	// Collect current metrics
	metrics := rl.GetMetrics()

	// Analyze pressure level
	pressure := rl.analyzePressure(metrics)

	// Adjust based on pressure
	switch pressure {
	case PressureLow:
		// Low pressure, try to increase concurrency
		if rl.maxConcurrent < rl.maxConcurrentLimit {
			newMax := min(rl.maxConcurrent+5, rl.maxConcurrentLimit)
			rl.updateMaxConcurrent(newMax)
			rl.logAdjustment("increased", rl.maxConcurrent, newMax)
		}
	case PressureCritical:
		// High pressure, try to decrease concurrency
		if rl.maxConcurrent > rl.minConcurrent {
			newMax := max(rl.maxConcurrent-10, rl.minConcurrent)
			rl.updateMaxConcurrent(newMax)
			rl.logAdjustment("decreased", rl.maxConcurrent, newMax)
		}
	}

	// Trigger metrics callback
	if rl.metricsHook != nil {
		rl.metricsHook(rl.GetMetrics())
	}
}

// analyzePressure analyzes system pressure
func (rl *RateLimiter) analyzePressure(metrics ConcurrencyMetrics) PressureLevel {
	// Sample current state
	sample := pressureSample{
		timestamp:   time.Now(),
		concurrent:  metrics.CurrentConcurrent,
		queueLength: metrics.CurrentQueue,
		accepted:    metrics.AcceptedRequests,
		rejected:    metrics.RejectedRequests,
		timeout:     metrics.TimeoutRequests,
	}

	// Add to sampling window
	rl.mutex.Lock()
	rl.pressureWindow = append(rl.pressureWindow, sample)
	if len(rl.pressureWindow) > rl.windowSize {
		rl.pressureWindow = rl.pressureWindow[1:]
	}
	window := rl.pressureWindow
	rl.mutex.Unlock()

	// Return normal if insufficient data
	if len(window) < 3 {
		return PressureNormal
	}

	// Calculate trend metrics
	recentRejected := window[len(window)-1].rejected - window[0].rejected
	recentTimeout := window[len(window)-1].timeout - window[0].timeout
	avgQueueLength := int64(0)
	avgConcurrent := int64(0)
	for _, s := range window {
		avgQueueLength += s.queueLength
		avgConcurrent += s.concurrent
	}
	avgQueueLength /= int64(len(window))
	avgConcurrent /= int64(len(window))

	// Pressure judgment logic
	if recentRejected > 5 || recentTimeout > 3 {
		return PressureCritical
	}
	if avgQueueLength > rl.queueDepth/2 || avgConcurrent > rl.maxConcurrent*80/100 {
		return PressureHigh
	}
	if avgQueueLength < rl.queueDepth/4 && avgConcurrent < rl.maxConcurrent*30/100 {
		return PressureLow
	}

	return PressureNormal
}

// updateMaxConcurrent updates maximum concurrency
func (rl *RateLimiter) updateMaxConcurrent(newMax int64) {
	oldMax := atomic.LoadInt64(&rl.maxConcurrent)
	if newMax == oldMax {
		return
	}

	// Rebuild semaphore channel
	newSem := make(chan struct{}, newMax)

	// Copy existing slots
	go func() {
		for i := int64(0); i < min(oldMax, newMax); i++ {
			select {
			case <-rl.sem:
				newSem <- struct{}{}
			default:
				return
			}
		}
	}()

	rl.sem = newSem
	atomic.StoreInt64(&rl.maxConcurrent, newMax)
}

// logAdjustment records adjustment log
func (rl *RateLimiter) logAdjustment(direction string, oldVal, newVal int64) {
	if rl.logger != nil {
		rl.logger.WithFields(log.Fields{
			"direction": direction,
			"old_value": oldVal,
			"new_value": newVal,
		}).Info("concurrency limit adjusted", nil)
	}
}

// GetMetrics gets current runtime metrics
func (rl *RateLimiter) GetMetrics() ConcurrencyMetrics {
	metrics := ConcurrencyMetrics{
		Timestamp:         time.Now(),
		CurrentConcurrent: atomic.LoadInt64(&rl.currentConcurrent),
		CurrentQueue:      atomic.LoadInt64(&rl.currentQueue),
		MaxConcurrent:     atomic.LoadInt64(&rl.maxConcurrent),
		QueueDepth:        atomic.LoadInt64(&rl.queueDepth),
		TotalRequests:     atomic.LoadInt64(&rl.totalRequests),
		AcceptedRequests:  atomic.LoadInt64(&rl.acceptedRequests),
		RejectedRequests:  atomic.LoadInt64(&rl.rejectedRequests),
		TimeoutRequests:   atomic.LoadInt64(&rl.timeoutRequests),
		AdjustmentCount:   0,
	}
	// Analyze pressure (avoid recursive call)
	metrics.PressureLevel = rl.analyzePressure(metrics)
	return metrics
}

// Stop stops the dynamic adjuster
func (rl *RateLimiter) Stop() {
	if rl.adjustTicker != nil {
		rl.adjustTicker.Stop()
	}
	close(rl.stopChan)
}

// Helper functions
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// RateLimitMiddleware creates a middleware for advanced concurrency limiting
func RateLimitMiddleware(config RateLimiterConfig) Middleware {
	rl := NewRateLimiter(config)
	return rl.Middleware()
}

// Legacy RateLimit functions for backward compatibility
// NewRateLimiterLegacy creates a token bucket rate limiter (legacy)
func NewRateLimiterLegacy(rate float64, capacity int, cleanupInterval, maxIdleTime time.Duration) *RateLimiterLegacy {
	return &RateLimiterLegacy{
		buckets:         make(map[string]*tokenBucket),
		rate:            rate,
		capacity:        capacity,
		cleanupInterval: cleanupInterval,
		maxIdleTime:     maxIdleTime,
	}
}

// RateLimiterLegacy represents a rate limiter that uses token bucket algorithm
type RateLimiterLegacy struct {
	mu              sync.RWMutex
	buckets         map[string]*tokenBucket
	rate            float64
	capacity        int
	cleanupInterval time.Duration
	maxIdleTime     time.Duration
}

// tokenBucket represents a single token bucket for rate limiting
type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
	lastAccess time.Time
}

// Allow checks if a request from the given key is allowed
func (rl *RateLimiterLegacy) Allow(key string) bool {
	now := time.Now()
	bucket := rl.getBucket(key, now)

	// Calculate tokens to add since last refill
	elapsed := now.Sub(bucket.lastRefill)
	addTokens := rl.rate * elapsed.Seconds()

	// Update bucket
	newTokens := bucket.tokens + addTokens
	if newTokens > float64(rl.capacity) {
		newTokens = float64(rl.capacity)
	}
	bucket.tokens = newTokens
	bucket.lastRefill = now
	bucket.lastAccess = now

	// Check if we have enough tokens
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}

// getBucket returns the token bucket for the given key, creating it if needed
func (rl *RateLimiterLegacy) getBucket(key string, now time.Time) *tokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(rl.capacity),
			lastRefill: now,
			lastAccess: now,
		}
		rl.buckets[key] = bucket
	}

	return bucket
}

// cleanup removes idle buckets to free memory
func (rl *RateLimiterLegacy) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()

		for key, bucket := range rl.buckets {
			if now.Sub(bucket.lastAccess) > rl.maxIdleTime {
				delete(rl.buckets, key)
			}
		}

		rl.mu.Unlock()
	}
}

// RateLimit creates a middleware that limits requests based on IP address (legacy)
func RateLimit(rate float64, capacity int, cleanupInterval, maxIdleTime time.Duration) Middleware {
	limiter := NewRateLimiterLegacy(rate, capacity, cleanupInterval, maxIdleTime)
	go limiter.cleanup()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !limiter.Allow(ip) {
				contract.WriteError(w, r, contract.APIError{
					Status:   http.StatusTooManyRequests,
					Code:     "rate_limited",
					Message:  "Too many requests, please try again later",
					Category: contract.CategoryRateLimit,
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitFunc creates a Middleware version of RateLimit (legacy)
func RateLimitFunc(rate float64, capacity int, cleanupInterval, maxIdleTime time.Duration) Middleware {
	limiter := NewRateLimiterLegacy(rate, capacity, cleanupInterval, maxIdleTime)
	go limiter.cleanup()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !limiter.Allow(ip) {
				contract.WriteError(w, r, contract.APIError{
					Status:   http.StatusTooManyRequests,
					Code:     "rate_limited",
					Message:  "Too many requests, please try again later",
					Category: contract.CategoryRateLimit,
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
