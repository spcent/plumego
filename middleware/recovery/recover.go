package recovery

import (
	"net/http"

	contract "github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
)

// RecoveryMiddleware recovers from panics in request handlers and returns a 500 Internal Server Error.
//
// This middleware prevents the entire application from crashing when a panic occurs in a request handler.
// It logs the panic details and returns a structured error response to the client.
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
//  3. Returns a 500 Internal Server Error response with structured error message
//
// Error response format:
//
//	{
//	  "status": 500,
//	  "code": "internal_error",
//	  "message": "internal server error",
//	  "category": "server",
//	  "details": {
//	    "panic": "panic details"
//	  }
//	}
//
// Note: This middleware should be placed early in the middleware chain to ensure
// it can catch panics from all downstream handlers.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				contract.WriteError(w, r, contract.APIError{
					Status:   http.StatusInternalServerError,
					Code:     "internal_error",
					Category: contract.CategoryServer,
					Message:  "internal server error",
					Details:  map[string]any{"panic": rec},
				})
				logger := log.NewGLogger()
				logger.WithFields(log.Fields{"panic": rec, "trace_id": contract.TraceIDFromContext(r.Context())}).Error("panic recovered", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
