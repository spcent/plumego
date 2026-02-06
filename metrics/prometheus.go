package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/middleware/observability"
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

// PrometheusCollector implements MetricsCollector without third-party dependencies.
// It exposes a text-based metrics handler compatible with Prometheus exposition format.
//
// This collector is designed for production use with:
//   - Memory-efficient metric storage
//   - Automatic eviction of old metrics
//   - Prometheus-compatible output format
//   - No external dependencies
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	defer collector.Clear()
//
//	// Use in HTTP middleware
//	mux := http.NewServeMux()
//	mux.Handle("/metrics", collector.Handler())
//
//	// Record metrics
//	collector.ObserveHTTP(context.Background(), "GET", "/api/users", 200, 100, 50*time.Millisecond)
type PrometheusCollector struct {
	baseForwarder // provides ObservePubSub, ObserveMQ, ObserveKV, ObserveIPC, ObserveDB

	namespace string
	maxMemory int // Maximum number of metric series to store

	mu        sync.RWMutex
	requests  map[labelKey]uint64
	durations map[labelKey]latencyStats
	startTime time.Time

	smsQueueDepth     map[smsKey]float64
	smsSendLatency    map[smsKey]latencyStats
	smsReceiptDelay   map[smsKey]latencyStats
	smsProviderResult map[smsKey]uint64
	smsRetry          map[smsKey]uint64
	smsStatus         map[smsKey]uint64
}

// NewPrometheusCollector constructs an in-memory collector with the provided namespace.
// An empty namespace defaults to "plumego".
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
func NewPrometheusCollector(namespace string) *PrometheusCollector {
	if namespace == "" {
		namespace = "plumego"
	}
	return &PrometheusCollector{
		namespace: namespace,
		requests:  make(map[labelKey]uint64),
		durations: make(map[labelKey]latencyStats),
		startTime: time.Now(),
		maxMemory: 10000, // Default: allow up to 10k unique label combinations

		smsQueueDepth:     make(map[smsKey]float64),
		smsSendLatency:    make(map[smsKey]latencyStats),
		smsReceiptDelay:   make(map[smsKey]latencyStats),
		smsProviderResult: make(map[smsKey]uint64),
		smsRetry:          make(map[smsKey]uint64),
		smsStatus:         make(map[smsKey]uint64),
	}
}

// WithMaxMemory sets the maximum number of unique metric series to store.
// When exceeded, oldest entries are evicted.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp").
//		WithMaxMemory(50000)
func (p *PrometheusCollector) WithMaxMemory(max int) *PrometheusCollector {
	if max > 0 {
		p.maxMemory = max
	}
	return p
}

// Observe records a single HTTP request metric set.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//	import "github.com/spcent/plumego/middleware/observability"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	collector.Observe(context.Background(), observability.RequestMetrics{
//		Method:   "GET",
//		Path:     "/api/users",
//		Status:   200,
//		Bytes:    100,
//		Duration: 50 * time.Millisecond,
//	})
func (p *PrometheusCollector) Observe(_ context.Context, m observability.RequestMetrics) {
	key := labelKey{m.Method, m.Path, strconv.Itoa(m.Status)}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check memory limit and evict if needed
	if p.maxMemory > 0 && len(p.requests) >= p.maxMemory {
		p.evictOldest()
	}

	p.requests[key]++

	duration := m.Duration.Seconds()
	stats := p.durations[key]
	stats.count++
	stats.sum += duration

	// Update min/max
	if stats.count == 1 || duration < stats.min {
		stats.min = duration
	}
	if stats.count == 1 || duration > stats.max {
		stats.max = duration
	}

	p.durations[key] = stats
}

// Handler returns an HTTP handler that emits the current metrics snapshot.
//
// This handler exposes metrics in Prometheus exposition format (text/plain).
// The endpoint can be scraped by Prometheus or viewed directly in a browser.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	http.Handle("/metrics", collector.Handler())
//	http.ListenAndServe(":8080", nil)
func (p *PrometheusCollector) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests, durations, uptime := p.snapshot()
		smsSnapshot := p.snapshotSMSGateway()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		// Write metrics
		fmt.Fprintf(w, "# HELP %s_http_requests_total Total number of HTTP requests processed.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_requests_total counter\n", p.namespace)

		reqKeys := sortedKeys(requests)
		for _, k := range reqKeys {
			fmt.Fprintf(w, "%s_http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				p.namespace, escapeLabelValue(k.method), escapeLabelValue(k.path), escapeLabelValue(k.status), requests[k])
		}

		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_http_request_duration_seconds_sum Sum of HTTP request latencies in seconds.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_request_duration_seconds_summary summary\n", p.namespace)

		durKeys := sortedKeys(durations)
		for _, k := range durKeys {
			stats := durations[k]
			em, ep, es := escapeLabelValue(k.method), escapeLabelValue(k.path), escapeLabelValue(k.status)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_sum{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				p.namespace, em, ep, es, stats.sum)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_count{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				p.namespace, em, ep, es, stats.count)
			// Add min and max as additional metrics
			fmt.Fprintf(w, "%s_http_request_duration_seconds_min{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				p.namespace, em, ep, es, stats.min)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_max{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				p.namespace, em, ep, es, stats.max)
		}

		// Add uptime metric
		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_uptime_seconds Total uptime in seconds.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_uptime_seconds gauge\n", p.namespace)
		fmt.Fprintf(w, "%s_uptime_seconds %.3f\n", p.namespace, uptime.Seconds())

		// Add total request count
		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_http_requests_total_all Total requests across all labels.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_requests_total_all counter\n", p.namespace)
		totalRequests := uint64(0)
		for _, count := range requests {
			totalRequests += count
		}
		fmt.Fprintf(w, "%s_http_requests_total_all %d\n", p.namespace, totalRequests)

		if smsSnapshot != nil {
			p.writeSMSGatewayMetrics(w, smsSnapshot)
		}
	})
}

