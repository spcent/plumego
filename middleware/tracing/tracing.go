package tracing

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
)

type TraceSpan = metrics.TraceSpan

type Tracer interface {
	Start(ctx context.Context, r *http.Request) (context.Context, TraceSpan)
}

func Middleware(tracer Tracer) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		if tracer == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := internalobs.EnsureTraceID(r)
			ctx := contract.WithTraceIDString(r.Context(), traceID)
			ctx = log.WithTraceID(ctx, traceID)

			ctx, span := tracer.Start(ctx, r)
			spanTraceID, spanID := internalobs.ExtractSpanContext(ctx, span)
			if spanTraceID != "" {
				traceID = spanTraceID
				ctx = contract.WithTraceIDString(ctx, traceID)
				ctx = log.WithTraceID(ctx, traceID)
			}
			if spanID != "" {
				existing := contract.TraceContextFromContext(ctx)
				if existing == nil || existing.SpanID == "" {
					ctx = contract.ContextWithTraceContext(ctx, contract.TraceContext{
						TraceID: contract.TraceID(traceID),
						SpanID:  contract.SpanID(spanID),
					})
				}
			}

			r = r.WithContext(ctx)
			w.Header().Set(contract.RequestIDHeader, traceID)
			if spanID != "" {
				w.Header().Set("X-Span-ID", spanID)
			}

			recorder := internalobs.NewResponseRecorder(w)
			next.ServeHTTP(recorder, r)

			if span != nil {
				span.End(recorder.StatusCode(), recorder.BytesWritten(), traceID)
			}
		})
	}
}
