package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
	"github.com/spcent/plumego/middleware/requestid"
)

// RequestMetrics captures common observability attributes for a request.
type RequestMetrics struct {
	Method    string
	Path      string
	Route     string
	Status    int
	Bytes     int
	Duration  time.Duration
	RequestID string
	UserAgent string
}

// PreparedRequest captures the shared request state stable observability
// middleware needs before delegating to the next handler.
type PreparedRequest struct {
	Request   *http.Request
	Recorder  *ResponseRecorder
	RequestID string
	StartedAt time.Time
}

// PrepareRequest establishes the canonical observability fallback path for
// stable middleware. When requestid.Middleware() is not present, observability
// middleware still stamps a request ID so logging, metrics, and tracing share
// the same correlation surface.
func PrepareRequest(w http.ResponseWriter, r *http.Request) PreparedRequest {
	requestID := requestid.EnsureRequestID(r, nil)
	r = requestid.AttachRequestID(w, r, requestID, false)
	return PreparedRequest{
		Request:   r,
		Recorder:  NewResponseRecorder(w),
		RequestID: requestID,
		StartedAt: time.Now(),
	}
}

// Complete derives request metrics from the prepared request lifecycle state.
func (p PreparedRequest) Complete(r *http.Request) RequestMetrics {
	return BuildRequestMetrics(r, p.Recorder, p.StartedAt, p.RequestID)
}

type spanContextCarrier interface {
	TraceID() string
	SpanID() string
}

type TraceSpan interface {
	spanContextCarrier
	End(status, bytes int, requestID string)
}

type TraceStarter func(ctx context.Context, r *http.Request) (context.Context, TraceSpan)

func ExtractSpanContext(ctx context.Context, span spanContextCarrier) (string, string) {
	if span != nil {
		return span.TraceID(), span.SpanID()
	}
	if tc := contract.TraceContextFromContext(ctx); tc != nil {
		return string(tc.TraceID), string(tc.SpanID)
	}
	return "", ""
}

// BeginTrace applies the shared tracing start flow used by stable observability
// middleware. When tracer is nil the request is returned unchanged.
func BeginTrace(w http.ResponseWriter, prepared PreparedRequest, start TraceStarter) (*http.Request, TraceSpan, string) {
	r := prepared.Request
	if start == nil {
		return r, nil, ""
	}

	ctx, span := start(r.Context(), r)
	_, spanID := ExtractSpanContext(ctx, span)
	r = r.WithContext(ctx)
	r = AttachSpanID(w, r, spanID)
	return r, span, spanID
}

// EndTrace applies the canonical trace completion path for stable
// observability middleware.
func EndTrace(span TraceSpan, metrics RequestMetrics) {
	if span == nil {
		return
	}
	span.End(metrics.Status, metrics.Bytes, metrics.RequestID)
}

func BuildRequestMetrics(r *http.Request, recorder *ResponseRecorder, started time.Time, requestID string) RequestMetrics {
	rc := contract.RequestContextFromContext(r.Context())
	metricsData := RequestMetrics{
		Method:    r.Method,
		Path:      r.URL.Path,
		Status:    recorder.StatusCode(),
		Bytes:     recorder.BytesWritten(),
		Duration:  time.Since(started),
		RequestID: requestID,
		UserAgent: r.UserAgent(),
	}
	if rc.RoutePattern != "" {
		metricsData.Route = rc.RoutePattern
	}
	return metricsData
}

// ObservedPath returns the canonical path label for transport metrics.
func (m RequestMetrics) ObservedPath() string {
	if m.Route != "" {
		return m.Route
	}
	return m.Path
}

type ResponseRecorder = internaltransport.ResponseRecorder

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return internaltransport.NewResponseRecorder(w)
}
