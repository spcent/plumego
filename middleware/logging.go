package middleware

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
)

// RequestMetrics captures common observability attributes for a request.
//
// This struct is used by the logging middleware to collect metrics about each request.
// It can be passed to custom metrics collectors for monitoring and analysis.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	type MyMetricsCollector struct{}
//
//	func (c *MyMetricsCollector) Observe(ctx context.Context, metrics middleware.RequestMetrics) {
//		// Send metrics to your monitoring system
//		fmt.Printf("Request: %s %s - Status: %d - Duration: %v\n",
//			metrics.Method, metrics.Path, metrics.Status, metrics.Duration)
//	}
//
//	collector := &MyMetricsCollector{}
//	handler := middleware.Logging(middleware.NewGLogger(), collector, nil)(myHandler)
type RequestMetrics struct {
	// Method is the HTTP method (GET, POST, etc.)
	Method string

	// Path is the request path
	Path string

	// Status is the HTTP status code
	Status int

	// Bytes is the number of bytes written in the response
	Bytes int

	// Duration is the request processing time
	Duration time.Duration

	// TraceID is the trace identifier for request correlation
	TraceID string

	// UserAgent is the client's user agent string
	UserAgent string
}

// MetricsCollector can be plugged into the logging middleware to export metrics.
// This interface is maintained for backward compatibility but new code should use
// the unified metrics.MetricsCollector interface.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	type MyMetricsCollector struct{}
//
//	func (c *MyMetricsCollector) Observe(ctx context.Context, metrics middleware.RequestMetrics) {
//		// Collect metrics
//	}
//
//	collector := &MyMetricsCollector{}
//	handler := middleware.Logging(middleware.NewGLogger(), collector, nil)(myHandler)
type MetricsCollector interface {
	Observe(ctx context.Context, metrics RequestMetrics)
}

// UnifiedMetricsCollector is an adapter that allows the unified metrics.MetricsCollector
// to be used with the logging middleware.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector()
//	handler := middleware.Logging(middleware.NewGLogger(), collector, nil)(myHandler)
type UnifiedMetricsCollector interface {
	ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration)
}

// TraceSpan represents a started tracing span.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	type MyTraceSpan struct{}
//
//	func (s *MyTraceSpan) End(metrics middleware.RequestMetrics) {
//		// End the span and record metrics
//	}
//
//	type MyTracer struct{}
//
//	func (t *MyTracer) Start(ctx context.Context, r *http.Request) (context.Context, middleware.TraceSpan) {
//		span := &MyTraceSpan{}
//		return ctx, span
//	}
//
//	tracer := &MyTracer{}
//	handler := middleware.Logging(middleware.NewGLogger(), nil, tracer)(myHandler)
type TraceSpan interface {
	End(metrics RequestMetrics)
}

// SpanContext exposes identifiers for log correlation.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	type MySpanContext struct{}
//
//	func (s *MySpanContext) TraceID() string {
//		return "trace-123"
//	}
//
//	func (s *MySpanContext) SpanID() string {
//		return "span-456"
//	}
//
//	type MyTraceSpan struct{}
//
//	func (s *MyTraceSpan) End(metrics middleware.RequestMetrics) {
//		// End the span
//	}
//
//	type MyTracer struct{}
//
//	func (t *MyTracer) Start(ctx context.Context, r *http.Request) (context.Context, middleware.TraceSpan) {
//		span := &MyTraceSpan{}
//		return ctx, span
//	}
//
//	tracer := &MyTracer{}
//	handler := middleware.Logging(middleware.NewGLogger(), nil, tracer)(myHandler)
type SpanContext interface {
	TraceID() string
	SpanID() string
}

// Tracer starts a span for the incoming request.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	type MyTraceSpan struct{}
//
//	func (s *MyTraceSpan) End(metrics middleware.RequestMetrics) {
//		// End the span
//	}
//
//	type MyTracer struct{}
//
//	func (t *MyTracer) Start(ctx context.Context, r *http.Request) (context.Context, middleware.TraceSpan) {
//		span := &MyTraceSpan{}
//		return ctx, span
//	}
//
//	tracer := &MyTracer{}
//	handler := middleware.Logging(middleware.NewGLogger(), nil, tracer)(myHandler)
type Tracer interface {
	Start(ctx context.Context, r *http.Request) (context.Context, TraceSpan)
}

