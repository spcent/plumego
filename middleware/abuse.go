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
type AbuseGuardConfig struct {
	Rate            float64
	Capacity        int
	CleanupInterval time.Duration
	MaxIdle         time.Duration
	Limiter         *abuse.Limiter
	KeyFunc         func(*http.Request) string
	Skip            func(*http.Request) bool
	IncludeHeaders  *bool
	Logger          log.StructuredLogger
}

// DefaultAbuseGuardConfig returns baseline settings for abuse protection.
func DefaultAbuseGuardConfig() AbuseGuardConfig {
	defaults := abuse.DefaultConfig()
	includeHeaders := true
	return AbuseGuardConfig{
		Rate:            defaults.Rate,
		Capacity:        defaults.Capacity,
		CleanupInterval: defaults.CleanupInterval,
		MaxIdle:         defaults.MaxIdle,
		IncludeHeaders:  &includeHeaders,
	}
}

// AbuseGuard applies per-key rate limiting to defend against abuse.
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
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = defaults.CleanupInterval
	}
	if config.MaxIdle <= 0 {
		config.MaxIdle = defaults.MaxIdle
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
			CleanupInterval: config.CleanupInterval,
			MaxIdle:         config.MaxIdle,
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
