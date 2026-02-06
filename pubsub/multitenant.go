package pubsub

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Quota errors
var (
	ErrQuotaExceeded      = errors.New("quota exceeded")
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrInvalidTenant      = errors.New("invalid tenant ID")
)

// TenantQuota defines resource limits for a tenant
type TenantQuota struct {
	// MaxPublishRate is the maximum messages per second
	MaxPublishRate int

	// MaxSubscriptions is the maximum number of active subscriptions
	MaxSubscriptions int

	// MaxTopics is the maximum number of topics the tenant can use
	MaxTopics int

	// MaxMessageSize is the maximum message size in bytes
	MaxMessageSize int

	// MaxBandwidth is the maximum bandwidth in bytes per second
	MaxBandwidth int64
}

// DefaultTenantQuota returns default quota settings
func DefaultTenantQuota() TenantQuota {
	return TenantQuota{
		MaxPublishRate:   1000,
		MaxSubscriptions: 100,
		MaxTopics:        1000,
		MaxMessageSize:   1024 * 1024, // 1 MB
		MaxBandwidth:     10 * 1024 * 1024, // 10 MB/s
	}
}

// UnlimitedQuota returns a quota with no limits
func UnlimitedQuota() TenantQuota {
	return TenantQuota{
		MaxPublishRate:   0, // 0 = unlimited
		MaxSubscriptions: 0,
		MaxTopics:        0,
		MaxMessageSize:   0,
		MaxBandwidth:     0,
	}
}

// TenantUsage tracks current resource usage
type TenantUsage struct {
	PublishCount     atomic.Uint64
	PublishRate      atomic.Uint64 // Per second
	ActiveSubs       atomic.Int64
	TopicsUsed       atomic.Int64
	BytesPublished   atomic.Uint64
	BytesPerSecond   atomic.Uint64
	QuotaViolations  atomic.Uint64
	LastPublish      time.Time
	LastPublishMu    sync.RWMutex
}

// TenantConfig holds tenant configuration
type TenantConfig struct {
	ID      string
	Quota   TenantQuota
	Enabled bool
}

// MultiTenantPubSub provides tenant isolation and quotas
type MultiTenantPubSub struct {
	ps *InProcPubSub

	// Tenant management
	tenants   map[string]*tenantData
	tenantsMu sync.RWMutex

	// Default quota for new tenants
	defaultQuota TenantQuota

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed atomic.Bool
}

// tenantData holds per-tenant state
type tenantData struct {
	config TenantConfig
	usage  TenantUsage

	// Rate limiting
	tokens       atomic.Int64
	lastRefill   time.Time
	lastRefillMu sync.Mutex

	// Bandwidth tracking
	bandwidthTokens   atomic.Int64
	lastBandwidthMu   sync.Mutex
	lastBandwidthTime time.Time
}

