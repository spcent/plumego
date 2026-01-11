package pubsub

import (
	"sync"
	"sync/atomic"
)

// MetricsSnapshot provides a consistent view of all metrics.
type MetricsSnapshot struct {
	Topics map[string]TopicMetrics `json:"topics"`
}

// TopicMetrics contains metrics for a single topic.
type TopicMetrics struct {
	PublishTotal     uint64            `json:"publish_total"`
	DeliveredTotal   uint64            `json:"delivered_total"`
	DroppedByPolicy  map[string]uint64 `json:"dropped_by_policy"`
	SubscribersGauge int               `json:"subscribers"`
}

// metricsPubSub manages pubsub metricsPubSub with lock-free operations.
type metricsPubSub struct {
	topics sync.Map // map[string]*topicMetrics
}

// topicMetrics holds atomic counters for a topic.
type topicMetrics struct {
	publishTotal   atomic.Uint64
	deliveredTotal atomic.Uint64
	subsGauge      atomic.Int64
	dropped        sync.Map // map[string]*atomic.Uint64
}

// policyNames converts policy to string (cached for performance).
var policyNames = map[BackpressurePolicy]string{
	DropOldest:       "drop_oldest",
	DropNewest:       "drop_newest",
	BlockWithTimeout: "block_with_timeout",
	CloseSubscriber:  "close_subscriber",
}

// ensureTopic gets or creates topic metrics.
func (m *metricsPubSub) ensureTopic(topic string) *topicMetrics {
	if v, ok := m.topics.Load(topic); ok {
		return v.(*topicMetrics)
	}
	tm := &topicMetrics{}
	actual, _ := m.topics.LoadOrStore(topic, tm)
	return actual.(*topicMetrics)
}

// incPublish increments publish counter.
func (m *metricsPubSub) incPublish(topic string) {
	m.ensureTopic(topic).publishTotal.Add(1)
}

// incDelivered increments delivered counter.
func (m *metricsPubSub) incDelivered(topic string) {
	m.ensureTopic(topic).deliveredTotal.Add(1)
}

// addSubs modifies subscriber count (can be negative).
func (m *metricsPubSub) addSubs(topic string, delta int64) {
	m.ensureTopic(topic).subsGauge.Add(delta)
}

// incDropped increments dropped counter for a policy.
func (m *metricsPubSub) incDropped(topic string, policy BackpressurePolicy) {
	tm := m.ensureTopic(topic)
	key := policyName(policy)

	// Fast path: try existing key
	if v, ok := tm.dropped.Load(key); ok {
		v.(*atomic.Uint64).Add(1)
		return
	}

	// Slow path: create if not exists
	v, _ := tm.dropped.LoadOrStore(key, new(atomic.Uint64))
	v.(*atomic.Uint64).Add(1)
}

// Snapshot creates a consistent snapshot of all metrics.
func (m *metricsPubSub) Snapshot() MetricsSnapshot {
	out := MetricsSnapshot{Topics: map[string]TopicMetrics{}}

	m.topics.Range(func(k, v any) bool {
		topic := k.(string)
		tm := v.(*topicMetrics)

		drops := map[string]uint64{}
		tm.dropped.Range(func(pk, pv any) bool {
			drops[pk.(string)] = pv.(*atomic.Uint64).Load()
			return true
		})

		out.Topics[topic] = TopicMetrics{
			PublishTotal:     tm.publishTotal.Load(),
			DeliveredTotal:   tm.deliveredTotal.Load(),
			DroppedByPolicy:  drops,
			SubscribersGauge: int(tm.subsGauge.Load()),
		}
		return true
	})

	return out
}

// policyName converts policy to human-readable string.
func policyName(p BackpressurePolicy) string {
	if name, ok := policyNames[p]; ok {
		return name
	}
	return "unknown"
}
