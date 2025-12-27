package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/spcent/plumego/middleware"
)

type labelKey struct {
	method string
	path   string
	status string
}

type latencyStats struct {
	count uint64
	sum   float64
}

// PrometheusCollector implements middleware.MetricsCollector without third-party dependencies.
// It exposes a text-based metrics handler compatible with Prometheus exposition format.
type PrometheusCollector struct {
	namespace string

	mu        sync.Mutex
	requests  map[labelKey]uint64
	durations map[labelKey]latencyStats
}

// NewPrometheusCollector constructs an in-memory collector with the provided namespace.
// An empty namespace defaults to "plumego".
func NewPrometheusCollector(namespace string) *PrometheusCollector {
	if namespace == "" {
		namespace = "plumego"
	}
	return &PrometheusCollector{
		namespace: namespace,
		requests:  make(map[labelKey]uint64),
		durations: make(map[labelKey]latencyStats),
	}
}

// Observe records a single HTTP request metric set.
func (p *PrometheusCollector) Observe(_ context.Context, metrics middleware.RequestMetrics) {
	key := labelKey{metrics.Method, metrics.Path, strconv.Itoa(metrics.Status)}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.requests[key]++
	stats := p.durations[key]
	stats.count++
	stats.sum += metrics.Duration.Seconds()
	p.durations[key] = stats
}

// Handler returns an HTTP handler that emits the current metrics snapshot.
func (p *PrometheusCollector) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests, durations := p.snapshot()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP %s_http_requests_total Total number of HTTP requests processed.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_requests_total counter\n", p.namespace)

		reqKeys := sortedKeys(requests)
		for _, k := range reqKeys {
			fmt.Fprintf(w, "%s_http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				p.namespace, k.method, k.path, k.status, requests[k])
		}

		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_http_request_duration_seconds Sum of HTTP request latencies in seconds.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_request_duration_seconds summary\n", p.namespace)

		durKeys := sortedKeys(durations)
		for _, k := range durKeys {
			stats := durations[k]
			fmt.Fprintf(w, "%s_http_request_duration_seconds_sum{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				p.namespace, k.method, k.path, k.status, stats.sum)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_count{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				p.namespace, k.method, k.path, k.status, stats.count)
		}
	})
}

func (p *PrometheusCollector) snapshot() (map[labelKey]uint64, map[labelKey]latencyStats) {
	p.mu.Lock()
	defer p.mu.Unlock()

	reqCopy := make(map[labelKey]uint64, len(p.requests))
	for k, v := range p.requests {
		reqCopy[k] = v
	}

	durCopy := make(map[labelKey]latencyStats, len(p.durations))
	for k, v := range p.durations {
		durCopy[k] = v
	}
	return reqCopy, durCopy
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
