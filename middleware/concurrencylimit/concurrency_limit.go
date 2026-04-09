package concurrencylimit

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	mw "github.com/spcent/plumego/middleware"
)

// Middleware restricts how many requests can be processed concurrently and
// optionally bounds how many can wait in a queue.
func Middleware(maxConcurrent, queueDepth int, queueTimeout time.Duration, logger log.StructuredLogger) mw.Middleware {
	_ = logger
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
