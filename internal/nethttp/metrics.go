package nethttp

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// ClientMetrics is the interface for recording outbound HTTP client request metrics.
//
// Implement this interface to forward client metrics to any backend:
// Prometheus, OpenTelemetry, StatsD, or the built-in InMemoryClientMetrics.
//
// Example – plug in the project's unified collector:
//
//	import (
//	    nethttp  "github.com/spcent/plumego/internal/nethttp"
//	    "github.com/spcent/plumego/x/observability"
//	)
//
//	col := observability.NewPrometheusCollector("myapp")
//
//	client := nethttp.New(
//	    nethttp.WithMiddleware(nethttp.MetricsMiddleware(&prometheusAdapter{col})),
//	)
type ClientMetrics interface {
	// RecordRequest is called once after every completed attempt (success or final failure).
	//
	//   method     – HTTP method ("GET", "POST", …)
	//   host       – target host[:port] (no path)
	//   statusCode – HTTP status code; 0 when the request failed before receiving a response
	//   duration   – wall-clock time from the start of the attempt to completion
	//   err        – non-nil when the attempt ended with a transport-level error
	RecordRequest(ctx context.Context, method, host string, statusCode int, duration time.Duration, err error)

	// IncInflight increments the in-flight request gauge for the given method and host.
	// Called immediately before the request is dispatched.
	IncInflight(method, host string)

	// DecInflight decrements the in-flight request gauge for the given method and host.
	// Called immediately after the response (or error) is received.
	DecInflight(method, host string)
}

// ClientMetricsSnapshot is a point-in-time view of the metrics gathered by InMemoryClientMetrics.
type ClientMetricsSnapshot struct {
	// TotalRequests is the total number of completed requests (all outcomes).
	TotalRequests int64
	// ErrorRequests is the number of requests that ended with a transport-level error
	// (i.e., err != nil, which includes connection failures, timeouts, and context cancellations).
	ErrorRequests int64
	// Inflight is the current number of in-flight requests.
	Inflight int64
	// TotalDuration is the cumulative duration of all completed requests.
	TotalDuration time.Duration
	// AverageDuration is TotalDuration / TotalRequests. Zero when TotalRequests == 0.
	AverageDuration time.Duration
	// ByStatus is a histogram of response status codes (key 0 = transport error).
	ByStatus map[int]int64
	// ByMethod is a counter per HTTP method.
	ByMethod map[string]int64
	// ByHost is a counter per target host.
	ByHost map[string]int64
}

// InMemoryClientMetrics is a thread-safe, zero-dependency implementation of ClientMetrics
// that keeps all metrics in process memory.
//
// It is suitable for:
//   - Integration tests that assert on request counts and status codes
//   - Low-traffic services that want a lightweight metrics solution
//   - Health-check endpoints that expose live stats
//
// For high-cardinality production use, implement ClientMetrics using your existing
// metrics backend (Prometheus, OpenTelemetry, etc.).
//
// Example:
//
//	m := nethttp.NewInMemoryClientMetrics()
//	client := nethttp.New(
//	    nethttp.WithMiddleware(nethttp.MetricsMiddleware(m)),
//	)
//
//	// Later, inspect metrics:
//	snap := m.Snapshot()
//	fmt.Printf("total=%d errors=%d avg=%s\n",
//	    snap.TotalRequests, snap.ErrorRequests, snap.AverageDuration)
type InMemoryClientMetrics struct {
	// Atomic counters for hot-path, contention-free updates.
	total    atomic.Int64
	errors   atomic.Int64
	inflight atomic.Int64
	totalNs  atomic.Int64 // nanoseconds, to allow atomic adds

	mu       sync.RWMutex
	byStatus map[int]int64
	byMethod map[string]int64
	byHost   map[string]int64
}

// NewInMemoryClientMetrics creates a ready-to-use in-memory metrics recorder.
func NewInMemoryClientMetrics() *InMemoryClientMetrics {
	return &InMemoryClientMetrics{
		byStatus: make(map[int]int64),
		byMethod: make(map[string]int64),
		byHost:   make(map[string]int64),
	}
}

