package timeout

import (
	"context"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	mw "github.com/spcent/plumego/middleware"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
)

const (
	// Default timeout buffer limit: 10MB
	defaultTimeoutMaxBytes = 10 << 20
	// Responses larger than this are rejected because timeout middleware must
	// buffer the complete response before it can safely decide whether to replay it.
	streamingThresholdBytes = 512 << 10 // 512KB
)

// TimeoutConfig customizes timeout middleware behavior.
//
// Timeout middleware enforces a maximum duration for request processing.
// If the downstream handler does not complete before the deadline, the request
// context is canceled and a 504 Gateway Timeout response is returned.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/timeout"
//
//	// Explicit configuration
//	config := timeout.TimeoutConfig{
//		Timeout:            10 * time.Second,
//		MaxBufferBytes:     5 << 20,      // 5MB max buffer
//		StreamingThreshold: 1 << 20,      // 1MB streaming threshold
//	}
//	handler := timeout.Timeout(config)(myHandler)
//
// The middleware buffers responses to allow timeout enforcement. Responses that
// exceed StreamingThreshold are rejected with a structured 500 before any
// downstream bytes are committed to the client.
//
// When a timeout occurs, it returns a 504 Gateway Timeout response with a
// structured error message.
type TimeoutConfig struct {
	// Timeout is the maximum duration for request processing
	Timeout time.Duration

	// MaxBufferBytes is the maximum response size to buffer for timeout enforcement
	// Responses larger than this will bypass buffering
	MaxBufferBytes int

	// StreamingThreshold is the maximum response size that timeout middleware will
	// buffer for replay. Larger responses are rejected because they cannot be
	// safely replayed after timeout arbitration.
	StreamingThreshold int
}

// Timeout creates a timeout middleware with explicit configuration.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	config := timeout.TimeoutConfig{
//		Timeout:            10 * time.Second,
//		MaxBufferBytes:     5 << 20,      // 5MB max buffer
//		StreamingThreshold: 1 << 20,      // 1MB streaming threshold
//	}
//	handler := timeout.Timeout(config)(myHandler)
func Timeout(cfg TimeoutConfig) mw.Middleware {
	if cfg.Timeout <= 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Ensure reasonable defaults
	if cfg.MaxBufferBytes <= 0 {
		cfg.MaxBufferBytes = defaultTimeoutMaxBytes
	}
	if cfg.StreamingThreshold <= 0 {
		cfg.StreamingThreshold = streamingThresholdBytes
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), cfg.Timeout)
			defer cancel()

			r = r.WithContext(ctx)

			tw := newTimeoutResponseWriter(ctx, cfg)
			done := make(chan struct{})

			go func() {
				defer close(done)
				next.ServeHTTP(tw, r)
			}()

			select {
			case <-done:
				if tw.Overflowed() {
					mw.WriteTransportError(w, r, http.StatusInternalServerError, contract.CodeInternalError, "response exceeded buffer limit", contract.CategoryServer, nil)
					return
				}
				if tw.TooLarge() {
					mw.WriteTransportError(w, r, http.StatusInternalServerError, contract.CodeInternalError, "response exceeded timeout buffering threshold", contract.CategoryServer, nil)
					return
				}
				tw.WriteTo(w)
			case <-ctx.Done():
				mw.WriteTransportError(w, r, http.StatusGatewayTimeout, contract.CodeTimeout, "request timed out", contract.CategoryTimeout, nil)
			}
		})
	}
}

func newTimeoutResponseWriter(ctx context.Context, cfg TimeoutConfig) *timeoutResponseWriter {
	return &timeoutResponseWriter{
		ctx:       ctx,
		cfg:       cfg,
		buffer:    internaltransport.NewBufferedResponse(cfg.MaxBufferBytes),
		buffering: true,
	}
}

type timeoutResponseWriter struct {
	ctx       context.Context
	cfg       TimeoutConfig
	buffer    *internaltransport.BufferedResponse
	overflow  bool
	tooLarge  bool
	buffering bool
}

func (w *timeoutResponseWriter) Header() http.Header {
	return w.buffer.Header()
}

func (w *timeoutResponseWriter) WriteHeader(statusCode int) {
	w.buffer.WriteHeader(statusCode)
}

func (w *timeoutResponseWriter) Write(p []byte) (int, error) {
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	default:
	}

	if w.overflow {
		return 0, http.ErrBodyNotAllowed
	}

	if w.tooLarge {
		return 0, http.ErrBodyNotAllowed
	}

	if w.buffering {
		currentSize := w.buffer.Len() + len(p)

		if currentSize > w.cfg.StreamingThreshold {
			w.buffering = false
			w.tooLarge = true
			w.buffer.ClearBody()
			return 0, http.ErrBodyNotAllowed
		}

		n, err := w.buffer.Write(p)
		if err != nil {
			w.overflow = true
		}
		return n, err
	}

	return 0, http.ErrBodyNotAllowed
}

func (w *timeoutResponseWriter) WriteTo(dst http.ResponseWriter) {
	if w.tooLarge || w.overflow {
		return
	}

	_, _ = w.buffer.WriteTo(dst)
}

func (w *timeoutResponseWriter) Overflowed() bool {
	return w.overflow
}

func (w *timeoutResponseWriter) TooLarge() bool {
	return w.tooLarge
}