// GetStats returns statistics about the collected metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	// ... use collector ...
//	stats := collector.GetStats()
//	fmt.Printf("Total requests: %d, Average latency: %.3f\n", stats.TotalRequests, stats.AverageLatency)
func (p *PrometheusCollector) GetStats() CollectorStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var stats CollectorStats
	stats.Series = len(p.requests)
	stats.StartTime = p.startTime

	// Calculate request rate
	totalRequests := uint64(0)
	for _, count := range p.requests {
		totalRequests += count
	}
	stats.TotalRequests = totalRequests

	// Calculate average latency
	var totalDuration float64
	var totalSamples uint64
	for _, d := range p.durations {
		totalDuration += d.sum
		totalSamples += d.count
	}
	if totalSamples > 0 {
		stats.AverageLatency = totalDuration / float64(totalSamples)
	}

	return stats
}

// Clear resets all collected metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	// ... use collector ...
//	collector.Clear()
func (p *PrometheusCollector) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.requests = make(map[labelKey]uint64)
	p.durations = make(map[labelKey]latencyStats)
	p.startTime = time.Now()
	p.smsQueueDepth = make(map[smsKey]float64)
	p.smsSendLatency = make(map[smsKey]latencyStats)
	p.smsReceiptDelay = make(map[smsKey]latencyStats)
	p.smsProviderResult = make(map[smsKey]uint64)
	p.smsRetry = make(map[smsKey]uint64)
	p.smsStatus = make(map[smsKey]uint64)

	p.clearBase()
}

// Record implements the unified MetricsCollector interface
func (p *PrometheusCollector) Record(ctx context.Context, record MetricRecord) {
	// For HTTP requests, use the existing Prometheus format
	if record.Type == MetricHTTPRequest {
		if metrics, ok := httpMetricsFromRecord(record); ok {
			p.Observe(ctx, metrics)
			return
		}
	}
	if record.Type == MetricSMSGateway {
		p.observeSMSGateway(record)
		return
	}

	// For other metric types, use the base collector
	p.getBase().Record(ctx, record)
}

// ObserveHTTP implements the unified MetricsCollector interface
func (p *PrometheusCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	metrics := observability.RequestMetrics{
		Method:    method,
		Path:      path,
		Status:    status,
		Bytes:     bytes,
		Duration:  duration,
		TraceID:   "",
		UserAgent: "",
	}
	p.Observe(ctx, metrics)
}

// ObservePubSub, ObserveMQ, ObserveKV, ObserveIPC, ObserveDB are provided
// by the embedded baseForwarder.

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

	uptime := time.Since(p.startTime)
	return reqCopy, durCopy, uptime
}

func (p *PrometheusCollector) evictOldest() {
	// Simple eviction: remove 10% of least-used entries
	evictCount := len(p.requests) / 10
	if evictCount == 0 {
		evictCount = 1
	}

	// Collect all keys
	keys := make([]labelKey, 0, len(p.requests))
	for k := range p.requests {
		keys = append(keys, k)
	}

	// Sort by request count (evict least used first)
	sort.Slice(keys, func(i, j int) bool {
		return p.requests[keys[i]] < p.requests[keys[j]]
	})

	// Remove oldest
	for i := 0; i < evictCount && i < len(keys); i++ {
		delete(p.requests, keys[i])
		delete(p.durations, keys[i])
	}
}


