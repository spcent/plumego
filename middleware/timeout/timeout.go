package timeout

import (
	"context"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
	mw "github.com/spcent/plumego/middleware"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
)

const (
	// Default timeout buffer limit: 10MB
	defaultTimeoutMaxBytes = 10 << 20
	// Streaming threshold: responses larger than this will bypass buffering
	// to avoid memory spikes in streaming/large response scenarios
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
// The middleware buffers responses to allow timeout enforcement, but switches to
// bypass mode for large/streaming responses to avoid memory spikes.
//
// When a timeout occurs, it returns a 504 Gateway Timeout response with a
// structured error message.
type TimeoutConfig struct {
	// Timeout is the maximum duration for request processing
	Timeout time.Duration

	// MaxBufferBytes is the maximum response size to buffer for timeout enforcement
	// Responses larger than this will bypass buffering
	MaxBufferBytes int

	// StreamingThreshold specifies when to bypass buffering for large/streaming responses
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
func Timeout(cfg TimeoutConfig) middleware.Middleware {
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
					_ = contract.WriteError(w, r, contract.NewErrorBuilder().
						Type(contract.TypeInternal).
						Message("response exceeded buffer limit").
						Build())
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
	ctx        context.Context
	cfg        TimeoutConfig
	buffer     *internaltransport.BufferedResponse
	overflow   bool
	buffering  bool // Whether currently buffering
	bypassUsed bool // Whether bypass mode was triggered
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

	// If bypass was triggered, discard data but don't error
	if w.bypassUsed {
		return len(p), nil
	}

	// Check if we should switch to bypass mode
	if w.buffering {
		currentSize := w.buffer.Len() + len(p)

		// If exceeds streaming threshold, switch to bypass mode
		if currentSize > w.cfg.StreamingThreshold {
			w.buffering = false
			w.bypassUsed = true
			w.buffer.ClearBody() // Free memory
			return len(p), nil
		}

		// Continue buffering
		n, err := w.buffer.Write(p)
		if err != nil {
			w.overflow = true
		}
		return n, err
	}

	// Bypass mode (shouldn't reach here due to bypassUsed check)
	return len(p), nil
}

func (w *timeoutResponseWriter) WriteTo(dst http.ResponseWriter) {
	// If bypass mode was used, we cannot replay the response
	if w.bypassUsed {
		_ = contract.WriteError(dst, nil, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("response too large for timeout buffering").
			Build())
		return
	}

	// Normal buffered response
	_, _ = w.buffer.WriteTo(dst)
}

func (w *timeoutResponseWriter) Overflowed() bool {
	return w.overflow
}
