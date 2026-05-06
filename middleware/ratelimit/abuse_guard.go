// Package ratelimit adapts stable abuse primitives to HTTP middleware.
//
// NewAbuseGuard is the production entrypoint when middleware creates limiter
// resources because it exposes Stop for application shutdown. AbuseGuard remains
// a compatibility convenience constructor for source-stable middleware wiring.
package ratelimit

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	mw "github.com/spcent/plumego/middleware"
	internalobs "github.com/spcent/plumego/middleware/internal/observability"
	internaltransport "github.com/spcent/plumego/middleware/internal/transport"
	"github.com/spcent/plumego/security/abuse"
)

const (
	headerRateLimitLimit     = "X-RateLimit-Limit"
	headerRateLimitRemaining = "X-RateLimit-Remaining"
	headerRateLimitReset     = "X-RateLimit-Reset"
	headerRetryAfter         = "Retry-After"
)

// AbuseGuardConfig configures the abuse guard middleware.
//
// AbuseGuard provides per-key rate limiting to defend against abuse attacks.
// It uses a token bucket algorithm to limit the rate of requests from each client.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/ratelimit"
//
//	config := ratelimit.AbuseGuardConfig{
//		Rate:            10.0,      // 10 requests per second
//		Capacity:        100,       // Burst capacity of 100 requests
//		CleanupInterval: time.Minute, // Clean up idle entries every minute
//		MaxIdle:         5 * time.Minute, // Remove entries idle for 5 minutes
//		// KeyFunc defaults to RemoteAddr when omitted
//	}
//	guard := ratelimit.NewAbuseGuard(config)
//	defer guard.Stop()
//	handler := guard.Middleware()(myHandler)
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

	// KeyFunc extracts a rate limiting key from the request (e.g., client IP).
	// Default: the direct RemoteAddr peer IP. Applications behind trusted
	// proxies may opt into forwarded headers by setting KeyFunc explicitly.
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

// AbuseGuardMiddleware owns the runtime state for abuse guard middleware.
//
// Use NewAbuseGuard when the middleware creates the limiter so callers can call
// Stop during application shutdown. If Config.Limiter is injected, Stop is a
// no-op because the caller owns that limiter lifecycle.
type AbuseGuardMiddleware struct {
	state *abuseGuardState
}

