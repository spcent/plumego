package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	plog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/utils"
)

// AccessLogMiddleware provides structured logging for all requests
func AccessLogMiddleware(logger plog.StructuredLogger, collector metrics.MetricsCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(rw, r)

			// Record metrics
			duration := time.Since(start)

			// Log access
			logger.Info("access", plog.Fields{
				"method":        r.Method,
				"path":          r.URL.Path,
				"query":         r.URL.RawQuery,
				"status":        rw.statusCode,
				"duration_ms":   duration.Milliseconds(),
				"remote_addr":   r.RemoteAddr,
				"user_agent":    r.UserAgent(),
				"bytes_written": rw.bytesWritten,
			})

			// Record Prometheus metrics
			if collector != nil {
				collector.ObserveHTTP(
					context.Background(),
					r.Method,
					r.URL.Path,
					rw.statusCode,
					rw.bytesWritten,
					duration,
				)
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	wroteHeader  bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.wroteHeader {
		rw.statusCode = statusCode
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := utils.SafeWrite(rw.ResponseWriter, b)
	rw.bytesWritten += n
	return n, err
}

// RateLimitMiddleware implements gateway-level rate limiting
func RateLimitMiddleware(cfg RateLimitConfig) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter
	limiter := newTokenBucketLimiter(cfg.RequestsPerSecond, cfg.BurstSize)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				w.Header().Set("X-RateLimit-Limit", formatInt(cfg.RequestsPerSecond))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Retry-After", "1")

				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware adds timeout to requests
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Request completed
			case <-ctx.Done():
				// Timeout
				if ctx.Err() == context.DeadlineExceeded {
					http.Error(w, "Gateway timeout", http.StatusGatewayTimeout)
				}
			}
		})
	}
}

// tokenBucketLimiter implements a simple token bucket rate limiter
type tokenBucketLimiter struct {
	rate       float64
	tokens     float64
	maxTokens  float64
	lastUpdate time.Time
}

func newTokenBucketLimiter(requestsPerSecond int, burstSize int) *tokenBucketLimiter {
	return &tokenBucketLimiter{
		rate:       float64(requestsPerSecond),
		tokens:     float64(burstSize),
		maxTokens:  float64(burstSize),
		lastUpdate: time.Now(),
	}
}

func (l *tokenBucketLimiter) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(l.lastUpdate).Seconds()

	// Add tokens based on time elapsed
	l.tokens += elapsed * l.rate
	if l.tokens > l.maxTokens {
		l.tokens = l.maxTokens
	}

	l.lastUpdate = now

	// Try to consume a token
	if l.tokens >= 1 {
		l.tokens -= 1
		return true
	}

	return false
}

func formatInt(i int) string {
	return strconv.Itoa(i)
}
