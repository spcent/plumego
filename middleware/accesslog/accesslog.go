package accesslog

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
)

func Middleware(logger log.StructuredLogger) middleware.Middleware {
	if logger == nil {
		panic("access logger cannot be nil")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := internalobs.EnsureTraceID(r)
			r = r.WithContext(contract.WithTraceIDString(r.Context(), traceID))
			w.Header().Set(contract.RequestIDHeader, traceID)
			recorder := internalobs.NewResponseRecorder(w)
			start := time.Now()

			next.ServeHTTP(recorder, r)

			if headerTraceID := recorder.Header().Get(contract.RequestIDHeader); headerTraceID != "" {
				traceID = headerTraceID
			}
			metricsData := internalobs.BuildRequestMetrics(r, recorder, start, traceID)
			rc := contract.RequestContextFromContext(r.Context())
			fields := contract.NewObservabilityPolicy().MiddlewareLogFields(r, metricsData.Status, metricsData.Duration)
			fields["bytes"] = metricsData.Bytes
			fields["user_agent"] = metricsData.UserAgent
			fields["client_ip"] = internalobs.ClientIP(r)
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
			traceID := internalobs.EnsureTraceID(r)
			ctx := contract.WithTraceIDString(r.Context(), traceID)

			recorder := internalobs.NewResponseRecorder(w)
			start := time.Now()

			var span metrics.TraceSpan
			if tracer != nil {
				ctx, span = tracer.Start(ctx, r)
			}

			spanTraceID, spanID := internalobs.ExtractSpanContext(ctx, span)
			if spanTraceID != "" {
				traceID = spanTraceID
			}
			ctx = contract.WithTraceIDString(ctx, traceID)
			if spanID != "" {
				ctx = contract.WithSpanIDString(ctx, spanID)
			}

			r = r.WithContext(ctx)
			w.Header().Set(contract.RequestIDHeader, traceID)
			if spanID != "" {
				w.Header().Set("X-Span-ID", spanID)
			}

			next.ServeHTTP(recorder, r)

			rc := contract.RequestContextFromContext(r.Context())
			metricsData := internalobs.BuildRequestMetrics(r, recorder, start, traceID)

			if observer != nil {
				path := metricsData.Path
				if metricsData.Route != "" {
					path = metricsData.Route
				}
				observer.ObserveHTTP(r.Context(), metricsData.Method, path, metricsData.Status, metricsData.Bytes, metricsData.Duration)
			}
			if span != nil {
				span.End(metricsData.Status, metricsData.Bytes, traceID)
			}

			fields := contract.NewObservabilityPolicy().MiddlewareLogFields(r, metricsData.Status, metricsData.Duration)
			fields["bytes"] = metricsData.Bytes
			fields["user_agent"] = metricsData.UserAgent
			fields["client_ip"] = internalobs.ClientIP(r)
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
