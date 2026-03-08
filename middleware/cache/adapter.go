// Package cache provides a thin HTTP adapter for gateway cache capabilities.
package cache

import gatewaycache "github.com/spcent/plumego/gateway/cache"

type (
	Store                  = gatewaycache.Store
	StoreStats             = gatewaycache.StoreStats
	KeyStrategy            = gatewaycache.KeyStrategy
	CachedResponse         = gatewaycache.CachedResponse
	Config                 = gatewaycache.Config
	DefaultKeyStrategy     = gatewaycache.DefaultKeyStrategy
	CustomKeyStrategy      = gatewaycache.CustomKeyStrategy
	URLOnlyKeyStrategy     = gatewaycache.URLOnlyKeyStrategy
	HeaderAwareKeyStrategy = gatewaycache.HeaderAwareKeyStrategy
	QueryParamKeyStrategy  = gatewaycache.QueryParamKeyStrategy
	UserAwareKeyStrategy   = gatewaycache.UserAwareKeyStrategy
	TenantAwareKeyStrategy = gatewaycache.TenantAwareKeyStrategy
	CompositeKeyStrategy   = gatewaycache.CompositeKeyStrategy
	PathPatternKeyStrategy = gatewaycache.PathPatternKeyStrategy
	NoCacheKeyStrategy     = gatewaycache.NoCacheKeyStrategy
	MemoryStore            = gatewaycache.MemoryStore
)

var (
	Middleware                = gatewaycache.Middleware
	NewDefaultKeyStrategy     = gatewaycache.NewDefaultKeyStrategy
	Stats                     = gatewaycache.Stats
	NewURLOnlyKeyStrategy     = gatewaycache.NewURLOnlyKeyStrategy
	NewHeaderAwareKeyStrategy = gatewaycache.NewHeaderAwareKeyStrategy
	NewQueryParamKeyStrategy  = gatewaycache.NewQueryParamKeyStrategy
	NewUserAwareKeyStrategy   = gatewaycache.NewUserAwareKeyStrategy
	NewTenantAwareKeyStrategy = gatewaycache.NewTenantAwareKeyStrategy
	NewCompositeKeyStrategy   = gatewaycache.NewCompositeKeyStrategy
	NewPathPatternKeyStrategy = gatewaycache.NewPathPatternKeyStrategy
	NewNoCacheKeyStrategy     = gatewaycache.NewNoCacheKeyStrategy
	NewMemoryStore            = gatewaycache.NewMemoryStore
)
