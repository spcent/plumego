package tenant

import (
	"context"
	"time"
)

// Hooks provides optional callbacks for audit/metrics hooks.
type Hooks struct {
	OnResolve func(ctx context.Context, info ResolveInfo)
	OnPolicy  func(ctx context.Context, info PolicyDecision)
	OnQuota   func(ctx context.Context, info QuotaDecision)
}

// ResolveInfo describes how a tenant id was resolved.
type ResolveInfo struct {
	TenantID string
	Source   string
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
