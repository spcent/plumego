package observability

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/metrics"
)

type labelKey struct {
	method string
	path   string
	status string
}

type latencyStats struct {
	count uint64
	sum   float64
	min   float64
	max   float64
}

type PrometheusCollector struct {
	base *metrics.BaseMetricsCollector

	namespace string
	maxMemory int

	mu        sync.RWMutex
	requests  map[labelKey]uint64
	durations map[labelKey]latencyStats
	startTime time.Time
}

func NewPrometheusCollector(namespace string) *PrometheusCollector {
	if namespace == "" {
		namespace = "plumego"
	}
	return &PrometheusCollector{
		base:      metrics.NewBaseMetricsCollector(),
		namespace: namespace,
		maxMemory: 10000,
		requests:  make(map[labelKey]uint64),
		durations: make(map[labelKey]latencyStats),
		startTime: time.Now(),
	}
}

func (p *PrometheusCollector) WithMaxMemory(max int) *PrometheusCollector {
	if max > 0 {
		p.maxMemory = max
	}
	return p
}

func (p *PrometheusCollector) Handler() http.Handler {
	return NewPrometheusExporter(p).Handler()
}

func (p *PrometheusCollector) Record(ctx context.Context, record metrics.MetricRecord) {
	if record.Type == metrics.MetricHTTPRequest {
		method, _ := record.Labels["method"]
		path, _ := record.Labels["path"]
		statusStr, _ := record.Labels["status"]
		status, _ := strconv.Atoi(statusStr)
		if method != "" && path != "" && status != 0 {
			p.recordHTTP(method, path, status, record.Duration)
			return
		}
	}
	p.base.Record(ctx, record)
}

func (p *PrometheusCollector) ObserveHTTP(_ context.Context, method, path string, status, _ int, duration time.Duration) {
	p.recordHTTP(method, path, status, duration)
}

func (p *PrometheusCollector) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	p.base.ObservePubSub(ctx, operation, topic, duration, err)
}

func (p *PrometheusCollector) ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool) {
	p.base.ObserveMQ(ctx, operation, topic, duration, err, panicked)
}

func (p *PrometheusCollector) ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool) {
	p.base.ObserveKV(ctx, operation, key, duration, err, hit)
}

func (p *PrometheusCollector) ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error) {
	p.base.ObserveIPC(ctx, operation, addr, transport, bytes, duration, err)
}

func (p *PrometheusCollector) ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
	p.base.ObserveDB(ctx, operation, driver, query, rows, duration, err)
}

func (p *PrometheusCollector) GetStats() metrics.CollectorStats {
	p.mu.RLock()
	totalRequests := uint64(0)
	errorRecords := int64(0)
	for key, count := range p.requests {
		totalRequests += count
		status, err := strconv.Atoi(key.status)
		if err == nil && status >= http.StatusBadRequest {
			errorRecords += int64(count)
		}
	}
	startTime := p.startTime
	activeSeries := len(p.requests)
	p.mu.RUnlock()

	stats := metrics.CollectorStats{
		TotalRecords:  int64(totalRequests),
		ErrorRecords:  errorRecords,
		ActiveSeries:  activeSeries,
		StartTime:     startTime,
		TypeBreakdown: make(map[metrics.MetricType]int64),
	}
	if totalRequests > 0 {
		stats.TypeBreakdown[metrics.MetricHTTPRequest] = int64(totalRequests)
	}

	baseStats := p.base.GetStats()
	stats.TotalRecords += baseStats.TotalRecords
	stats.ErrorRecords += baseStats.ErrorRecords
	stats.ActiveSeries += baseStats.ActiveSeries
	if stats.StartTime.IsZero() || (!baseStats.StartTime.IsZero() && baseStats.StartTime.Before(stats.StartTime)) {
		stats.StartTime = baseStats.StartTime
	}
	for key, value := range baseStats.TypeBreakdown {
		stats.TypeBreakdown[key] += value
	}

	return stats
}

func (p *PrometheusCollector) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.requests = make(map[labelKey]uint64)
	p.durations = make(map[labelKey]latencyStats)
	p.startTime = time.Now()
	p.base.Clear()
}

func (p *PrometheusCollector) snapshot() (map[labelKey]uint64, map[labelKey]latencyStats, time.Duration) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	reqCopy := make(map[labelKey]uint64, len(p.requests))
	for k, v := range p.requests {
		reqCopy[k] = v
	}

	durCopy := make(map[labelKey]latencyStats, len(p.durations))
	for k, v := range p.durations {
		durCopy[k] = v
	}

	return reqCopy, durCopy, time.Since(p.startTime)
}

func (p *PrometheusCollector) recordHTTP(method, path string, status int, duration time.Duration) {
	key := labelKey{method: method, path: path, status: strconv.Itoa(status)}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.maxMemory > 0 && len(p.requests) >= p.maxMemory {
		p.evictLeastUsed()
	}

	p.requests[key]++

	seconds := duration.Seconds()
	stats := p.durations[key]
	stats.count++
	stats.sum += seconds
	if stats.count == 1 || seconds < stats.min {
		stats.min = seconds
	}
	if stats.count == 1 || seconds > stats.max {
		stats.max = seconds
	}
	p.durations[key] = stats
}

func (p *PrometheusCollector) evictLeastUsed() {
	evictCount := len(p.requests) / 10
	if evictCount == 0 {
		evictCount = 1
	}

	keys := make([]labelKey, 0, len(p.requests))
	for key := range p.requests {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return p.requests[keys[i]] < p.requests[keys[j]]
	})

	for i := 0; i < evictCount && i < len(keys); i++ {
		delete(p.requests, keys[i])
		delete(p.durations, keys[i])
	}
}

func sortedKeys[T any](input map[labelKey]T) []labelKey {
	keys := make([]labelKey, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		if keys[i].path != keys[j].path {
			return keys[i].path < keys[j].path
		}
		return keys[i].status < keys[j].status
	})
	return keys
}

func escapeLabelValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

var _ metrics.AggregateCollector = (*PrometheusCollector)(nil)
