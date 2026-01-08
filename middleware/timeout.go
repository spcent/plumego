package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

// Timeout creates a middleware that enforces a maximum duration for a request.
// If the downstream handler does not complete before the deadline, the request
// context is canceled and a 504 Gateway Timeout response is returned.
func Timeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()

			r = r.WithContext(ctx)

			tw := newTimeoutResponseWriter(ctx)
			done := make(chan struct{})

			go func() {
				defer close(done)
				next.ServeHTTP(tw, r)
			}()

			select {
			case <-done:
				tw.WriteTo(w)
			case <-ctx.Done():
				contract.WriteError(w, r, contract.APIError{
					Status:   http.StatusGatewayTimeout,
					Code:     "request_timeout",
					Category: contract.CategoryServer,
					Message:  "request timed out",
				})
			}
		})
	}
}

func newTimeoutResponseWriter(ctx context.Context) *timeoutResponseWriter {
	return &timeoutResponseWriter{
		header: http.Header{},
		ctx:    ctx,
	}
}

type timeoutResponseWriter struct {
	header http.Header
	body   []byte
	status int
	ctx    context.Context
}

func (w *timeoutResponseWriter) Header() http.Header {
	return w.header
}

func (w *timeoutResponseWriter) WriteHeader(statusCode int) {
	if w.status == 0 {
		w.status = statusCode
	}
}

func (w *timeoutResponseWriter) Write(p []byte) (int, error) {
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	default:
	}

	if w.status == 0 {
		w.status = http.StatusOK
	}

	w.body = append(w.body, p...)
	return len(p), nil
}

func (w *timeoutResponseWriter) WriteTo(dst http.ResponseWriter) {
	for k, values := range w.header {
		for _, v := range values {
			dst.Header().Add(k, v)
		}
	}

	status := w.status
	if status == 0 {
		status = http.StatusOK
	}

	dst.WriteHeader(status)
	_, _ = dst.Write(w.body)
}
