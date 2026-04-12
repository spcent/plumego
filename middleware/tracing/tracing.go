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
			r, span, _ := internalobs.BeginTrace(w, prepared, func(ctx context.Context, r *http.Request) (context.Context, internalobs.TraceSpan) {
				return tracer.Start(ctx, r)
			})
			recorder := prepared.Recorder
			next.ServeHTTP(recorder, r)

			internalobs.EndTrace(span, prepared.Complete(r))
		})
	}
}
