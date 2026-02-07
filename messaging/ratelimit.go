package messaging

import (
	"context"
	"sync"
	"time"
)

// RateLimiter limits the throughput to a provider to avoid exceeding
// external API quotas. It uses a simple token-bucket algorithm.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   int
	max      int
	interval time.Duration
	last     time.Time
}

// NewRateLimiter creates a limiter that allows maxPerInterval calls
// per interval. For example NewRateLimiter(100, time.Second) allows
// 100 calls per second.
func NewRateLimiter(maxPerInterval int, interval time.Duration) *RateLimiter {
	if maxPerInterval <= 0 {
		maxPerInterval = 100
	}
	if interval <= 0 {
		interval = time.Second
	}
	return &RateLimiter{
		tokens:   maxPerInterval,
		max:      maxPerInterval,
		interval: interval,
		last:     time.Now(),
	}
}

// Wait blocks until a token is available or ctx is cancelled.
func (l *RateLimiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		l.refill()
		if l.tokens > 0 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.interval / time.Duration(l.max+1)):
		}
	}
}

func (l *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.last)
	if elapsed < l.interval {
		return
	}
	periods := int(elapsed / l.interval)
	l.tokens += periods * l.max
	if l.tokens > l.max {
		l.tokens = l.max
	}
	l.last = l.last.Add(time.Duration(periods) * l.interval)
}

// rateLimitedSend waits for a token then delegates to the inner send function.
// Used by both RateLimitedSMSProvider and RateLimitedEmailProvider to avoid
// duplicating the wait-then-send pattern.
func rateLimitedSend[M any, R any](ctx context.Context, limiter *RateLimiter, inner func(context.Context, M) (*R, error), msg M) (*R, error) {
	if err := limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return inner(ctx, msg)
}

// RateLimitedSMSProvider wraps an SMSProvider with rate limiting.
type RateLimitedSMSProvider struct {
	inner   SMSProvider
	limiter *RateLimiter
}

// NewRateLimitedSMSProvider wraps provider with rate limiting.
func NewRateLimitedSMSProvider(inner SMSProvider, maxPerSecond int) *RateLimitedSMSProvider {
	return &RateLimitedSMSProvider{
		inner:   inner,
		limiter: NewRateLimiter(maxPerSecond, time.Second),
	}
}

func (p *RateLimitedSMSProvider) Name() string { return p.inner.Name() }

func (p *RateLimitedSMSProvider) Send(ctx context.Context, msg SMSMessage) (*SMSResult, error) {
	return rateLimitedSend(ctx, p.limiter, p.inner.Send, msg)
}

// RateLimitedEmailProvider wraps an EmailProvider with rate limiting.
type RateLimitedEmailProvider struct {
	inner   EmailProvider
	limiter *RateLimiter
}

// NewRateLimitedEmailProvider wraps provider with rate limiting.
func NewRateLimitedEmailProvider(inner EmailProvider, maxPerSecond int) *RateLimitedEmailProvider {
	return &RateLimitedEmailProvider{
		inner:   inner,
		limiter: NewRateLimiter(maxPerSecond, time.Second),
	}
}

func (p *RateLimitedEmailProvider) Name() string { return p.inner.Name() }

func (p *RateLimitedEmailProvider) Send(ctx context.Context, msg EmailMessage) (*EmailResult, error) {
	return rateLimitedSend(ctx, p.limiter, p.inner.Send, msg)
}
