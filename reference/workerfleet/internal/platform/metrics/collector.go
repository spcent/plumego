package metrics

import (
	"sort"
	"strings"
	"sync"
)

type Collector struct {
	mu         sync.RWMutex
	gauges     map[seriesKey]float64
	counters   map[seriesKey]float64
	histograms map[seriesKey]*histogramState
	buckets    []float64
}

type seriesKey struct {
	name   string
	labels string
}

type histogramState struct {
	buckets []uint64
	count   uint64
	sum     float64
}

func NewCollector() *Collector {
	return &Collector{
		gauges:     make(map[seriesKey]float64),
		counters:   make(map[seriesKey]float64),
		histograms: make(map[seriesKey]*histogramState),
		buckets:    defaultBuckets(),
	}
}

func (c *Collector) SetGauge(name string, labels map[string]string, value float64) error {
	if err := ValidateLabels(labels); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gauges[newSeriesKey(name, labels)] = value
	return nil
}

func (c *Collector) AddCounter(name string, labels map[string]string, delta float64) error {
	if err := ValidateLabels(labels); err != nil {
		return err
	}
	if delta < 0 {
		delta = 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[newSeriesKey(name, labels)] += delta
	return nil
}

func (c *Collector) ObserveHistogram(name string, labels map[string]string, value float64) error {
	if err := ValidateLabels(labels); err != nil {
		return err
	}
	if value < 0 {
		value = 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	key := newSeriesKey(name, labels)
	state := c.histograms[key]
	if state == nil {
		state = &histogramState{buckets: make([]uint64, len(c.buckets))}
		c.histograms[key] = state
	}
	for i, bucket := range c.buckets {
		if value <= bucket {
			state.buckets[i]++
		}
	}
	state.count++
	state.sum += value
	return nil
}

func (c *Collector) snapshot() collectorSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snap := collectorSnapshot{
		gauges:     make(map[seriesKey]float64, len(c.gauges)),
		counters:   make(map[seriesKey]float64, len(c.counters)),
		histograms: make(map[seriesKey]histogramSnapshot, len(c.histograms)),
		buckets:    append([]float64(nil), c.buckets...),
	}
	for key, value := range c.gauges {
		snap.gauges[key] = value
	}
	for key, value := range c.counters {
		snap.counters[key] = value
	}
	for key, value := range c.histograms {
		snap.histograms[key] = histogramSnapshot{
			buckets: append([]uint64(nil), value.buckets...),
			count:   value.count,
			sum:     value.sum,
		}
	}
	return snap
}

type collectorSnapshot struct {
	gauges     map[seriesKey]float64
	counters   map[seriesKey]float64
	histograms map[seriesKey]histogramSnapshot
	buckets    []float64
}

type histogramSnapshot struct {
	buckets []uint64
	count   uint64
	sum     float64
}

func newSeriesKey(name string, labels map[string]string) seriesKey {
	return seriesKey{name: name, labels: canonicalLabels(labels)}
}

func canonicalLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, key := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(labels[key])
	}
	return b.String()
}

func defaultBuckets() []float64 {
	return []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 300, 900, 1800, 3600}
}
