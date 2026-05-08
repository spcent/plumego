package accesslog

import (
	"errors"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
)

// ErrNilLogger is returned when access logging is configured without a logger.
var ErrNilLogger = errors.New("accesslog: logger cannot be nil")

// Config controls access-log middleware behavior.
type Config struct {
	Logger log.StructuredLogger
}

// Middleware is the canonical access-log middleware constructor.
func Middleware(config Config) (middleware.Middleware, error) {
	if config.Logger == nil {
		return nil, ErrNilLogger
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
				} else if tc := contract.TraceContextFromContext(r.Context()); tc != nil && tc.HasSpanID() {
					fields["span_id"] = tc.SpanID
				}

				config.Logger.WithFields(log.Fields(internalobs.RedactFields(fields))).Info("request completed")
			})

			next.ServeHTTP(recorder, r)
		})
	}, nil
}
