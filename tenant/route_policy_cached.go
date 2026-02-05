package tenant

import "context"

// CachedRoutePolicyProvider wraps a provider with a cache.
type CachedRoutePolicyProvider struct {
	Provider RoutePolicyProvider
	Cache    RoutePolicyCache
}

// NewCachedRoutePolicyProvider creates a cached provider.
func NewCachedRoutePolicyProvider(provider RoutePolicyProvider, cache RoutePolicyCache) *CachedRoutePolicyProvider {
	return &CachedRoutePolicyProvider{
		Provider: provider,
		Cache:    cache,
	}
}

// RoutePolicy resolves policy using cache before hitting provider.
func (c *CachedRoutePolicyProvider) RoutePolicy(ctx context.Context, tenantID string) (RoutePolicy, error) {
	if c == nil || c.Provider == nil {
		return RoutePolicy{}, ErrRoutePolicyNotFound
	}

	if c.Cache != nil {
		if policy, ok := c.Cache.Get(ctx, tenantID); ok {
			return policy, nil
		}
	}

	policy, err := c.Provider.RoutePolicy(ctx, tenantID)
	if err != nil {
		return RoutePolicy{}, err
	}

	if c.Cache != nil {
		_ = c.Cache.Set(ctx, tenantID, policy)
	}

	return policy, nil
}

// Invalidate removes a cached policy.
func (c *CachedRoutePolicyProvider) Invalidate(ctx context.Context, tenantID string) error {
	if c == nil || c.Cache == nil {
		return nil
	}
	return c.Cache.Delete(ctx, tenantID)
}
