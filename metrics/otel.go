package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware/observability"
)

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

// OpenTelemetryTracer implements observability.Tracer without external dependencies.
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
//		defer span.End(metrics.RequestMetrics{
//			Status:  http.StatusOK,
//			Bytes:   100,
//			TraceID: span.TraceID(),
//		})
//		// ... handle request ...
//	})
type OpenTelemetryTracer struct {
	baseForwarder // provides Record, ObserveHTTP, ObservePubSub, ObserveMQ, ObserveKV, ObserveIPC, ObserveDB

	name string

	mu    sync.RWMutex
	spans []Span
}

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
	return &OpenTelemetryTracer{name: name}
}

// Start begins a new span for the incoming request.
//
// This method:
//   - Generates a unique span ID and trace ID
//   - Extracts trace context from headers or context
//   - Creates a span with HTTP-specific attributes
//   - Returns a context with the trace information
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	tracer := metrics.NewOpenTelemetryTracer("my-service")
//	ctx, span := tracer.Start(context.Background(), r)
//	defer span.End(metrics.RequestMetrics{
//		Status:  http.StatusOK,
//		Bytes:   100,
//		TraceID: span.TraceID(),
//	})
func (t *OpenTelemetryTracer) Start(ctx context.Context, r *http.Request) (context.Context, observability.TraceSpan) {
	// Generate simple span and trace IDs
	spanID := generateSpanID()
	traceID := contract.TraceIDFromContext(ctx)
	if traceID == "" {
		if headerTraceID := r.Header.Get("X-Trace-ID"); headerTraceID != "" {
			traceID = headerTraceID
		} else {
			traceID = generateTraceID()
		}
	}

	handle := &spanHandle{
		tracer:    t,
		name:      "http.request",
		startTime: time.Now(),
		spanID:    spanID,
		traceID:   traceID,
		attrs: map[string]string{
			"http.method":     r.Method,
			"http.route":      r.URL.Path,
			"http.user_agent": r.UserAgent(),
			"http.scheme":     "http",
			"net.peer.name":   r.Host,
			"net.transport":   "tcp",
			"service.name":    t.name,
			"service.version": "1.0.0", // Could be configurable
		},
	}

	// Track parent span when trace context is available.
	parentSpanID := ""
	if parent := contract.TraceContextFromContext(ctx); parent != nil && parent.SpanID != "" {
		parentSpanID = string(parent.SpanID)
		handle.parentID = parentSpanID
		handle.attrs["parent.span_id"] = parentSpanID
	}

	// Preserve parent trace ID from inbound headers for compatibility.
	if parentID := r.Header.Get("X-Trace-ID"); parentID != "" {
		handle.attrs["parent.trace_id"] = parentID
	}

	traceCtx := contract.TraceContext{
		TraceID: contract.TraceID(traceID),
		SpanID:  contract.SpanID(spanID),
	}
	if parentSpanID != "" {
		parentID := contract.SpanID(parentSpanID)
		traceCtx.ParentSpanID = &parentID
	}
	ctx = contract.ContextWithTraceContext(ctx, traceCtx)

	return ctx, handle
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
	t.spans = append(t.spans, span)
}

// End finalizes the span and records its metrics attributes.
//
// This method:
//   - Calculates the span duration
//   - Adds HTTP-specific attributes
//   - Determines the span status based on HTTP status code
//   - Records the span for later inspection
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	span.End(metrics.RequestMetrics{
//		Status:  http.StatusOK,
//		Bytes:   100,
//		TraceID: span.TraceID(),
//	})
func (s *spanHandle) End(metrics observability.RequestMetrics) {
	duration := time.Since(s.startTime)

	// Build attributes
	attrs := make(map[string]string, len(s.attrs)+5)
	for k, v := range s.attrs {
		attrs[k] = v
	}

	attrs["http.status_code"] = strconv.Itoa(metrics.Status)
	attrs["http.response_content_length"] = strconv.Itoa(metrics.Bytes)
	attrs["plumego.trace_id"] = metrics.TraceID
	attrs["http.status_text"] = http.StatusText(metrics.Status)
	attrs["duration_ms"] = strconv.FormatInt(duration.Milliseconds(), 10)

	// Determine status and message
	status := "OK"
	statusMsg := ""
	if metrics.Status >= 500 {
		status = "ERROR"
		statusMsg = fmt.Sprintf("Server error: %d", metrics.Status)
	} else if metrics.Status >= 400 {
		status = "ERROR"
		statusMsg = fmt.Sprintf("Client error: %d", metrics.Status)
	}

	s.tracer.record(Span{
		Name:          s.name,
		Attributes:    attrs,
		Status:        status,
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

// SpanID and TraceID generation (simple implementation)
var (
	spanCounter  uint64
	traceCounter uint64
	spanMu       sync.Mutex
)

func generateSpanID() string {
	spanMu.Lock()
	defer spanMu.Unlock()

	spanCounter++
	// Use nanosecond timestamp + counter to ensure uniqueness
	return fmt.Sprintf("%x-%x", time.Now().UnixNano(), spanCounter)
}

func generateTraceID() string {
	spanMu.Lock()
	defer spanMu.Unlock()

	traceCounter++
	// Use nanosecond timestamp + counter to ensure uniqueness
	return fmt.Sprintf("%x-%x-%x", time.Now().Unix(), time.Now().UnixNano(), traceCounter)
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
}

// Record, ObserveHTTP, ObservePubSub, ObserveMQ, ObserveKV, ObserveIPC, ObserveDB
// are provided by the embedded baseForwarder.

// GetStats implements the unified MetricsCollector interface
// This method returns span statistics
func (t *OpenTelemetryTracer) GetStats() CollectorStats {
	spanStats := t.GetSpanStats()

	return CollectorStats{
		TotalSpans:      spanStats.TotalSpans,
		ErrorSpans:      spanStats.ErrorSpans,
		TotalDuration:   spanStats.TotalDuration,
		AverageDuration: spanStats.AverageDuration,
		// Include base collector stats if available
		TotalRecords: func() int64 {
			if t.baseForwarder.base != nil {
				return t.baseForwarder.base.GetStats().TotalRecords
			}
			return 0
		}(),
		ErrorRecords: func() int64 {
			if t.baseForwarder.base != nil {
				return t.baseForwarder.base.GetStats().ErrorRecords
			}
			return 0
		}(),
	}
}

// Clear implements the unified MetricsCollector interface
func (t *OpenTelemetryTracer) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = t.spans[:0]

	t.clearBase()
}
