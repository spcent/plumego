package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
)

// RequestMetrics captures common observability attributes for a request.
type RequestMetrics struct {
	Method    string
	Path      string
	Status    int
	Bytes     int
	Duration  time.Duration
	TraceID   string
	UserAgent string
}

// MetricsCollector can be plugged into the logging middleware to export metrics.
type MetricsCollector interface {
	Observe(ctx context.Context, metrics RequestMetrics)
}

// TraceSpan represents a started tracing span.
type TraceSpan interface {
	End(metrics RequestMetrics)
}

// SpanContext exposes identifiers for log correlation.
type SpanContext interface {
	TraceID() string
	SpanID() string
}

// Tracer starts a span for the incoming request.
type Tracer interface {
	Start(ctx context.Context, r *http.Request) (context.Context, TraceSpan)
}

// Logging provides request logging with structured fields and optional metrics/tracing hooks.
func Logging(logger log.StructuredLogger, metrics MetricsCollector, tracer Tracer) Middleware {
	return func(next Handler) Handler {
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