// Logging provides request logging with structured fields and optional metrics/tracing hooks.
//
// This middleware logs each request with structured fields including method, path, status,
// duration, bytes, trace ID, and user agent. It also supports custom metrics collectors
// and tracers for observability.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	// Basic logging
//	handler := middleware.Logging(middleware.NewGLogger(), nil, nil)(myHandler)
//
//	// With metrics collector
//	collector := &MyMetricsCollector{}
//	handler := middleware.Logging(middleware.NewGLogger(), collector, nil)(myHandler)
//
//	// With tracer
//	tracer := &MyTracer{}
//	handler := middleware.Logging(middleware.NewGLogger(), nil, tracer)(myHandler)
//
// Log format:
//
//	INFO request completed trace_id=abc123 method=GET path=/api/users status=200 bytes=1234 duration_ms=45 user_agent=Mozilla/5.0
//
// The middleware also sets the following headers:
//   - X-Request-ID: The trace ID (or generated if not provided)
//   - X-Span-ID: The span ID (if available)
func Logging(logger log.StructuredLogger, metrics MetricsCollector, tracer Tracer) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := ensureTraceID(r)
			ctx := context.WithValue(r.Context(), contract.TraceIDKey{}, traceID)

			recorder := newResponseRecorder(w)
			start := time.Now()

			var span TraceSpan
			if tracer != nil {
				ctx, span = tracer.Start(ctx, r)
			}

			spanTraceID, spanID := extractSpanContext(ctx, span)
			if spanTraceID != "" {
				traceID = spanTraceID
			}
			ctx = context.WithValue(ctx, contract.TraceIDKey{}, traceID)
			if spanID != "" && contract.TraceContextFromContext(ctx) == nil {
				ctx = contract.ContextWithTraceContext(ctx, contract.TraceContext{
					TraceID: contract.TraceID(traceID),
					SpanID:  contract.SpanID(spanID),
				})
			}

			r = r.WithContext(ctx)
			w.Header().Set("X-Request-ID", traceID)
			if spanID != "" {
				w.Header().Set("X-Span-ID", spanID)
			}

			next.ServeHTTP(recorder, r)

			metricsData := RequestMetrics{
				Method:    r.Method,
				Path:      r.URL.Path,
				Status:    recorder.StatusCode(),
				Bytes:     recorder.BytesWritten(),
				Duration:  time.Since(start),
				TraceID:   traceID,
				UserAgent: r.UserAgent(),
			}

			if metrics != nil {
				metrics.Observe(r.Context(), metricsData)
			}

			if span != nil {
				span.End(metricsData)
			}

			fields := log.Fields{
				"trace_id":    traceID,
				"method":      metricsData.Method,
				"path":        metricsData.Path,
				"status":      metricsData.Status,
				"bytes":       metricsData.Bytes,
				"duration_ms": metricsData.Duration.Milliseconds(),
				"user_agent":  metricsData.UserAgent,
			}
			if spanID != "" {
				fields["span_id"] = spanID
			}

			logger.WithFields(fields).Info("request completed", nil)
		})
	}
}

func ensureTraceID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	if id := r.Header.Get("X-Trace-ID"); id != "" {
		return id
	}
	return log.NewTraceID()
}

func extractSpanContext(ctx context.Context, span TraceSpan) (string, string) {
	if tc := contract.TraceContextFromContext(ctx); tc != nil {
		return string(tc.TraceID), string(tc.SpanID)
	}
	if sc, ok := span.(SpanContext); ok {
		return sc.TraceID(), sc.SpanID()
	}
	return "", ""
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	// SECURITY NOTE: This Write method only records response metrics.
	// The 'p' parameter contains response data from upstream handlers,
	// not user input. This middleware does not modify response content
	// and therefore does not introduce XSS vulnerabilities.
	// XSS protection should be implemented in handlers that generate HTML content.
	n, err := r.ResponseWriter.Write(p)
	r.bytes += n
	return n, err
}

func (r *responseRecorder) StatusCode() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func (r *responseRecorder) BytesWritten() int {
	return r.bytes
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hj.Hijack()
}

func (r *responseRecorder) Flush() {
	if fl, ok := r.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (r *responseRecorder) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := r.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
