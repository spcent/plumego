// Package tracing provides a transport-layer tracing middleware. It defines
// the [Tracer] and [TraceSpan] interfaces that callers must implement; it does
// not own the concrete tracing infrastructure.
//
// Concrete implementations (OpenTelemetry adapters, span collectors, samplers)
// belong in x/observability/tracer. The interface is defined here so that
// stable middleware can remain decoupled from extension packages: stable roots
// must not import x/*.
package tracing

import (
	"context"
	"net/http"

	"github.com/spcent/plumego/middleware"
	internaltelemetry "github.com/spcent/plumego/middleware/internal/telemetry"
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
			prepared := internaltelemetry.PrepareRequest(w, r)
			r, span := internaltelemetry.BeginTrace(w, prepared, func(ctx context.Context, r *http.Request) (context.Context, internaltelemetry.TraceSpan) {
				return tracer.Start(ctx, r)
			})
			recorder := prepared.Recorder
			defer internaltelemetry.FinishPreservingPanic(func() {
				internaltelemetry.EndTrace(span, prepared.Complete(r))
			})

			next.ServeHTTP(recorder, r)
		})
	}
}
