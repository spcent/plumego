package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
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

func EnsureRequestID(r *http.Request) string {
	if id := contract.RequestIDFromContext(r.Context()); id != "" {
		return id
	}
	if id := RequestIDFromRequest(r); id != "" {
		return id
	}
	return contract.NewRequestID()
}

func PrepareRequest(w http.ResponseWriter, r *http.Request) PreparedRequest {
	requestID := EnsureRequestID(r)
	r = r.WithContext(contract.WithRequestID(r.Context(), requestID))
	if w != nil {
		w.Header().Set(contract.RequestIDHeader, requestID)
	}
	return PreparedRequest{
		Request:   r,
		Recorder:  NewResponseRecorder(w),
		RequestID: requestID,
		StartedAt: time.Now(),
	}
}

type spanContextCarrier interface {
	TraceID() string
	SpanID() string
}

func ExtractSpanContext(ctx context.Context, span spanContextCarrier) (string, string) {
	if span != nil {
		return span.TraceID(), span.SpanID()
	}
	if tc := contract.TraceContextFromContext(ctx); tc != nil {
		return string(tc.TraceID), string(tc.SpanID)
	}
	return "", ""
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

type ResponseRecorder = internaltransport.ResponseRecorder

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return internaltransport.NewResponseRecorder(w)
}
