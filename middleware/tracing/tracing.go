package tracing

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
)

type TraceSpan interface {
	End(status, bytes int, requestID string)
	TraceID() string
	SpanID() string
}

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
			_, spanID := internalobs.ExtractSpanContext(ctx, span)
			r = r.WithContext(ctx)
			r = internalobs.AttachSpanID(w, r, spanID)

			recorder := prepared.Recorder
			next.ServeHTTP(recorder, r)

			if span != nil {
				span.End(recorder.StatusCode(), recorder.BytesWritten(), requestID)
			}
		})
	}
}