func httpMetricsFromRecord(record MetricRecord) (observability.RequestMetrics, bool) {
	if record.Labels == nil {
		return observability.RequestMetrics{}, false
	}

	statusStr, ok := record.Labels[labelStatus]
	if !ok {
		return observability.RequestMetrics{}, false
	}
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		return observability.RequestMetrics{}, false
	}
	method, ok := record.Labels[labelMethod]
	if !ok {
		return observability.RequestMetrics{}, false
	}
	path, ok := record.Labels[labelPath]
	if !ok {
		return observability.RequestMetrics{}, false
	}

	return observability.RequestMetrics{
		Method:   method,
		Path:     path,
		Status:   status,
		Duration: record.Duration,
	}, true
}

type smsKey struct {
	queue    string
	state    string
	provider string
	tenant   string
	result   string
	status   string
	attempt  string
}

type smsSnapshot struct {
	queueDepth     map[smsKey]float64
	sendLatency    map[smsKey]latencyStats
	receiptDelay   map[smsKey]latencyStats
	providerResult map[smsKey]uint64
	retry          map[smsKey]uint64
	status         map[smsKey]uint64
}

func (p *PrometheusCollector) observeSMSGateway(record MetricRecord) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := smsKey{
		queue:    record.Labels["queue"],
		state:    record.Labels["state"],
		provider: record.Labels["provider"],
		tenant:   record.Labels["tenant"],
		result:   record.Labels["result"],
		status:   record.Labels["status"],
		attempt:  record.Labels["attempt"],
	}

	switch record.Name {
	case "queue_depth":
		if p.maxMemory > 0 && len(p.smsQueueDepth) >= p.maxMemory {
			p.evictOneSMSGauge(p.smsQueueDepth)
		}
		p.smsQueueDepth[key] = record.Value
	case "send_latency":
		p.observeSMSLatency(p.smsSendLatency, key, record)
	case "receipt_delay":
		p.observeSMSLatency(p.smsReceiptDelay, key, record)
	case "provider_result":
		p.observeSMSCounter(&p.smsProviderResult, key, record)
	case "retry":
		p.observeSMSCounter(&p.smsRetry, key, record)
	case "status":
		p.observeSMSCounter(&p.smsStatus, key, record)
	default:
	}
}

func (p *PrometheusCollector) observeSMSLatency(target map[smsKey]latencyStats, key smsKey, record MetricRecord) {
	if p.maxMemory > 0 && len(target) >= p.maxMemory {
		p.evictOneSMSLatency(target)
	}

	duration := record.Duration
	if duration <= 0 {
		if record.Value > 0 {
			duration = time.Duration(record.Value) * time.Millisecond
		}
	}
	seconds := duration.Seconds()

	stats := target[key]
	stats.count++
	stats.sum += seconds
	if stats.count == 1 || seconds < stats.min {
		stats.min = seconds
	}
	if stats.count == 1 || seconds > stats.max {
		stats.max = seconds
	}
	target[key] = stats
}

func (p *PrometheusCollector) observeSMSCounter(target *map[smsKey]uint64, key smsKey, record MetricRecord) {
	if p.maxMemory > 0 && len(*target) >= p.maxMemory {
		p.evictOneSMSCounter(*target)
	}
	inc := uint64(1)
	if record.Value > 0 {
		inc = uint64(record.Value)
	}
	(*target)[key] += inc
}

func (p *PrometheusCollector) snapshotSMSGateway() *smsSnapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.smsQueueDepth) == 0 && len(p.smsSendLatency) == 0 && len(p.smsProviderResult) == 0 && len(p.smsRetry) == 0 && len(p.smsStatus) == 0 && len(p.smsReceiptDelay) == 0 {
		return nil
	}

	copyQueue := make(map[smsKey]float64, len(p.smsQueueDepth))
	for k, v := range p.smsQueueDepth {
		copyQueue[k] = v
	}
	copySend := make(map[smsKey]latencyStats, len(p.smsSendLatency))
	for k, v := range p.smsSendLatency {
		copySend[k] = v
	}
	copyReceipt := make(map[smsKey]latencyStats, len(p.smsReceiptDelay))
	for k, v := range p.smsReceiptDelay {
		copyReceipt[k] = v
	}
	copyProvider := make(map[smsKey]uint64, len(p.smsProviderResult))
	for k, v := range p.smsProviderResult {
		copyProvider[k] = v
	}
	copyRetry := make(map[smsKey]uint64, len(p.smsRetry))
	for k, v := range p.smsRetry {
		copyRetry[k] = v
	}
	copyStatus := make(map[smsKey]uint64, len(p.smsStatus))
	for k, v := range p.smsStatus {
		copyStatus[k] = v
	}

	return &smsSnapshot{
		queueDepth:     copyQueue,
		sendLatency:    copySend,
		receiptDelay:   copyReceipt,
		providerResult: copyProvider,
		retry:          copyRetry,
		status:         copyStatus,
	}
}

