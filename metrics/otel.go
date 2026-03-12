package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// TraceSpan represents an in-flight tracing span returned by OpenTelemetryTracer.Start.
// Callers end the span by calling End with the final HTTP outcome.
type TraceSpan interface {
	// End finalizes the span with the HTTP response outcome.
	End(status, bytes int, traceID string)

	// TraceID returns the trace identifier propagated through this span.
	TraceID() string

	// SpanID returns the unique identifier of this span.
	SpanID() string
}

// traceContextKey is the package-internal context key for span propagation.
type traceContextKey struct{}

// internalTraceContext carries trace identity within the metrics package.
type internalTraceContext struct {
	traceID string
	spanID  string
}

// Span captures finalized tracing data for inspection or export.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	span := metrics.Span{
//		Name:       "http.request",
//		TraceID:    "trace-123",
//		SpanID:     "span-456",
//		Status:     "OK",
//		Duration:   100 * time.Millisecond,
//		Attributes: map[string]string{
//			"http.method": "GET",
//			"http.route":  "/api/users",
//		},
//	}
type Span struct {
	// Name is the name of the span
	Name string `json:"name"`

	// Attributes are key-value pairs associated with the span
	Attributes map[string]string `json:"attributes,omitempty"`

	// Status is the status of the span (OK, ERROR, etc.)
	Status string `json:"status"`

	// StatusMessage provides additional context for the status
	StatusMessage string `json:"status_message,omitempty"`

	// Duration is the duration of the span
	Duration time.Duration `json:"duration,omitempty"`

	// Timestamp is when the span started
	Timestamp time.Time `json:"timestamp"`

	// ParentSpanID is the ID of the parent span
	ParentSpanID string `json:"parent_span_id,omitempty"`

	// SpanID is the unique identifier for this span
	SpanID string `json:"span_id"`

	// TraceID is the unique identifier for this trace
	TraceID string `json:"trace_id"`
}

type spanHandle struct {
	tracer    *OpenTelemetryTracer
	name      string
	attrs     map[string]string
	startTime time.Time
	spanID    string
	traceID   string
	parentID  string
}

// OpenTelemetryTracer implements distributed tracing without external dependencies.
//
// This tracer provides OpenTelemetry-compatible tracing without requiring
// external dependencies. It's designed for:
//   - HTTP request tracing
//   - Distributed tracing
//   - Performance monitoring
//   - Error tracking
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	tracer := metrics.NewOpenTelemetryTracer("my-service")
//	defer tracer.Clear()
//
//	// Use in HTTP middleware
//	mux := http.NewServeMux()
//	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//		ctx, span := tracer.Start(context.Background(), r)
//		defer span.End(http.StatusOK, 100, span.TraceID())
//		// ... handle request ...
//	})
type OpenTelemetryTracer struct {
	baseForwarder // provides Record, ObserveHTTP, ObservePubSub, ObserveMQ, ObserveKV, ObserveIPC, ObserveDB

	name string

	mu           sync.RWMutex
	spans        []Span
	maxSpans     int
	droppedSpan  int
	spanCounter  uint64
	traceCounter uint64
}

const defaultMaxSpans = 10000

// NewOpenTelemetryTracer creates a tracer with the given instrumentation name.
// An empty name defaults to the plumego metrics namespace.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	tracer := metrics.NewOpenTelemetryTracer("my-service")
func NewOpenTelemetryTracer(name string) *OpenTelemetryTracer {
	if name == "" {
		name = "github.com/spcent/plumego/metrics"
	}
	return &OpenTelemetryTracer{
		name:     name,
		maxSpans: defaultMaxSpans,
	}
}

// WithMaxSpans limits how many spans are retained in memory.
// A non-positive value disables the retention limit.
func (t *OpenTelemetryTracer) WithMaxSpans(max int) *OpenTelemetryTracer {
	t.mu.Lock()
	defer t.mu.Unlock()

	if max <= 0 {
		t.maxSpans = 0
		return t
	}

	t.maxSpans = max
	if len(t.spans) > max {
		t.droppedSpan += len(t.spans) - max
		t.spans = t.spans[len(t.spans)-max:]
	}

	return t
}