type abuseGuardState struct {
	config         AbuseGuardConfig
	limiter        *abuse.Limiter
	includeHeaders bool
	ownsLimiter    bool
	stopOnce       sync.Once
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

// NewAbuseGuard constructs abuse guard middleware with an explicit lifecycle.
// Call Stop during application shutdown when no Limiter was supplied.
func NewAbuseGuard(config AbuseGuardConfig) *AbuseGuardMiddleware {
	config, includeHeaders := normalizeConfig(config)

	limiter := config.Limiter
	ownsLimiter := false
	if limiter == nil {
		limiter = abuse.NewLimiter(abuse.Config{
			Rate:            config.Rate,
			Capacity:        config.Capacity,
			MaxEntries:      config.MaxEntries,
			CleanupInterval: config.CleanupInterval,
			MaxIdle:         config.MaxIdle,
			Shards:          config.Shards,
		})
		ownsLimiter = true
	}

	return &AbuseGuardMiddleware{state: &abuseGuardState{
		config:         config,
		limiter:        limiter,
		includeHeaders: includeHeaders,
		ownsLimiter:    ownsLimiter,
	}}
}

// Stop releases middleware-owned limiter resources. It is safe to call multiple
// times. Injected limiters remain caller-owned and are not stopped here.
func (g *AbuseGuardMiddleware) Stop() {
	if g == nil || g.state == nil || !g.state.ownsLimiter || g.state.limiter == nil {
		return
	}
	g.state.stopOnce.Do(func() {
		g.state.limiter.Stop()
	})
}

// Middleware returns the HTTP middleware function for this abuse guard.
func (g *AbuseGuardMiddleware) Middleware() mw.Middleware {
	if g == nil || g.state == nil || g.state.limiter == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	state := g.state

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if state.config.Skip != nil && state.config.Skip(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := strings.TrimSpace(state.config.KeyFunc(r))
			if key == "" {
				key = internaltransport.DirectClientIP(r)
			}
			decision := state.limiter.Allow(key)
			if state.includeHeaders {
				applyRateLimitHeaders(w, decision)
			}

			if !decision.Allowed {
				writeAbuseError(w, r, decision, state.config.Logger)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AbuseGuard applies per-key rate limiting to defend against abuse.
//
// AbuseGuard remains a compatibility convenience constructor for source-stable
// middleware wiring. When config.Limiter is nil it creates an internal limiter
// that cannot be stopped through this function; production code that lets
// middleware create limiter resources should use NewAbuseGuard, wire
// guard.Middleware(), and call guard.Stop() during application shutdown.
//
// AbuseGuard uses a token bucket algorithm to limit the rate of requests from
// each client. It tracks requests per key (default: client IP) and rejects
// requests when the limit is exceeded.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/ratelimit"
//
//	// Production lifecycle path
//	config := ratelimit.AbuseGuardConfig{
//		Rate:     5.0,  // 5 requests per second
//		Capacity: 10,   // Burst capacity of 10
//	}
//	guard := ratelimit.NewAbuseGuard(config)
//	defer guard.Stop()
//	handler := guard.Middleware()(myHandler)
//
//	// Compatibility path for injected limiter or short-lived wiring
//	handler = ratelimit.AbuseGuard(ratelimit.AbuseGuardConfig{Limiter: limiter})(myHandler)
//
// The middleware adds the following headers to responses:
//   - X-RateLimit-Limit: The maximum number of requests allowed
//   - X-RateLimit-Remaining: The number of requests remaining in the current window
//   - X-RateLimit-Reset: The Unix timestamp when the rate limit resets
//   - Retry-After: The number of seconds to wait before retrying (when rate limited)
//
// When a request is rate limited, it returns a 429 Too Many Requests response with
// a structured error message containing the limit details.
func AbuseGuard(config AbuseGuardConfig) mw.Middleware {
	return NewAbuseGuard(config).Middleware()
}

func normalizeConfig(config AbuseGuardConfig) (AbuseGuardConfig, bool) {
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
		config.KeyFunc = internaltransport.DirectClientIP
	}
	if config.IncludeHeaders != nil {
		includeHeaders = *config.IncludeHeaders
	}

	return config, includeHeaders
}

func applyRateLimitHeaders(w http.ResponseWriter, decision abuse.Decision) {
	w.Header().Set(headerRateLimitLimit, strconv.Itoa(decision.Limit))
	w.Header().Set(headerRateLimitRemaining, strconv.Itoa(decision.Remaining))
	if !decision.Reset.IsZero() {
		w.Header().Set(headerRateLimitReset, strconv.FormatInt(decision.Reset.Unix(), 10))
	}
	if !decision.Allowed && decision.RetryAfter > 0 {
		seconds := int(math.Ceil(decision.RetryAfter.Seconds()))
		if seconds < 1 {
			seconds = 1
		}
		w.Header().Set(headerRetryAfter, strconv.Itoa(seconds))
	}
}

func writeAbuseError(w http.ResponseWriter, r *http.Request, decision abuse.Decision, logger log.StructuredLogger) {
	mw.WriteTransportError(w, r, http.StatusTooManyRequests, contract.CodeRateLimited, "too many requests", contract.CategoryRateLimit, map[string]any{
		"limit":       decision.Limit,
		"remaining":   decision.Remaining,
		"retry_after": decision.RetryAfter.Seconds(),
		"reset_at":    decision.Reset.UTC(),
	})

	if logger != nil {
		fields := internalobs.MiddlewareLogFields(r, http.StatusTooManyRequests, 0)
		fields["limit"] = decision.Limit
		fields["remaining"] = decision.Remaining
		internalobs.RunSafeFinalizer(func() {
			logger.WithFields(log.Fields(internalobs.RedactFields(fields))).Warn("request rate limited")
		})
	}
}
