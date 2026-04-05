package observability

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
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
	TraceID   string
	UserAgent string
}

func EnsureTraceID(r *http.Request) string {
	if id := contract.TraceIDFromContext(r.Context()); id != "" {
		return id
	}
	if id := contract.DefaultObservabilityPolicy.RequestIDFromRequest(r); id != "" {
		return id
	}
	return log.NewTraceID()
}

func ExtractSpanContext(ctx context.Context, span metrics.TraceSpan) (string, string) {
	if span != nil {
		return span.TraceID(), span.SpanID()
	}
	if tc := contract.TraceContextFromContext(ctx); tc != nil {
		return string(tc.TraceID), string(tc.SpanID)
	}
	return "", ""
}

func BuildRequestMetrics(r *http.Request, recorder *ResponseRecorder, started time.Time, traceID string) RequestMetrics {
	rc := contract.RequestContextFromContext(r.Context())
	metricsData := RequestMetrics{
		Method:    r.Method,
		Path:      r.URL.Path,
		Status:    recorder.StatusCode(),
		Bytes:     recorder.BytesWritten(),
		Duration:  time.Since(started),
		TraceID:   traceID,
		UserAgent: r.UserAgent(),
	}
	if rc.RoutePattern != "" {
		metricsData.Route = rc.RoutePattern
	}
	return metricsData
}

type ResponseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := internaltransport.SafeWrite(r.ResponseWriter, p)
	r.bytes += n
	return n, err
}

func (r *ResponseRecorder) StatusCode() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func (r *ResponseRecorder) BytesWritten() int {
	return r.bytes
}

func (r *ResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hj.Hijack()
}

func (r *ResponseRecorder) Flush() {
	if fl, ok := r.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func ClientIP(r *http.Request) string {
	return internaltransport.ClientIP(r)
}
