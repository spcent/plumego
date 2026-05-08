package recovery

import (
	"bufio"
	"net"
	"net/http"
	"reflect"

	contract "github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
)

// Recovery recovers from panics in request handlers and returns a 500 Internal Server Error.
//
// This middleware prevents the entire application from crashing when a panic occurs in a request handler.
// It logs sanitized panic metadata server-side and returns a generic 500 response to the client.
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
//  2. Logs sanitized panic metadata with trace ID
//  3. Returns a generic 500 Internal Server Error response (no internal details exposed)
//
// Note: This middleware should be placed early in the middleware chain to ensure
// it can catch panics from all downstream handlers.
func Recovery(logger log.StructuredLogger) middleware.Middleware {
	if logger == nil {
		panic("recovery: logger cannot be nil")
	}
	return func(next http.Handler) http.Handler {
		return recoveryHandler(next, logger)
	}
}

func recoveryHandler(next http.Handler, logger log.StructuredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &recoveryResponseWriter{ResponseWriter: w}
		defer func() {
			if rec := recover(); rec != nil {
				fields := internalobs.MiddlewareLogFields(r, http.StatusInternalServerError, 0)
				fields["panic_type"] = panicType(rec)
				internalobs.RunSafeFinalizer(func() {
					logger.WithFields(log.Fields(internalobs.RedactFields(fields))).Error("panic recovered")
				})
				if rw.wrote {
					return
				}
				internaltransport.WriteTransportError(rw, r, http.StatusInternalServerError, contract.CodeInternalError, "internal server error", contract.CategoryServer, nil)
			}
		}()
		next.ServeHTTP(rw, r)
	})
}

type recoveryResponseWriter struct {
	http.ResponseWriter
	wrote bool
}

func (w *recoveryResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
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

func panicType(rec any) string {
	if rec == nil {
		return "unknown"
	}
	return reflect.TypeOf(rec).String()
}

func (w *recoveryResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	w.wrote = true
	return hijacker.Hijack()
}
