package recovery

import (
	"bufio"
	"errors"
	"net"
	"net/http"

	contract "github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
)

// ErrNilLogger is returned by RecoveryE when the logger dependency is nil.
var ErrNilLogger = errors.New("recovery: logger cannot be nil")

// Recovery recovers from panics in request handlers and returns a 500 Internal Server Error.
//
// This middleware prevents the entire application from crashing when a panic occurs in a request handler.
// It logs the panic details server-side and returns a generic 500 response to the client.
// Panic details are intentionally omitted from the response to avoid leaking internal state.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/recovery"
//
//	handler := recovery.Recovery(logger)(myHandler)
//
// When a panic occurs, the middleware:
//  1. Recovers the panic and prevents the application from crashing
//  2. Logs the panic details with trace ID
//  3. Returns a generic 500 Internal Server Error response (no internal details exposed)
//
// Note: This middleware should be placed early in the middleware chain to ensure
// it can catch panics from all downstream handlers.
func Recovery(logger log.StructuredLogger) middleware.Middleware {
	mw, err := RecoveryE(logger)
	if err != nil {
		panic(err.Error())
	}
	return mw
}

// RecoveryE creates recovery middleware and reports invalid dependencies without
// panicking.
func RecoveryE(logger log.StructuredLogger) (middleware.Middleware, error) {
	if logger == nil {
		return nil, ErrNilLogger
	}
	return func(next http.Handler) http.Handler {
		return recoveryHandler(next, logger)
	}, nil
}

func recoveryHandler(next http.Handler, logger log.StructuredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &recoveryResponseWriter{ResponseWriter: w}
		defer func() {
			if rec := recover(); rec != nil {
				// Log panic details server-side; never expose them in the response.
				fields := internalobs.MiddlewareLogFields(r, http.StatusInternalServerError, 0)
				fields["panic"] = rec
				logger.WithFields(log.Fields(internalobs.RedactFields(fields))).Error("panic recovered")
				if rw.wrote {
					return
				}
				middleware.WriteTransportError(rw, r, http.StatusInternalServerError, contract.CodeInternalError, "internal server error", contract.CategoryServer, nil)
			}
		}()
		next.ServeHTTP(rw, r)
	})
}

type recoveryResponseWriter struct {
	http.ResponseWriter
	wrote bool
}

func (w *recoveryResponseWriter) WriteHeader(statusCode int) {
	if w.wrote {
		return
	}
	w.wrote = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *recoveryResponseWriter) Write(p []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *recoveryResponseWriter) Flush() {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *recoveryResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	w.wrote = true
	return hijacker.Hijack()
}
