package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/abuse"
)

// AbuseGuardConfig configures the abuse guard middleware.
//
// AbuseGuard provides per-key rate limiting to defend against abuse attacks.
// It uses a token bucket algorithm to limit the rate of requests from each client.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	config := middleware.AbuseGuardConfig{
//		Rate:            10.0,      // 10 requests per second
//		Capacity:        100,       // Burst capacity of 100 requests
//		CleanupInterval: time.Minute, // Clean up idle entries every minute
//		MaxIdle:         5 * time.Minute, // Remove entries idle for 5 minutes
//		KeyFunc:         middleware.ClientIPKey, // Use client IP as key
//	}
//	handler := middleware.AbuseGuard(config)(myHandler)
type AbuseGuardConfig struct {
	// Rate is the number of requests per second allowed per key
	Rate float64

	// Capacity is the maximum burst size (number of tokens in the bucket)
	Capacity int

	// MaxEntries is the maximum number of tracked keys before eviction kicks in
	MaxEntries int

	// CleanupInterval is how often to clean up idle entries
	CleanupInterval time.Duration

	// MaxIdle is the maximum time an entry can be idle before being removed
	MaxIdle time.Duration

	// Shards is the number of lock shards for buckets
	Shards int

	// Limiter is a custom limiter instance (optional)
	Limiter *abuse.Limiter

	// KeyFunc extracts a rate limiting key from the request (e.g., client IP)
	// Default: clientIPKey (uses X-Forwarded-For, X-Real-IP, or RemoteAddr)
	KeyFunc func(*http.Request) string

	// Skip determines whether to skip rate limiting for a request
	// Useful for whitelisting certain clients or endpoints
	Skip func(*http.Request) bool

	// IncludeHeaders controls whether to include rate limit headers in responses
	// Default: true
	IncludeHeaders *bool

	// Logger is used for logging rate limit violations
	Logger log.StructuredLogger
}

// DefaultAbuseGuardConfig returns baseline settings for abuse protection.
func DefaultAbuseGuardConfig() AbuseGuardConfig {
	defaults := abuse.DefaultConfig()
	includeHeaders := true
	return AbuseGuardConfig{
		Rate:            defaults.Rate,
		Capacity:        defaults.Capacity,
		MaxEntries:      defaults.MaxEntries,
		CleanupInterval: defaults.CleanupInterval,
		MaxIdle:         defaults.MaxIdle,
		Shards:          defaults.Shards,
		IncludeHeaders:  &includeHeaders,
	}
}

// AbuseGuard applies per-key rate limiting to defend against abuse.
//
// AbuseGuard uses a token bucket algorithm to limit the rate of requests from each client.
// It tracks requests per key (default: client IP) and rejects requests when the limit is exceeded.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	// Create middleware with default settings
//	handler := middleware.AbuseGuard(middleware.DefaultAbuseGuardConfig())(myHandler)
//
//	// Or with custom configuration
//	config := middleware.AbuseGuardConfig{
//		Rate:     5.0,  // 5 requests per second
//		Capacity: 10,   // Burst capacity of 10
//	}
//	handler := middleware.AbuseGuard(config)(myHandler)
//
// The middleware adds the following headers to responses:
//   - X-RateLimit-Limit: The maximum number of requests allowed
//   - X-RateLimit-Remaining: The number of requests remaining in the current window
//   - X-RateLimit-Reset: The Unix timestamp when the rate limit resets
//   - Retry-After: The number of seconds to wait before retrying (when rate limited)
//
// When a request is rate limited, it returns a 429 Too Many Requests response with
// a structured error message containing the limit details.
func AbuseGuard(config AbuseGuardConfig) Middleware {
	defaults := DefaultAbuseGuardConfig()
	includeHeaders := true
	if defaults.IncludeHeaders != nil {
		includeHeaders = *defaults.IncludeHeaders
	}

	if config.Rate <= 0 {
		config.Rate = defaults.Rate
	}
	if config.Capacity <= 0 {
		config.Capacity = defaults.Capacity
	}
	if config.MaxEntries <= 0 {
		config.MaxEntries = defaults.MaxEntries
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = defaults.CleanupInterval
	}
	if config.MaxIdle <= 0 {
		config.MaxIdle = defaults.MaxIdle
	}
	if config.Shards <= 0 {
		config.Shards = defaults.Shards
	}
	if config.KeyFunc == nil {
		config.KeyFunc = clientIPKey
	}
	if config.IncludeHeaders != nil {
		includeHeaders = *config.IncludeHeaders
	}

	limiter := config.Limiter
	if limiter == nil {
		limiter = abuse.NewLimiter(abuse.Config{
			Rate:            config.Rate,
			Capacity:        config.Capacity,
			MaxEntries:      config.MaxEntries,
			CleanupInterval: config.CleanupInterval,
			MaxIdle:         config.MaxIdle,
			Shards:          config.Shards,
		})
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skip != nil && config.Skip(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := strings.TrimSpace(config.KeyFunc(r))
			decision := limiter.Allow(key)
			if includeHeaders {
				applyRateLimitHeaders(w, decision)
			}

			if !decision.Allowed {
				writeAbuseError(w, r, decision, config.Logger)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func applyRateLimitHeaders(w http.ResponseWriter, decision abuse.Decision) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(decision.Limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
	if !decision.Reset.IsZero() {
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(decision.Reset.Unix(), 10))
	}
	if !decision.Allowed && decision.RetryAfter > 0 {
		seconds := int(math.Ceil(decision.RetryAfter.Seconds()))
		if seconds < 1 {
			seconds = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(seconds))
	}
}

func writeAbuseError(w http.ResponseWriter, r *http.Request, decision abuse.Decision, logger log.StructuredLogger) {
	contract.WriteError(w, r, contract.APIError{
		Status:   http.StatusTooManyRequests,
		Code:     "rate_limited",
		Category: contract.CategoryRateLimit,
		Message:  "too many requests",
		Details: map[string]any{
			"limit":       decision.Limit,
			"remaining":   decision.Remaining,
			"retry_after": decision.RetryAfter.Seconds(),
			"reset_at":    decision.Reset.UTC(),
		},
	})

	if logger != nil {
		logger.WithFields(log.Fields{
			"limit":     decision.Limit,
			"remaining": decision.Remaining,
		}).Warn("request rate limited", nil)
	}
}

func clientIPKey(r *http.Request) string {
	if r == nil {
		return ""
	}

	if ip := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}
