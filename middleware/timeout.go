package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
)

const (
	// Default timeout buffer limit: 10MB
	defaultTimeoutMaxBytes = 10 << 20
	// Streaming threshold: responses larger than this will bypass buffering
	// to avoid memory spikes in streaming/large response scenarios
	streamingThresholdBytes = 512 << 10 // 512KB
)

// TimeoutConfig customizes timeout middleware behavior.
type TimeoutConfig struct {
	Timeout        time.Duration
	MaxBufferBytes int
	// StreamingThreshold specifies when to bypass buffering for large/streaming responses
	StreamingThreshold int
}

// Timeout creates a middleware that enforces a maximum duration for a request.
// If the downstream handler does not complete before the deadline, the request
// context is canceled and a 504 Gateway Timeout response is returned.
func Timeout(d time.Duration) Middleware {
	return TimeoutWithConfig(TimeoutConfig{
		Timeout:            d,
		MaxBufferBytes:     defaultTimeoutMaxBytes,
		StreamingThreshold: streamingThresholdBytes,
	})
}

// TimeoutWithConfig creates a timeout middleware with explicit configuration.
func TimeoutWithConfig(cfg TimeoutConfig) Middleware {
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
					http.Error(w, "response exceeded buffer limit", http.StatusInternalServerError)
					return
				}
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

func newTimeoutResponseWriter(ctx context.Context, cfg TimeoutConfig) *timeoutResponseWriter {
	return &timeoutResponseWriter{
		header:    http.Header{},
		ctx:       ctx,
		cfg:       cfg,
		buffering: true,
	}
}

type timeoutResponseWriter struct {
	header     http.Header
	body       []byte
	status     int
	ctx        context.Context
	cfg        TimeoutConfig
	overflow   bool
	buffering  bool // Whether currently buffering
	bypassUsed bool // Whether bypass mode was triggered
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

	if w.overflow {
		return 0, http.ErrBodyNotAllowed
	}

	if w.status == 0 {
		w.status = http.StatusOK
	}

	// If bypass was triggered, discard data but don't error
	if w.bypassUsed {
		return len(p), nil
	}

	// Check if we should switch to bypass mode
	if w.buffering {
		currentSize := len(w.body) + len(p)

		// If exceeds streaming threshold, switch to bypass mode
		if currentSize > w.cfg.StreamingThreshold {
			w.buffering = false
			w.bypassUsed = true
			w.body = nil // Free memory
			return len(p), nil
		}

		// If exceeds max buffer limit, mark overflow
		if w.cfg.MaxBufferBytes > 0 && currentSize > w.cfg.MaxBufferBytes {
			w.overflow = true
			return 0, http.ErrBodyNotAllowed
		}

		// Continue buffering
		w.body = append(w.body, p...)
		return len(p), nil
	}

	// Bypass mode (shouldn't reach here due to bypassUsed check)
	return len(p), nil
}

func (w *timeoutResponseWriter) WriteTo(dst http.ResponseWriter) {
	// If bypass mode was used, we cannot replay the response
	if w.bypassUsed {
		http.Error(dst, "response too large for timeout buffering", http.StatusInternalServerError)
		return
	}

	// Normal buffered response
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

func (w *timeoutResponseWriter) Overflowed() bool {
	return w.overflow
}
