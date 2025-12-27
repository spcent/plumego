package metrics

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/spcent/plumego/middleware"
)

// Span captures finalized tracing data for inspection or export.
type Span struct {
	Name          string
	Attributes    map[string]string
	Status        string
	StatusMessage string
}

type spanHandle struct {
	tracer *OpenTelemetryTracer
	name   string
	attrs  map[string]string
}

// OpenTelemetryTracer implements middleware.Tracer without external dependencies.
type OpenTelemetryTracer struct {
	name string

	mu    sync.Mutex
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
	handle := &spanHandle{
		tracer: t,
		name:   "http.request",
		attrs: map[string]string{
			"http.method":     r.Method,
			"http.route":      r.URL.Path,
			"http.user_agent": r.UserAgent(),
		},
	}
	return ctx, handle
}

// Spans returns a copy of completed spans.
func (t *OpenTelemetryTracer) Spans() []Span {
	t.mu.Lock()
	defer t.mu.Unlock()

	spans := make([]Span, len(t.spans))
	copy(spans, t.spans)
	return spans
}

func (t *OpenTelemetryTracer) record(span Span) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = append(t.spans, span)
}

// End finalizes the span and records its metrics attributes.
func (s *spanHandle) End(metrics middleware.RequestMetrics) {
	attrs := make(map[string]string, len(s.attrs)+3)
	for k, v := range s.attrs {
		attrs[k] = v
	}

	attrs["http.status_code"] = http.StatusText(metrics.Status)
	attrs["http.response_bytes"] = strconv.Itoa(metrics.Bytes)
	attrs["plumego.trace_id"] = metrics.TraceID

	status := "Ok"
	statusMsg := ""
	if metrics.Status >= http.StatusInternalServerError {
		status = "Error"
		statusMsg = http.StatusText(metrics.Status)
	}

	s.tracer.record(Span{
		Name:          s.name,
		Attributes:    attrs,
		Status:        status,
		StatusMessage: statusMsg,
	})
}
