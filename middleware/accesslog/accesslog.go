package accesslog

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
)

// Middleware is the canonical access-log middleware constructor.
func Middleware(logger log.StructuredLogger) middleware.Middleware {
	if logger == nil {
		panic("accesslog: logger cannot be nil")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			prepared := internalobs.PrepareRequest(w, r)
			r = prepared.Request
			recorder := prepared.Recorder

			defer internalobs.FinishPreservingPanic(func() {
				metricsData := prepared.Complete(r)
				rc := contract.RequestContextFromContext(r.Context())

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
				if headerSpanID := recorder.Header().Get(internalobs.SpanIDHeader); headerSpanID != "" {
					fields["span_id"] = headerSpanID
				} else if tc := contract.TraceContextFromContext(r.Context()); tc != nil && tc.SpanID != "" {
					fields["span_id"] = tc.SpanID
				}

				logger.WithFields(redactedLogFields(fields)).Info("request completed")
			})

			next.ServeHTTP(recorder, r)
		})
	}
}

func redactedLogFields(fields map[string]any) log.Fields {
	return log.Fields(internalobs.RedactFields(fields))
}
