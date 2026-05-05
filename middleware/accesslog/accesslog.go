package accesslog

import (
	"context"
	"errors"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
)

// ErrNilLogger is returned by MiddlewareE when the logger dependency is nil.
var ErrNilLogger = errors.New("accesslog: logger cannot be nil")

// Middleware is the canonical access-log middleware constructor.
//
// The observer and tracer parameters are compatibility wiring for stable
// transport observability. The recommended production composition is
// accesslog.Middleware(logger, nil, nil) with standalone httpmetrics and
// tracing middleware when those signals are needed. If observer or tracer is
// passed here, do not also stack the standalone middleware for the same signal.
// Exporter selection, backend configuration, and broader telemetry pipelines
// belong in x/observability.
func Middleware(logger log.StructuredLogger, observer metrics.HTTPObserver, tracer mwtracing.Tracer) middleware.Middleware {
	mw, err := MiddlewareE(logger, observer, tracer)
	if err != nil {
		panic(err.Error())
	}
	return mw
}

// MiddlewareE creates access-log middleware and reports invalid dependencies
// without panicking. The observer and tracer parameters follow the same
// transport-only ownership boundary as Middleware.
func MiddlewareE(logger log.StructuredLogger, observer metrics.HTTPObserver, tracer mwtracing.Tracer) (middleware.Middleware, error) {
	if logger == nil {
		return nil, ErrNilLogger
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			prepared := internalobs.PrepareRequest(w, r)
			r = prepared.Request
			recorder := prepared.Recorder
			var startTrace internalobs.TraceStarter
			if tracer != nil {
				startTrace = func(ctx context.Context, r *http.Request) (context.Context, internalobs.TraceSpan) {
					return tracer.Start(ctx, r)
				}
			}
			r, span, spanID := internalobs.BeginTrace(w, prepared, startTrace)

			defer func() {
				metricsData := prepared.Complete(r)
				rc := contract.RequestContextFromContext(r.Context())

				if observer != nil {
					observer.ObserveHTTP(r.Context(), metricsData.Method, metricsData.ObservedPath(), metricsData.Status, metricsData.Bytes, metricsData.Duration)
				}
				internalobs.EndTrace(span, metricsData)

				fields := internalobs.MiddlewareLogFields(r, metricsData.Status, metricsData.Duration)
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
				} else if headerSpanID := recorder.Header().Get(internalobs.SpanIDHeader); headerSpanID != "" {
					fields["span_id"] = headerSpanID
				} else if tc := contract.TraceContextFromContext(r.Context()); tc != nil && tc.SpanID != "" {
					fields["span_id"] = tc.SpanID
				}

				logger.WithFields(fields).Info("request completed")
			}()

			next.ServeHTTP(recorder, r)
		})
	}, nil
}
