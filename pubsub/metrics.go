package pubsub

import (
	"sync"
	"sync/atomic"
)

type MetricsSnapshot struct {
	Topics map[string]TopicMetrics `json:"topics"`
}

type TopicMetrics struct {
	PublishTotal     uint64            `json:"publish_total"`
	DeliveredTotal   uint64            `json:"delivered_total"`
	DroppedByPolicy  map[string]uint64 `json:"dropped_by_policy"`
	SubscribersGauge int               `json:"subscribers"`
}

type metrics struct {
	// topic -> *topicMetrics
	topics sync.Map
}

type topicMetrics struct {
	publishTotal   atomic.Uint64
	deliveredTotal atomic.Uint64
	subsGauge      atomic.Int64

	// policyName -> atomic counter (uint64)
	dropped sync.Map
}

func (m *metrics) ensureTopic(topic string) *topicMetrics {
	if v, ok := m.topics.Load(topic); ok {
		return v.(*topicMetrics)
	}
	tm := &topicMetrics{}
	actual, _ := m.topics.LoadOrStore(topic, tm)
	return actual.(*topicMetrics)
}

func (m *metrics) incPublish(topic string) {
	m.ensureTopic(topic).publishTotal.Add(1)
}

func (m *metrics) incDelivered(topic string) {
	m.ensureTopic(topic).deliveredTotal.Add(1)
}

func (m *metrics) addSubs(topic string, delta int64) {
	m.ensureTopic(topic).subsGauge.Add(delta)
}

func (m *metrics) incDropped(topic string, policy BackpressurePolicy) {
	tm := m.ensureTopic(topic)
	key := policyName(policy)
	v, _ := tm.dropped.LoadOrStore(key, new(atomic.Uint64))
	v.(*atomic.Uint64).Add(1)
}

func (m *metrics) Snapshot() MetricsSnapshot {
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

func policyName(p BackpressurePolicy) string {
	switch p {
	case DropOldest:
		return "drop_oldest"
	case DropNewest:
		return "drop_newest"
	case BlockWithTimeout:
		return "block_with_timeout"
	case CloseSubscriber:
		return "close_subscriber"
	default:
		return "unknown"
	}
}
