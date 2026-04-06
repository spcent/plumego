package accesslog

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
)

func Middleware(logger log.StructuredLogger) middleware.Middleware {
	if logger == nil {
		panic("access logger cannot be nil")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			prepared := internalobs.PrepareRequest(w, r)
			r = prepared.Request
			recorder := prepared.Recorder
			requestID := prepared.RequestID

			next.ServeHTTP(recorder, r)

			if headerRequestID := recorder.Header().Get(contract.RequestIDHeader); headerRequestID != "" {
				requestID = headerRequestID
			}
			metricsData := internalobs.BuildRequestMetrics(r, recorder, prepared.StartedAt, requestID)
			rc := contract.RequestContextFromContext(r.Context())
			fields := contract.NewObservabilityPolicy().MiddlewareLogFields(r, metricsData.Status, metricsData.Duration)
			fields["bytes"] = metricsData.Bytes
			fields["user_agent"] = metricsData.UserAgent
			fields["client_ip"] = internaltransport.ClientIP(r)
			if metricsData.Route != "" {
				fields["route"] = metricsData.Route
			}
			if rc.RouteName != "" {
				fields["route_name"] = rc.RouteName
			}
			if spanID := recorder.Header().Get("X-Span-ID"); spanID != "" {
				fields["span_id"] = spanID
			} else if tc := contract.TraceContextFromContext(r.Context()); tc != nil && tc.SpanID != "" {
				fields["span_id"] = tc.SpanID
			}

			logger.WithFields(fields).Info("request completed")
		})
	}
}

// Logging keeps the old combined convenience shape, but lives in a focused package.
func Logging(logger log.StructuredLogger, observer metrics.HTTPObserver, tracer mwtracing.Tracer) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			prepared := internalobs.PrepareRequest(w, r)
			r = prepared.Request
			recorder := prepared.Recorder
			requestID := prepared.RequestID
			ctx := r.Context()

			var span mwtracing.TraceSpan
			if tracer != nil {
				ctx, span = tracer.Start(ctx, r)
			}

			spanTraceID, spanID := internalobs.ExtractSpanContext(ctx, span)
			if spanTraceID != "" || spanID != "" {
				traceContext := contract.TraceContext{}
				if existing := contract.TraceContextFromContext(ctx); existing != nil {
					traceContext = *existing
				}
				if spanTraceID != "" {
					traceContext.TraceID = contract.TraceID(spanTraceID)
				}
				if spanID != "" {
					traceContext.SpanID = contract.SpanID(spanID)
				}
				ctx = contract.WithTraceContext(ctx, traceContext)
			}
			if spanID != "" {
				w.Header().Set("X-Span-ID", spanID)
			}

			r = r.WithContext(ctx)

			next.ServeHTTP(recorder, r)

			rc := contract.RequestContextFromContext(r.Context())
			metricsData := internalobs.BuildRequestMetrics(r, recorder, prepared.StartedAt, requestID)

			if observer != nil {
				path := metricsData.Path
				if metricsData.Route != "" {
					path = metricsData.Route
				}
				observer.ObserveHTTP(r.Context(), metricsData.Method, path, metricsData.Status, metricsData.Bytes, metricsData.Duration)
			}
			if span != nil {
				span.End(metricsData.Status, metricsData.Bytes, requestID)
			}

			fields := contract.NewObservabilityPolicy().MiddlewareLogFields(r, metricsData.Status, metricsData.Duration)
			fields["bytes"] = metricsData.Bytes
			fields["user_agent"] = metricsData.UserAgent
			fields["client_ip"] = internaltransport.ClientIP(r)
			if metricsData.Route != "" {
				fields["route"] = metricsData.Route
			}
			if rc.RouteName != "" {
				fields["route_name"] = rc.RouteName
			}
			if spanID != "" {
				fields["span_id"] = spanID
			}

			logger.WithFields(fields).Info("request completed")
		})
	}
}