// Start begins a new span for the incoming request.
//
// It reads X-Trace-ID from the request header to propagate an existing trace,
// and extracts any parent span context from the context value set by a prior
// Start call in the same request chain.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	tracer := metrics.NewOpenTelemetryTracer("my-service")
//	ctx, span := tracer.Start(context.Background(), r)
//	defer span.End(http.StatusOK, 100, span.TraceID())
func (t *OpenTelemetryTracer) Start(ctx context.Context, r *http.Request) (context.Context, TraceSpan) {
	spanID := t.generateSpanID()

	// Resolve trace ID: prefer inbound header, then parent context, then generate new.
	traceID := r.Header.Get("X-Trace-ID")
	if traceID == "" {
		if parent := traceContextFromContext(ctx); parent != nil {
			traceID = parent.traceID
		}
	}
	if traceID == "" {
		traceID = t.generateTraceID()
	}

	parentID := ""
	if parent := traceContextFromContext(ctx); parent != nil && parent.spanID != "" {
		parentID = parent.spanID
	}

	handle := &spanHandle{
		tracer:    t,
		name:      "http.request",
		startTime: time.Now(),
		spanID:    spanID,
		traceID:   traceID,
		parentID:  parentID,
		attrs: map[string]string{
			"http.method":     r.Method,
			"http.route":      r.URL.Path,
			"http.user_agent": r.UserAgent(),
			"http.scheme":     "http",
			"net.peer.name":   r.Host,
			"net.transport":   "tcp",
			"service.name":    t.name,
		},
	}
	if parentID != "" {
		handle.attrs["parent.span_id"] = parentID
	}
	if inboundTraceID := r.Header.Get("X-Trace-ID"); inboundTraceID != "" {
		handle.attrs["parent.trace_id"] = inboundTraceID
	}

	ctx = context.WithValue(ctx, traceContextKey{}, &internalTraceContext{
		traceID: traceID,
		spanID:  spanID,
	})

	return ctx, handle
}

func traceContextFromContext(ctx context.Context) *internalTraceContext {
	v, _ := ctx.Value(traceContextKey{}).(*internalTraceContext)
	return v
}

// Spans returns a copy of completed spans.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	tracer := metrics.NewOpenTelemetryTracer("my-service")
//	// ... use tracer ...
//	spans := tracer.Spans()
//	for _, span := range spans {
//		fmt.Printf("Span: %s, Duration: %v\n", span.Name, span.Duration)
//	}
func (t *OpenTelemetryTracer) Spans() []Span {
	t.mu.RLock()
	defer t.mu.RUnlock()

	spans := make([]Span, len(t.spans))
	copy(spans, t.spans)
	return spans
}

// GetSpanStats returns statistics about collected spans.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	tracer := metrics.NewOpenTelemetryTracer("my-service")
//	// ... use tracer ...
//	stats := tracer.GetSpanStats()
//	fmt.Printf("Total spans: %d, Error spans: %d\n", stats.TotalSpans, stats.ErrorSpans)
func (t *OpenTelemetryTracer) GetSpanStats() SpanStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var stats SpanStats
	stats.TotalSpans = len(t.spans)
	stats.MaxRetention = t.maxSpans
	stats.DroppedSpans = t.droppedSpan

	for _, span := range t.spans {
		if span.Status == "ERROR" {
			stats.ErrorSpans++
		}
		stats.TotalDuration += span.Duration
	}

	if stats.TotalSpans > 0 {
		stats.AverageDuration = stats.TotalDuration / time.Duration(stats.TotalSpans)
	}

	return stats
}

func (t *OpenTelemetryTracer) record(span Span) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.maxSpans > 0 && len(t.spans) >= t.maxSpans {
		t.spans = t.spans[1:]
		t.droppedSpan++
	}

	t.spans = append(t.spans, span)
}