func (p *PrometheusCollector) writeSMSGatewayMetrics(w http.ResponseWriter, snap *smsSnapshot) {
	if snap == nil {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "# HELP %s_sms_gateway_queue_depth Current queue depth by state.\n", p.namespace)
	fmt.Fprintf(w, "# TYPE %s_sms_gateway_queue_depth gauge\n", p.namespace)
	for key, val := range snap.queueDepth {
		fmt.Fprintf(w, "%s_sms_gateway_queue_depth%s %.0f\n", p.namespace, formatSMSLabels(key, []string{"queue", "state"}), val)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "# HELP %s_sms_gateway_send_latency_seconds Send latency summary by provider.\n", p.namespace)
	fmt.Fprintf(w, "# TYPE %s_sms_gateway_send_latency_seconds summary\n", p.namespace)
	for key, stats := range snap.sendLatency {
		labels := formatSMSLabels(key, []string{"tenant", "provider"})
		fmt.Fprintf(w, "%s_sms_gateway_send_latency_seconds_sum%s %.9f\n", p.namespace, labels, stats.sum)
		fmt.Fprintf(w, "%s_sms_gateway_send_latency_seconds_count%s %d\n", p.namespace, labels, stats.count)
		fmt.Fprintf(w, "%s_sms_gateway_send_latency_seconds_min%s %.9f\n", p.namespace, labels, stats.min)
		fmt.Fprintf(w, "%s_sms_gateway_send_latency_seconds_max%s %.9f\n", p.namespace, labels, stats.max)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "# HELP %s_sms_gateway_receipt_delay_seconds Receipt delay summary by provider.\n", p.namespace)
	fmt.Fprintf(w, "# TYPE %s_sms_gateway_receipt_delay_seconds summary\n", p.namespace)
	for key, stats := range snap.receiptDelay {
		labels := formatSMSLabels(key, []string{"tenant", "provider"})
		fmt.Fprintf(w, "%s_sms_gateway_receipt_delay_seconds_sum%s %.9f\n", p.namespace, labels, stats.sum)
		fmt.Fprintf(w, "%s_sms_gateway_receipt_delay_seconds_count%s %d\n", p.namespace, labels, stats.count)
		fmt.Fprintf(w, "%s_sms_gateway_receipt_delay_seconds_min%s %.9f\n", p.namespace, labels, stats.min)
		fmt.Fprintf(w, "%s_sms_gateway_receipt_delay_seconds_max%s %.9f\n", p.namespace, labels, stats.max)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "# HELP %s_sms_gateway_provider_result_total Provider send results.\n", p.namespace)
	fmt.Fprintf(w, "# TYPE %s_sms_gateway_provider_result_total counter\n", p.namespace)
	for key, val := range snap.providerResult {
		fmt.Fprintf(w, "%s_sms_gateway_provider_result_total%s %d\n", p.namespace, formatSMSLabels(key, []string{"tenant", "provider", "result"}), val)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "# HELP %s_sms_gateway_retry_total Retry attempts by provider.\n", p.namespace)
	fmt.Fprintf(w, "# TYPE %s_sms_gateway_retry_total counter\n", p.namespace)
	for key, val := range snap.retry {
		fmt.Fprintf(w, "%s_sms_gateway_retry_total%s %d\n", p.namespace, formatSMSLabels(key, []string{"tenant", "provider", "attempt"}), val)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "# HELP %s_sms_gateway_status_total Message status transitions.\n", p.namespace)
	fmt.Fprintf(w, "# TYPE %s_sms_gateway_status_total counter\n", p.namespace)
	for key, val := range snap.status {
		fmt.Fprintf(w, "%s_sms_gateway_status_total%s %d\n", p.namespace, formatSMSLabels(key, []string{"tenant", "status"}), val)
	}
}

func (p *PrometheusCollector) evictOneSMSGauge(target map[smsKey]float64) {
	for key := range target {
		delete(target, key)
		return
	}
}

func (p *PrometheusCollector) evictOneSMSLatency(target map[smsKey]latencyStats) {
	for key := range target {
		delete(target, key)
		return
	}
}

func (p *PrometheusCollector) evictOneSMSCounter(target map[smsKey]uint64) {
	for key := range target {
		delete(target, key)
		return
	}
}

func formatSMSLabels(key smsKey, order []string) string {
	if len(order) == 0 {
		return ""
	}
	var parts []string
	for _, label := range order {
		value := ""
		switch label {
		case "queue":
			value = key.queue
		case "state":
			value = key.state
		case "provider":
			value = key.provider
		case "tenant":
			value = key.tenant
		case "result":
			value = key.result
		case "status":
			value = key.status
		case "attempt":
			value = key.attempt
		}
		if value == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", label, escapeLabelValue(value)))
	}
	if len(parts) == 0 {
		return ""
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func escapeLabelValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return value
}

func sortedKeys[T any](m map[labelKey]T) []labelKey {
	keys := make([]labelKey, 0, len(m))
	for k := range m {
		keys = append(keys, k)
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
