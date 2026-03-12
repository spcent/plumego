package pubsub

import (
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// EnhancedMetrics provides detailed metrics including latency histograms.
type EnhancedMetrics struct {
	mu     sync.RWMutex
	topics map[string]*enhancedTopicMetrics
}

// enhancedTopicMetrics contains detailed metrics for a single topic.
type enhancedTopicMetrics struct {
	// Publish latency tracking
	publishLatencies *latencyHistogram

	// Deliver latency tracking
	deliverLatencies *latencyHistogram

	// Queue depth tracking
	queueDepthSum   atomic.Int64
	queueDepthCount atomic.Int64
	queueDepthMax   atomic.Int64

	// Message size tracking
	messageSizeSum   atomic.Int64
	messageSizeCount atomic.Int64
	messageSizeMax   atomic.Int64

	// Throughput tracking (messages per second)
	publishTimes *slidingWindow
	deliverTimes *slidingWindow
}

// latencyHistogram tracks latency distribution with percentiles.
type latencyHistogram struct {
	mu      sync.Mutex
	buckets []int64 // bucket counts
	bounds  []time.Duration
	count   int64
	sum     time.Duration
	min     time.Duration
	max     time.Duration
}

// Default latency buckets (microseconds to seconds)
var defaultLatencyBounds = []time.Duration{
	100 * time.Microsecond,
	500 * time.Microsecond,
	1 * time.Millisecond,
	5 * time.Millisecond,
	10 * time.Millisecond,
	25 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
	250 * time.Millisecond,
	500 * time.Millisecond,
	1 * time.Second,
	2500 * time.Millisecond,
	5 * time.Second,
	10 * time.Second,
}

// newLatencyHistogram creates a new latency histogram.
func newLatencyHistogram() *latencyHistogram {
	return &latencyHistogram{
		buckets: make([]int64, len(defaultLatencyBounds)+1),
		bounds:  defaultLatencyBounds,
		min:     time.Duration(math.MaxInt64),
	}
}

// Record records a latency observation.
func (h *latencyHistogram) Record(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.count++
	h.sum += d

	if d < h.min {
		h.min = d
	}
	if d > h.max {
		h.max = d
	}

	// Find bucket
	idx := len(h.bounds)
	for i, bound := range h.bounds {
		if d <= bound {
			idx = i
			break
		}
	}
	h.buckets[idx]++
}

// Percentile calculates the approximate percentile value.
func (h *latencyHistogram) Percentile(p float64) time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.count == 0 {
		return 0
	}

	target := int64(float64(h.count) * p / 100.0)
	var cumulative int64

	for i, count := range h.buckets {
		cumulative += count
		if cumulative >= target {
			if i == 0 {
				return h.bounds[0]
			}
			if i >= len(h.bounds) {
				return h.max
			}
			return h.bounds[i-1]
		}
	}

	return h.max
}

// Stats returns histogram statistics.
func (h *latencyHistogram) Stats() LatencyStats {
	h.mu.Lock()
	defer h.mu.Unlock()

	stats := LatencyStats{
		Count: h.count,
	}

	if h.count > 0 {
		stats.Min = h.min
		stats.Max = h.max
		stats.Avg = time.Duration(int64(h.sum) / h.count)
	}

	return stats
}

// LatencyStats contains latency statistics.
type LatencyStats struct {
	Count int64
	Min   time.Duration
	Max   time.Duration
	Avg   time.Duration
	P50   time.Duration
	P90   time.Duration
	P95   time.Duration
	P99   time.Duration
}

// slidingWindow tracks events in a time window for throughput calculation.
type slidingWindow struct {
	mu      sync.Mutex
	events  []time.Time
	window  time.Duration
	maxSize int
}

// newSlidingWindow creates a new sliding window.
func newSlidingWindow(window time.Duration, maxSize int) *slidingWindow {
	if maxSize <= 0 {
		maxSize = 10000
	}
	return &slidingWindow{
		events:  make([]time.Time, 0, maxSize),
		window:  window,
		maxSize: maxSize,
	}
}

// Record records an event.
func (sw *slidingWindow) Record() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()

	// Evict old events
	cutoff := now.Add(-sw.window)
	idx := sort.Search(len(sw.events), func(i int) bool {
		return sw.events[i].After(cutoff)
	})
	if idx > 0 {
		sw.events = sw.events[idx:]
	}

	// Add new event (with size limit)
	if len(sw.events) < sw.maxSize {
		sw.events = append(sw.events, now)
	}
}

// Rate returns the events per second.
func (sw *slidingWindow) Rate() float64 {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if len(sw.events) == 0 {
		return 0
	}

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// Count events in window
	var count int
	for i := len(sw.events) - 1; i >= 0; i-- {
		if sw.events[i].After(cutoff) {
			count++
		} else {
			break
		}
	}

	return float64(count) / sw.window.Seconds()
}