// NewMultiTenantPubSub creates a multi-tenant pubsub wrapper
func NewMultiTenantPubSub(ps *InProcPubSub, defaultQuota TenantQuota) *MultiTenantPubSub {
	ctx, cancel := context.WithCancel(context.Background())

	mtps := &MultiTenantPubSub{
		ps:           ps,
		tenants:      make(map[string]*tenantData),
		defaultQuota: defaultQuota,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start background workers
	mtps.wg.Add(1)
	go mtps.metricsCollector()

	return mtps
}

// CreateTenant creates a new tenant with specified quota
func (mtps *MultiTenantPubSub) CreateTenant(tenantID string, quota TenantQuota) error {
	if tenantID == "" {
		return ErrInvalidTenant
	}

	mtps.tenantsMu.Lock()
	defer mtps.tenantsMu.Unlock()

	if _, exists := mtps.tenants[tenantID]; exists {
		return ErrTenantAlreadyExists
	}

	now := time.Now()
	tenant := &tenantData{
		config: TenantConfig{
			ID:      tenantID,
			Quota:   quota,
			Enabled: true,
		},
		lastRefill:        now,
		lastBandwidthTime: now,
	}

	// Initialize token buckets
	tenant.tokens.Store(int64(quota.MaxPublishRate))
	tenant.bandwidthTokens.Store(quota.MaxBandwidth)

	mtps.tenants[tenantID] = tenant

	return nil
}

// DeleteTenant removes a tenant
func (mtps *MultiTenantPubSub) DeleteTenant(tenantID string) error {
	mtps.tenantsMu.Lock()
	defer mtps.tenantsMu.Unlock()

	if _, exists := mtps.tenants[tenantID]; !exists {
		return ErrTenantNotFound
	}

	delete(mtps.tenants, tenantID)
	return nil
}

// UpdateQuota updates a tenant's quota
func (mtps *MultiTenantPubSub) UpdateQuota(tenantID string, quota TenantQuota) error {
	mtps.tenantsMu.RLock()
	tenant, exists := mtps.tenants[tenantID]
	mtps.tenantsMu.RUnlock()

	if !exists {
		return ErrTenantNotFound
	}

	tenant.config.Quota = quota
	return nil
}

// GetUsage returns current usage for a tenant
func (mtps *MultiTenantPubSub) GetUsage(tenantID string) (*TenantUsage, error) {
	mtps.tenantsMu.RLock()
	tenant, exists := mtps.tenants[tenantID]
	mtps.tenantsMu.RUnlock()

	if !exists {
		return nil, ErrTenantNotFound
	}

	return &tenant.usage, nil
}

// Publish publishes a message with tenant quota enforcement
func (mtps *MultiTenantPubSub) Publish(tenantID, topic string, msg Message) error {
	if mtps.closed.Load() {
		return ErrPublishToClosed
	}

	// Get tenant
	tenant, err := mtps.getTenant(tenantID)
	if err != nil {
		return err
	}

	// Check if tenant is enabled
	if !tenant.config.Enabled {
		return fmt.Errorf("tenant %s is disabled", tenantID)
	}

	// Enforce quotas
	if err := mtps.enforcePublishQuota(tenant, msg); err != nil {
		tenant.usage.QuotaViolations.Add(1)
		return err
	}

	// Namespace the topic with tenant ID
	namespacedTopic := mtps.namespaceTopic(tenantID, topic)

	// Publish to underlying pubsub
	if err := mtps.ps.Publish(namespacedTopic, msg); err != nil {
		return err
	}

	// Update usage
	tenant.usage.PublishCount.Add(1)
	tenant.usage.PublishRate.Add(1)
	tenant.usage.LastPublishMu.Lock()
	tenant.usage.LastPublish = time.Now()
	tenant.usage.LastPublishMu.Unlock()

	return nil
}

// Subscribe creates a subscription with tenant quota enforcement
func (mtps *MultiTenantPubSub) Subscribe(tenantID, topic string, opts SubOptions) (Subscription, error) {
	// Get tenant
	tenant, err := mtps.getTenant(tenantID)
	if err != nil {
		return nil, err
	}

	// Check subscription quota
	if tenant.config.Quota.MaxSubscriptions > 0 {
		currentSubs := tenant.usage.ActiveSubs.Load()
		if currentSubs >= int64(tenant.config.Quota.MaxSubscriptions) {
			tenant.usage.QuotaViolations.Add(1)
			return nil, fmt.Errorf("%w: max subscriptions reached", ErrQuotaExceeded)
		}
	}

	// Namespace the topic
	namespacedTopic := mtps.namespaceTopic(tenantID, topic)

	// Create subscription
	sub, err := mtps.ps.Subscribe(namespacedTopic, opts)
	if err != nil {
		return nil, err
	}

	// Update usage
	tenant.usage.ActiveSubs.Add(1)

	// Wrap subscription to track lifecycle
	return &tenantSubscription{
		Subscription: sub,
		tenant:       tenant,
		tenantID:     tenantID,
		topic:        topic,
	}, nil
}

// enforcePublishQuota checks and enforces publish quotas
func (mtps *MultiTenantPubSub) enforcePublishQuota(tenant *tenantData, msg Message) error {
	quota := tenant.config.Quota

	// Check message size
	if quota.MaxMessageSize > 0 {
		// Estimate message size
		size := len(fmt.Sprintf("%v", msg.Data))
		if size > quota.MaxMessageSize {
			return fmt.Errorf("%w: message size %d exceeds limit %d",
				ErrQuotaExceeded, size, quota.MaxMessageSize)
		}
		tenant.usage.BytesPublished.Add(uint64(size))
		tenant.usage.BytesPerSecond.Add(uint64(size))
	}

	// Check publish rate (token bucket)
	if quota.MaxPublishRate > 0 {
		// Refill tokens
		tenant.lastRefillMu.Lock()
		now := time.Now()
		elapsed := now.Sub(tenant.lastRefill)
		if elapsed >= time.Second {
			tokensToAdd := int64(quota.MaxPublishRate)
			tenant.tokens.Store(tokensToAdd)
			tenant.lastRefill = now
		}
		tenant.lastRefillMu.Unlock()

		// Try to consume a token
		tokens := tenant.tokens.Load()
		if tokens <= 0 {
			return fmt.Errorf("%w: publish rate limit exceeded", ErrQuotaExceeded)
		}

		if !tenant.tokens.CompareAndSwap(tokens, tokens-1) {
			// Retry once
			tokens = tenant.tokens.Load()
			if tokens <= 0 {
				return fmt.Errorf("%w: publish rate limit exceeded", ErrQuotaExceeded)
			}
			tenant.tokens.Add(-1)
		}
	}

	// Check bandwidth (token bucket)
	if quota.MaxBandwidth > 0 {
		size := int64(len(fmt.Sprintf("%v", msg.Data)))

		tenant.lastBandwidthMu.Lock()
		now := time.Now()
		elapsed := now.Sub(tenant.lastBandwidthTime)
		if elapsed >= time.Second {
			tenant.bandwidthTokens.Store(quota.MaxBandwidth)
			tenant.lastBandwidthTime = now
		}
		tenant.lastBandwidthMu.Unlock()

		tokens := tenant.bandwidthTokens.Load()
		if tokens < size {
			return fmt.Errorf("%w: bandwidth limit exceeded", ErrQuotaExceeded)
		}

		tenant.bandwidthTokens.Add(-size)
	}

	return nil
}

// getTenant retrieves tenant data or creates it with default quota
func (mtps *MultiTenantPubSub) getTenant(tenantID string) (*tenantData, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenant
	}

	mtps.tenantsMu.RLock()
	tenant, exists := mtps.tenants[tenantID]
	mtps.tenantsMu.RUnlock()

	if !exists {
		// Auto-create tenant with default quota
		if err := mtps.CreateTenant(tenantID, mtps.defaultQuota); err != nil {
			return nil, err
		}

		mtps.tenantsMu.RLock()
		tenant = mtps.tenants[tenantID]
		mtps.tenantsMu.RUnlock()
	}

	return tenant, nil
}

