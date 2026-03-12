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
//	exporter := metrics.NewPrometheusExporter(collector)
//	mux := http.NewServeMux()
//	mux.Handle("/metrics", exporter.Handler())
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

func (p *PrometheusCollector) recordHTTP(method, path string, status int, duration time.Duration) {
	key := labelKey{method, path, strconv.Itoa(status)}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.maxMemory > 0 && len(p.requests) >= p.maxMemory {
		p.evictLeastUsed()
	}

	p.requests[key]++

	d := duration.Seconds()
	stats := p.durations[key]
	stats.count++
	stats.sum += d
	if stats.count == 1 || d < stats.min {
		stats.min = d
	}
	if stats.count == 1 || d > stats.max {
		stats.max = d
	}
	p.durations[key] = stats
}

// Handler returns a compatibility HTTP handler for Prometheus scraping.
// Prefer constructing an explicit exporter with NewPrometheusExporter.
func (p *PrometheusCollector) Handler() http.Handler {
	return NewPrometheusExporter(p).Handler()
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
//	fmt.Printf("Total records: %d, Active series: %d\n", stats.TotalRecords, stats.ActiveSeries)
func (p *PrometheusCollector) GetStats() CollectorStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	totalRequests := uint64(0)
	errorRecords := int64(0)
	for key, count := range p.requests {
		totalRequests += count
		status, err := strconv.Atoi(key.status)
		if err == nil && status >= http.StatusBadRequest {
			errorRecords += int64(count)
		}
	}

	stats := CollectorStats{
		TotalRecords:  int64(totalRequests),
		ErrorRecords:  errorRecords,
		ActiveSeries:  len(p.requests),
		StartTime:     p.startTime,
		TypeBreakdown: make(map[MetricType]int64),
	}
	if totalRequests > 0 {
		stats.TypeBreakdown[MetricHTTPRequest] = int64(totalRequests)
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

// Record implements the unified MetricsCollector interface.
func (p *PrometheusCollector) Record(ctx context.Context, record MetricRecord) {
	if record.Type == MetricHTTPRequest {
		method, _ := record.Labels[labelMethod]
		path, _ := record.Labels[labelPath]
		statusStr, _ := record.Labels[labelStatus]
		status, _ := strconv.Atoi(statusStr)
		if method != "" && path != "" && status != 0 {
			p.recordHTTP(method, path, status, record.Duration)
			return
		}
	}
	if record.Type == MetricSMSGateway {
		p.observeSMSGateway(record)
		return
	}
	p.getBase().Record(ctx, record)
}

// ObserveHTTP implements the unified MetricsCollector interface.
func (p *PrometheusCollector) ObserveHTTP(_ context.Context, method, path string, status, _ int, duration time.Duration) {
	p.recordHTTP(method, path, status, duration)
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

// evictLeastUsed removes the least-requested 10% of label combinations to stay within
// the maxMemory limit. Eviction is by ascending request count (least used first).
func (p *PrometheusCollector) evictLeastUsed() {
	evictCount := len(p.requests) / 10
	if evictCount == 0 {
		evictCount = 1
	}

	keys := make([]labelKey, 0, len(p.requests))
	for k := range p.requests {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return p.requests[keys[i]] < p.requests[keys[j]]
	})

	for i := 0; i < evictCount && i < len(keys); i++ {
		delete(p.requests, keys[i])
		delete(p.durations, keys[i])
	}
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
	if duration <= 0 && record.Value > 0 {
		seconds := record.Value
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
		return
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
