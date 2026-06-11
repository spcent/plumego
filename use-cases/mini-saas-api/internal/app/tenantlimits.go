package app

import (
	"context"

	tenantcore "github.com/spcent/plumego/x/tenant/core"
)

// uniformRateLimits serves the same token-bucket configuration for every
// tenant; enforcement state (the buckets) is still per tenant inside
// tenantcore.TokenBucketRateLimiter. Per-tenant overrides would replace this
// with a store-backed provider.
type uniformRateLimits struct {
	rps   int64
	burst int64
}

func (p uniformRateLimits) RateLimitConfig(_ context.Context, _ string) (tenantcore.RateLimitConfig, error) {
	return tenantcore.RateLimitConfig{RequestsPerSecond: p.rps, Burst: p.burst}, nil
}

// uniformQuota serves the same fixed-window request quota for every tenant;
// window counters are per tenant inside tenantcore.FixedWindowQuotaManager.
type uniformQuota struct {
	requestsPerMinute int64
}

func (p uniformQuota) QuotaConfig(_ context.Context, _ string) (tenantcore.QuotaConfig, error) {
	return tenantcore.QuotaConfig{
		Limits: []tenantcore.QuotaLimit{
			{Window: tenantcore.QuotaWindowMinute, Requests: p.requestsPerMinute},
		},
	}, nil
}
