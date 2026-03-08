package tenant

import (
	"context"
	"time"
)

// Hooks provides optional callbacks for observability and audit integration.
type Hooks struct {
	OnResolve   func(ctx context.Context, info ResolveInfo)
	OnPolicy    func(ctx context.Context, info PolicyDecision)
	OnQuota     func(ctx context.Context, info QuotaDecision)
	OnRateLimit func(ctx context.Context, info RateLimitDecision)
}

// ResolveInfo describes how a tenant id was resolved.
type ResolveInfo struct {
	TenantID string
	Source   string // "principal", "header", "extractor"
}

// PolicyDecision describes a policy evaluation outcome.
type PolicyDecision struct {
	TenantID string
	Allowed  bool
	Reason   string
	Model    string
	Tool     string
	Method   string
	Path     string
	Status   int
	// Err is non-nil when the decision was caused by a backend error rather than
	// a policy rule, allowing observers to distinguish system failures from
	// intentional denials.
	Err error
}

// QuotaDecision describes a quota evaluation outcome.
type QuotaDecision struct {
	TenantID          string
	Allowed           bool
	Tokens            int
	Requests          int
	RemainingTokens   int
	RemainingRequests int
	RetryAfter        time.Duration
	Status            int
	// Err is non-nil when the decision was caused by a backend error rather than
	// an explicit quota limit, allowing observers to distinguish system failures
	// from intentional throttling.
	Err error
}

// RateLimitDecision describes a rate limit evaluation outcome.
type RateLimitDecision struct {
	TenantID   string
	Allowed    bool
	Tokens     int64
	Remaining  int64
	Limit      int64
	Burst      int64
	RetryAfter time.Duration
	Status     int
	// Err is non-nil when the decision was caused by a backend error rather than
	// an explicit rate limit, allowing observers to distinguish system failures
	// from intentional throttling.
	Err error
}

// Resolve invokes the resolve hook when configured.
func (h Hooks) Resolve(ctx context.Context, info ResolveInfo) {
	if h.OnResolve != nil {
		h.OnResolve(ctx, info)
	}
}

// Policy invokes the policy hook when configured.
func (h Hooks) Policy(ctx context.Context, info PolicyDecision) {
	if h.OnPolicy != nil {
		h.OnPolicy(ctx, info)
	}
}

// Quota invokes the quota hook when configured.
func (h Hooks) Quota(ctx context.Context, info QuotaDecision) {
	if h.OnQuota != nil {
		h.OnQuota(ctx, info)
	}
}

// RateLimit invokes the rate limit hook when configured.
func (h Hooks) RateLimit(ctx context.Context, info RateLimitDecision) {
	if h.OnRateLimit != nil {
		h.OnRateLimit(ctx, info)
	}
}