// namespaceTopic prefixes topic with tenant ID
func (mtps *MultiTenantPubSub) namespaceTopic(tenantID, topic string) string {
	return fmt.Sprintf("_tenant_%s.%s", tenantID, topic)
}

// stripNamespace removes tenant prefix from topic
func (mtps *MultiTenantPubSub) stripNamespace(tenantID, namespacedTopic string) string {
	prefix := fmt.Sprintf("_tenant_%s.", tenantID)
	return strings.TrimPrefix(namespacedTopic, prefix)
}

// metricsCollector periodically resets rate counters
func (mtps *MultiTenantPubSub) metricsCollector() {
	defer mtps.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mtps.tenantsMu.RLock()
			for _, tenant := range mtps.tenants {
				// Reset rate counters
				tenant.usage.PublishRate.Store(0)
				tenant.usage.BytesPerSecond.Store(0)
			}
			mtps.tenantsMu.RUnlock()
		case <-mtps.ctx.Done():
			return
		}
	}
}

// ListTenants returns all tenant IDs
func (mtps *MultiTenantPubSub) ListTenants() []string {
	mtps.tenantsMu.RLock()
	defer mtps.tenantsMu.RUnlock()

	tenants := make([]string, 0, len(mtps.tenants))
	for id := range mtps.tenants {
		tenants = append(tenants, id)
	}

	return tenants
}

// Close stops the multi-tenant pubsub
func (mtps *MultiTenantPubSub) Close() error {
	if !mtps.closed.CompareAndSwap(false, true) {
		return nil
	}

	mtps.cancel()
	mtps.wg.Wait()

	return nil
}

// tenantSubscription wraps a subscription to track tenant usage
type tenantSubscription struct {
	Subscription
	tenant   *tenantData
	tenantID string
	topic    string
	canceled atomic.Bool
}

// Cancel cancels the subscription and updates usage
func (ts *tenantSubscription) Cancel() {
	if !ts.canceled.CompareAndSwap(false, true) {
		return
	}

	ts.Subscription.Cancel()
	ts.tenant.usage.ActiveSubs.Add(-1)
}
