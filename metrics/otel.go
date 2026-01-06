package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
)

// Span captures finalized tracing data for inspection or export.
type Span struct {
	Name          string            `json:"name"`
	Attributes    map[string]string `json:"attributes,omitempty"`
	Status        string            `json:"status"`
	StatusMessage string            `json:"status_message,omitempty"`
	Duration      time.Duration     `json:"duration,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	ParentSpanID  string            `json:"parent_span_id,omitempty"`
	SpanID        string            `json:"span_id"`
	TraceID       string            `json:"trace_id"`
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

// OpenTelemetryTracer implements middleware.Tracer without external dependencies.
type OpenTelemetryTracer struct {
	name string

	mu    sync.RWMutex
	spans []Span
}

// NewOpenTelemetryTracer creates a tracer with the given instrumentation name.
// An empty name defaults to the plumego metrics namespace.
func NewOpenTelemetryTracer(name string) *OpenTelemetryTracer {
	if name == "" {
		name = "github.com/spcent/plumego/metrics"
	}
	return &OpenTelemetryTracer{name: name}
}

// Start begins a new span for the incoming request.
func (t *OpenTelemetryTracer) Start(ctx context.Context, r *http.Request) (context.Context, middleware.TraceSpan) {
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
func (t *OpenTelemetryTracer) Spans() []Span {
	t.mu.RLock()
	defer t.mu.RUnlock()

	spans := make([]Span, len(t.spans))
	copy(spans, t.spans)
	return spans
}

// GetStats returns statistics about collected spans.
func (t *OpenTelemetryTracer) GetStats() SpanStats {
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

// Clear clears all collected spans.
func (t *OpenTelemetryTracer) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = t.spans[:0]
}

func (t *OpenTelemetryTracer) record(span Span) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = append(t.spans, span)
}

// End finalizes the span and records its metrics attributes.
func (s *spanHandle) End(metrics middleware.RequestMetrics) {
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

func (s *spanHandle) TraceID() string {
	return s.traceID
}

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
type SpanStats struct {
	TotalSpans      int
	ErrorSpans      int
	TotalDuration   time.Duration
	AverageDuration time.Duration
}
