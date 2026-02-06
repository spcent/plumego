package circuitbreaker

import (
	"encoding/json"
	"net/http"

	"github.com/spcent/plumego/utils"
)

// Middleware creates a circuit breaker middleware
//
// Example:
//
//	app.Use(circuitbreaker.Middleware(circuitbreaker.Config{
//		Name:             "api",
//		FailureThreshold: 0.6,
//		Timeout:          30 * time.Second,
//	}))
func Middleware(config Config) func(http.Handler) http.Handler {
	cb := New(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to execute request through circuit breaker
			err := cb.Call(func() error {
				// Wrap response writer to capture status code
				wrapper := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

				// Execute handler
				next.ServeHTTP(wrapper, r)

				// Consider 5xx status codes as failures
				if wrapper.statusCode >= 500 {
					return ErrServerError
				}

				return nil
			})

			// If circuit breaker rejected the request
			if err == ErrCircuitOpen {
				writeCircuitOpenResponse(w, cb)
				return
			}

			if err == ErrTooManyRequests {
				writeTooManyRequestsResponse(w, cb)
				return
			}
		})
	}
}

// MiddlewareWithErrorHandler creates a circuit breaker middleware with custom error handling
func MiddlewareWithErrorHandler(config Config, errorHandler ErrorHandler) func(http.Handler) http.Handler {
	cb := New(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := cb.Call(func() error {
				wrapper := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
				next.ServeHTTP(wrapper, r)

				if wrapper.statusCode >= 500 {
					return ErrServerError
				}

				return nil
			})

			if err != nil {
				errorHandler(w, r, cb, err)
				return
			}
		})
	}
}

// ErrorHandler handles circuit breaker errors
type ErrorHandler func(w http.ResponseWriter, r *http.Request, cb *CircuitBreaker, err error)

// statusWriter wraps http.ResponseWriter to capture status code
type statusWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return utils.SafeWrite(w.ResponseWriter, b)
}

// writeCircuitOpenResponse writes a 503 response when circuit is open
func writeCircuitOpenResponse(w http.ResponseWriter, cb *CircuitBreaker) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Circuit-Breaker-State", cb.State().String())
	utils.EnsureNoSniff(w.Header())
	w.WriteHeader(http.StatusServiceUnavailable)

	json.NewEncoder(w).Encode(map[string]any{
		"error":   "Service Unavailable",
		"message": "Circuit breaker is open. The service is temporarily unavailable.",
		"circuit": cb.Name(),
		"state":   cb.State().String(),
	})
}

// writeTooManyRequestsResponse writes a 429 response for rate limiting
func writeTooManyRequestsResponse(w http.ResponseWriter, cb *CircuitBreaker) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Circuit-Breaker-State", cb.State().String())
	utils.EnsureNoSniff(w.Header())
	w.WriteHeader(http.StatusTooManyRequests)

	json.NewEncoder(w).Encode(map[string]any{
		"error":   "Too Many Requests",
		"message": "Circuit breaker is in half-open state. Too many concurrent requests.",
		"circuit": cb.Name(),
		"state":   cb.State().String(),
	})
}

// ErrServerError indicates a server error (5xx)
var ErrServerError = http.ErrAbortHandler
