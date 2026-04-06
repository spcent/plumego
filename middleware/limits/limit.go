package limits

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	mw "github.com/spcent/plumego/middleware"
)

var errRequestTooLarge = errors.New("request body too large")

// BodyLimit enforces a maximum request body size using a protective reader that
// surfaces a structured error to the client instead of the default plaintext
// response from http.MaxBytesReader.
//
// This middleware is useful for preventing denial-of-service attacks that send
// large request bodies to exhaust server resources.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/limits"
//
//	// Limit request body to 10MB
//	handler := limits.BodyLimit(10<<20, nil)(myHandler)
//
//	// With logging
//	logger := log.NewGLogger()
//	handler := limits.BodyLimit(10<<20, logger)(myHandler)
//
// When a request body exceeds the limit, it returns a 413 Request Entity Too Large
// response with a structured error message containing the limit details.
func BodyLimit(maxBytes int64, logger log.StructuredLogger) mw.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			limited := &limitedBodyReader{
				w:        w,
				r:        r.Body,
				maxBytes: maxBytes,
				logger:   logger,
				now:      time.Now,
			}
			r.Body = limited

			next.ServeHTTP(w, r)
		})
	}
}

type limitedBodyReader struct {
	w        http.ResponseWriter
	r        io.ReadCloser
	maxBytes int64
	used     int64
	exceeded bool
	logger   log.StructuredLogger
	now      func() time.Time
}

func (l *limitedBodyReader) Read(p []byte) (int, error) {
	if l.exceeded {
		return 0, errRequestTooLarge
	}

	remaining := l.maxBytes - l.used
	if remaining <= 0 {
		return l.fail()
	}

	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err := l.r.Read(p)
	l.used += int64(n)

	if l.used > l.maxBytes {
		return l.fail()
	}

	return n, err
}

func (l *limitedBodyReader) Close() error {
	return l.r.Close()
}

func (l *limitedBodyReader) fail() (int, error) {
	if !l.exceeded {
		l.exceeded = true
		mw.WriteTransportError(l.w, nil, http.StatusRequestEntityTooLarge, mw.CodeRequestBodyTooLarge, "request body exceeds configured limit", contract.CategoryClient, map[string]any{
			"max_bytes":  l.maxBytes,
			"seen_bytes": l.used,
			"at":         l.now().UTC(),
		})
		if l.logger != nil {
			fields := contract.NewObservabilityPolicy().MiddlewareLogFields(nil, http.StatusRequestEntityTooLarge, 0)
			fields["max_bytes"] = l.maxBytes
			fields["seen_bytes"] = l.used
			l.logger.WithFields(log.Fields(contract.NewObservabilityPolicy().RedactFields(fields))).Warn("request body too large")
		}
	}

	return 0, errRequestTooLarge
}

// ConcurrencyLimit restricts how many requests can be processed concurrently
// and optionally bounds how many can wait in a queue. Requests that cannot
// enter the queue within the configured timeout receive a 503 response.
func ConcurrencyLimit(maxConcurrent, queueDepth int, queueTimeout time.Duration, logger log.StructuredLogger) mw.Middleware {
	if maxConcurrent <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	if queueDepth < maxConcurrent {
		queueDepth = maxConcurrent
	}

	sem := make(chan struct{}, maxConcurrent)
	queue := make(chan struct{}, queueDepth)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timeout := queueTimeout
			if timeout <= 0 {
				timeout = 100 * time.Millisecond
			}

			select {
			case queue <- struct{}{}:
				defer func() { <-queue }()
			default:
				mw.WriteTransportError(w, r, http.StatusServiceUnavailable, mw.CodeServerBusy, "server is throttling concurrent requests", contract.CategoryServer, nil)
				return
			}

			timer := time.NewTimer(timeout)
			defer timer.Stop()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-timer.C:
				mw.WriteTransportError(w, r, http.StatusServiceUnavailable, mw.CodeServerQueueTimeout, "request timed out waiting for an available worker", contract.CategoryServer, map[string]any{"queue_depth": len(queue)})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