// RecordRequest implements ClientMetrics.
func (m *InMemoryClientMetrics) RecordRequest(_ context.Context, method, host string, statusCode int, duration time.Duration, err error) {
	m.total.Add(1)
	m.totalNs.Add(int64(duration))
	if err != nil {
		m.errors.Add(1)
	}

	m.mu.Lock()
	m.byStatus[statusCode]++
	m.byMethod[method]++
	m.byHost[host]++
	m.mu.Unlock()
}

// IncInflight implements ClientMetrics.
func (m *InMemoryClientMetrics) IncInflight(_, _ string) {
	m.inflight.Add(1)
}

// DecInflight implements ClientMetrics.
func (m *InMemoryClientMetrics) DecInflight(_, _ string) {
	m.inflight.Add(-1)
}

// Snapshot returns a consistent point-in-time copy of all collected metrics.
// The returned maps are safe to read and modify independently.
func (m *InMemoryClientMetrics) Snapshot() ClientMetricsSnapshot {
	total := m.total.Load()
	errCount := m.errors.Load()
	inflight := m.inflight.Load()
	totalNs := m.totalNs.Load()

	var avgDur time.Duration
	if total > 0 {
		avgDur = time.Duration(totalNs / total)
	}

	m.mu.RLock()
	byStatus := cloneIntMap(m.byStatus)
	byMethod := cloneStringMap(m.byMethod)
	byHost := cloneStringMap(m.byHost)
	m.mu.RUnlock()

	return ClientMetricsSnapshot{
		TotalRequests:   total,
		ErrorRequests:   errCount,
		Inflight:        inflight,
		TotalDuration:   time.Duration(totalNs),
		AverageDuration: avgDur,
		ByStatus:        byStatus,
		ByMethod:        byMethod,
		ByHost:          byHost,
	}
}

// Reset clears all accumulated metrics. Useful between test cases.
func (m *InMemoryClientMetrics) Reset() {
	m.total.Store(0)
	m.errors.Store(0)
	m.inflight.Store(0)
	m.totalNs.Store(0)

	m.mu.Lock()
	m.byStatus = make(map[int]int64)
	m.byMethod = make(map[string]int64)
	m.byHost = make(map[string]int64)
	m.mu.Unlock()
}

// MetricsMiddleware returns a Middleware that records outbound HTTP client metrics
// using the provided ClientMetrics implementation.
//
// For each request the middleware:
//  1. Calls IncInflight before dispatching
//  2. Calls DecInflight after the response (or error) is received
//  3. Calls RecordRequest with the final outcome
//
// The middleware is transparent: it never modifies the request or response.
//
// Example:
//
//	m := nethttp.NewInMemoryClientMetrics()
//	client := nethttp.New(
//	    nethttp.WithMiddleware(nethttp.MetricsMiddleware(m)),
//	    nethttp.WithMiddleware(nethttp.Logging),
//	)
func MetricsMiddleware(recorder ClientMetrics) Middleware {
	return func(next RoundTripperFunc) RoundTripperFunc {
		if recorder == nil {
			return next
		}
		return func(req *http.Request) (*http.Response, error) {
			method := req.Method
			host := req.URL.Host

			recorder.IncInflight(method, host)
			start := time.Now()

			resp, err := next(req)

			dur := time.Since(start)
			recorder.DecInflight(method, host)

			statusCode := 0
			if resp != nil {
				statusCode = resp.StatusCode
			}
			recorder.RecordRequest(req.Context(), method, host, statusCode, dur, err)

			return resp, err
		}
	}
}

// cloneIntMap returns a shallow copy of a map[int]int64.
func cloneIntMap(src map[int]int64) map[int]int64 {
	dst := make(map[int]int64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// cloneStringMap returns a shallow copy of a map[string]int64.
func cloneStringMap(src map[string]int64) map[string]int64 {
	dst := make(map[string]int64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