// End finalizes the span and records its metrics attributes.
//
// Example:
//
//	ctx, span := tracer.Start(r.Context(), r)
//	defer span.End(http.StatusOK, bytesWritten, span.TraceID())
func (s *spanHandle) End(status, bytes int, traceID string) {
	duration := time.Since(s.startTime)

	attrs := make(map[string]string, len(s.attrs)+4)
	for k, v := range s.attrs {
		attrs[k] = v
	}

	attrs["http.status_code"] = strconv.Itoa(status)
	attrs["http.response_content_length"] = strconv.Itoa(bytes)
	attrs["duration_ms"] = strconv.FormatInt(duration.Milliseconds(), 10)
	attrs["http.status_text"] = http.StatusText(status)
	if traceID != "" {
		attrs["plumego.trace_id"] = traceID
	}

	spanStatus := "OK"
	statusMsg := ""
	if status >= 500 {
		spanStatus = "ERROR"
		statusMsg = fmt.Sprintf("Server error: %d", status)
	} else if status >= 400 {
		spanStatus = "ERROR"
		statusMsg = fmt.Sprintf("Client error: %d", status)
	}

	s.tracer.record(Span{
		Name:          s.name,
		Attributes:    attrs,
		Status:        spanStatus,
		StatusMessage: statusMsg,
		Duration:      duration,
		Timestamp:     s.startTime,
		SpanID:        s.spanID,
		TraceID:       s.traceID,
		ParentSpanID:  s.parentID,
	})
}

// TraceID returns the trace ID of the span.
func (s *spanHandle) TraceID() string {
	return s.traceID
}

// SpanID returns the span ID of the span.
func (s *spanHandle) SpanID() string {
	return s.spanID
}

func (t *OpenTelemetryTracer) generateSpanID() string {
	n := atomic.AddUint64(&t.spanCounter, 1)
	return fmt.Sprintf("%x-%x", time.Now().UnixNano(), n)
}

func (t *OpenTelemetryTracer) generateTraceID() string {
	n := atomic.AddUint64(&t.traceCounter, 1)
	return fmt.Sprintf("%x-%x-%x", time.Now().Unix(), time.Now().UnixNano(), n)
}

// SpanStats provides statistics about collected spans.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	tracer := metrics.NewOpenTelemetryTracer("my-service")
//	// ... use tracer ...
//	stats := tracer.GetSpanStats()
//	fmt.Printf("Average duration: %v\n", stats.AverageDuration)
type SpanStats struct {
	TotalSpans      int
	ErrorSpans      int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MaxRetention    int
	DroppedSpans    int
}

// Record, ObserveHTTP, ObservePubSub, ObserveMQ, ObserveKV, ObserveIPC, ObserveDB
// are provided by the embedded baseForwarder.

// GetStats implements the aggregate collector contract.
func (t *OpenTelemetryTracer) GetStats() CollectorStats {
	spanStats := t.GetSpanStats()

	stats := CollectorStats{
		TotalRecords:  int64(spanStats.TotalSpans),
		ErrorRecords:  int64(spanStats.ErrorSpans),
		ActiveSeries:  map[bool]int{true: 1, false: 0}[spanStats.TotalSpans > 0],
		TypeBreakdown: make(map[MetricType]int64),
	}

	if spanStats.TotalSpans > 0 {
		stats.TypeBreakdown[MetricHTTPRequest] = int64(spanStats.TotalSpans)
	}

	if t.baseForwarder.base != nil {
		baseStats := t.baseForwarder.base.GetStats()
		stats.TotalRecords += baseStats.TotalRecords
		stats.ErrorRecords += baseStats.ErrorRecords
		if len(baseStats.TypeBreakdown) > 0 {
			if stats.TypeBreakdown == nil {
				stats.TypeBreakdown = make(map[MetricType]int64, len(baseStats.TypeBreakdown))
			}
			for key, value := range baseStats.TypeBreakdown {
				stats.TypeBreakdown[key] += value
			}
			stats.ActiveSeries = len(stats.TypeBreakdown)
		}
		if !baseStats.StartTime.IsZero() {
			stats.StartTime = baseStats.StartTime
		}
	}

	return stats
}

// Clear implements the aggregate collector contract.
func (t *OpenTelemetryTracer) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = t.spans[:0]
	t.droppedSpan = 0

	t.clearBase()
}
