package tracing

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/contract"
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
			prepared := internalobs.PrepareRequest(w, r)
			r = prepared.Request
			requestID := prepared.RequestID
			ctx := r.Context()

			ctx, span := tracer.Start(ctx, r)
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

			r = r.WithContext(ctx)
			if spanID != "" {
				w.Header().Set("X-Span-ID", spanID)
			}

			recorder := prepared.Recorder
			next.ServeHTTP(recorder, r)

			if span != nil {
				span.End(recorder.StatusCode(), recorder.BytesWritten(), requestID)
			}
		})
	}
}