// NewEnhancedMetrics creates new enhanced metrics.
func NewEnhancedMetrics() *EnhancedMetrics {
	return &EnhancedMetrics{
		topics: make(map[string]*enhancedTopicMetrics),
	}
}

// getOrCreateTopic gets or creates metrics for a topic.
func (em *EnhancedMetrics) getOrCreateTopic(topic string) *enhancedTopicMetrics {
	em.mu.Lock()
	defer em.mu.Unlock()

	if tm, ok := em.topics[topic]; ok {
		return tm
	}

	tm := &enhancedTopicMetrics{
		publishLatencies: newLatencyHistogram(),
		deliverLatencies: newLatencyHistogram(),
		publishTimes:     newSlidingWindow(time.Minute, 10000),
		deliverTimes:     newSlidingWindow(time.Minute, 10000),
	}
	em.topics[topic] = tm
	return tm
}

// RecordPublishLatency records a publish latency.
func (em *EnhancedMetrics) RecordPublishLatency(topic string, d time.Duration) {
	tm := em.getOrCreateTopic(topic)
	tm.publishLatencies.Record(d)
	tm.publishTimes.Record()
}

// RecordDeliverLatency records a deliver latency.
func (em *EnhancedMetrics) RecordDeliverLatency(topic string, d time.Duration) {
	tm := em.getOrCreateTopic(topic)
	tm.deliverLatencies.Record(d)
	tm.deliverTimes.Record()
}

// RecordQueueDepth records a queue depth observation.
func (em *EnhancedMetrics) RecordQueueDepth(topic string, depth int) {
	tm := em.getOrCreateTopic(topic)
	tm.queueDepthSum.Add(int64(depth))
	tm.queueDepthCount.Add(1)

	// Update max
	for {
		current := tm.queueDepthMax.Load()
		if int64(depth) <= current {
			break
		}
		if tm.queueDepthMax.CompareAndSwap(current, int64(depth)) {
			break
		}
	}
}

// RecordMessageSize records a message size.
func (em *EnhancedMetrics) RecordMessageSize(topic string, size int) {
	tm := em.getOrCreateTopic(topic)
	tm.messageSizeSum.Add(int64(size))
	tm.messageSizeCount.Add(1)

	// Update max
	for {
		current := tm.messageSizeMax.Load()
		if int64(size) <= current {
			break
		}
		if tm.messageSizeMax.CompareAndSwap(current, int64(size)) {
			break
		}
	}
}

// Snapshot returns a snapshot of all enhanced metrics.
func (em *EnhancedMetrics) Snapshot() EnhancedMetricsSnapshot {
	em.mu.RLock()
	defer em.mu.RUnlock()

	snapshot := EnhancedMetricsSnapshot{
		Topics: make(map[string]EnhancedTopicMetrics, len(em.topics)),
	}

	for topic, tm := range em.topics {
		pubStats := tm.publishLatencies.Stats()
		pubStats.P50 = tm.publishLatencies.Percentile(50)
		pubStats.P90 = tm.publishLatencies.Percentile(90)
		pubStats.P95 = tm.publishLatencies.Percentile(95)
		pubStats.P99 = tm.publishLatencies.Percentile(99)

		delStats := tm.deliverLatencies.Stats()
		delStats.P50 = tm.deliverLatencies.Percentile(50)
		delStats.P90 = tm.deliverLatencies.Percentile(90)
		delStats.P95 = tm.deliverLatencies.Percentile(95)
		delStats.P99 = tm.deliverLatencies.Percentile(99)

		var queueDepthAvg float64
		if count := tm.queueDepthCount.Load(); count > 0 {
			queueDepthAvg = float64(tm.queueDepthSum.Load()) / float64(count)
		}

		var messageSizeAvg float64
		if count := tm.messageSizeCount.Load(); count > 0 {
			messageSizeAvg = float64(tm.messageSizeSum.Load()) / float64(count)
		}

		snapshot.Topics[topic] = EnhancedTopicMetrics{
			PublishLatency: pubStats,
			DeliverLatency: delStats,
			QueueDepthAvg:  queueDepthAvg,
			QueueDepthMax:  int(tm.queueDepthMax.Load()),
			MessageSizeAvg: messageSizeAvg,
			MessageSizeMax: int(tm.messageSizeMax.Load()),
			PublishRate:    tm.publishTimes.Rate(),
			DeliverRate:    tm.deliverTimes.Rate(),
		}
	}

	return snapshot
}

// EnhancedMetricsSnapshot contains all enhanced metrics.
type EnhancedMetricsSnapshot struct {
	Topics map[string]EnhancedTopicMetrics
}

// EnhancedTopicMetrics contains detailed metrics for a topic.
type EnhancedTopicMetrics struct {
	PublishLatency LatencyStats
	DeliverLatency LatencyStats
	QueueDepthAvg  float64
	QueueDepthMax  int
	MessageSizeAvg float64
	MessageSizeMax int
	PublishRate    float64 // messages per second
	DeliverRate    float64 // messages per second
}
