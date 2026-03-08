package recovery

import (
	"net/http"

	contract "github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
)

// RecoveryMiddleware recovers from panics in request handlers and returns a 500 Internal Server Error.
//
// This middleware prevents the entire application from crashing when a panic occurs in a request handler.
// It logs the panic details server-side and returns a generic 500 response to the client.
// Panic details are intentionally omitted from the response to avoid leaking internal state.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/recovery"
//
//	handler := recovery.RecoveryMiddleware(myHandler)
//
// When a panic occurs, the middleware:
//  1. Recovers the panic and prevents the application from crashing
//  2. Logs the panic details with trace ID
//  3. Returns a generic 500 Internal Server Error response (no internal details exposed)
//
// Note: This middleware should be placed early in the middleware chain to ensure
// it can catch panics from all downstream handlers.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return RecoveryWithLogger(nil)(next)
}

// RecoveryWithLogger returns recovery middleware using the provided logger.
// If logger is nil, it falls back to the default logger.
func RecoveryWithLogger(logger log.StructuredLogger) middleware.Middleware {
	if logger == nil {
		logger = log.NewGLogger()
	}

	return func(next http.Handler) http.Handler {
		return recoveryHandler(next, logger)
	}
}

func recoveryHandler(next http.Handler, logger log.StructuredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Log panic details server-side; never expose them in the response.
				logger.WithFields(log.Fields{"panic": rec, "trace_id": contract.TraceIDFromContext(r.Context())}).Error("panic recovered")
				contract.WriteError(w, r, contract.APIError{
					Status:   http.StatusInternalServerError,
					Code:     "internal_error",
					Category: contract.CategoryServer,
					Message:  "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
