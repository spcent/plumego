package provider

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages multiple providers and routes requests.
type Manager struct {
	providers map[string]Provider
	router    Router
	mu        sync.RWMutex
}

// NewManager creates a new provider manager.
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		providers: make(map[string]Provider),
		router:    &DefaultRouter{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// ManagerOption configures the manager.
type ManagerOption func(*Manager)

// WithRouter sets the routing strategy.
func WithRouter(router Router) ManagerOption {
	return func(m *Manager) {
		m.router = router
	}
}

// Register registers a provider.
func (m *Manager) Register(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// Get returns a provider by name.
func (m *Manager) Get(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return provider, nil
}

// Route routes a request to the appropriate provider.
func (m *Manager) Route(ctx context.Context, req *CompletionRequest) (Provider, error) {
	m.mu.RLock()
	providers := make([]Provider, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	m.mu.RUnlock()

	return m.router.Route(ctx, req, providers)
}

// Complete sends a completion request using routing.
func (m *Manager) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	provider, err := m.Route(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("route: %w", err)
	}

	return provider.Complete(ctx, req)
}

// CompleteStream sends a streaming completion request using routing.
func (m *Manager) CompleteStream(ctx context.Context, req *CompletionRequest) (*StreamReader, error) {
	provider, err := m.Route(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("route: %w", err)
	}

	return provider.CompleteStream(ctx, req)
}

// Router defines routing strategy.
type Router interface {
	Route(ctx context.Context, req *CompletionRequest, providers []Provider) (Provider, error)
}

// DefaultRouter implements default routing logic.
type DefaultRouter struct{}

// Route implements Router.
func (r *DefaultRouter) Route(ctx context.Context, req *CompletionRequest, providers []Provider) (Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Try to match by model prefix
	for _, p := range providers {
		if matchesProvider(req.Model, p.Name()) {
			return p, nil
		}
	}

	// Use first provider as fallback
	return providers[0], nil
}

// matchesProvider checks if a model belongs to a provider.
func matchesProvider(model, providerName string) bool {
	switch providerName {
	case "claude":
		return len(model) >= 6 && model[:6] == "claude"
	case "openai":
		return len(model) >= 3 && (model[:3] == "gpt" || model[:4] == "text")
	}
	return false
}

// LoadBalancerRouter implements load balancing across providers.
type LoadBalancerRouter struct {
	index uint64
	mu    sync.Mutex
}

// Route implements Router with round-robin load balancing.
func (r *LoadBalancerRouter) Route(ctx context.Context, req *CompletionRequest, providers []Provider) (Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	r.mu.Lock()
	idx := r.index % uint64(len(providers))
	r.index++
	r.mu.Unlock()

	return providers[idx], nil
}

// CostOptimizedRouter chooses the cheapest provider for a model.
type CostOptimizedRouter struct {
	// Model -> Provider mapping
	modelProviders map[string]string
}

// NewCostOptimizedRouter creates a cost-optimized router.
func NewCostOptimizedRouter() *CostOptimizedRouter {
	return &CostOptimizedRouter{
		modelProviders: map[string]string{
			// Map models to preferred providers based on cost
			"claude-3-haiku":  "claude",
			"claude-3-sonnet": "claude",
			"claude-3-opus":   "claude",
			"gpt-3.5-turbo":   "openai",
			"gpt-4":           "openai",
		},
	}
}

// Route implements Router.
func (r *CostOptimizedRouter) Route(ctx context.Context, req *CompletionRequest, providers []Provider) (Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Check if we have a preferred provider for this model
	if preferredProvider, ok := r.modelProviders[req.Model]; ok {
		for _, p := range providers {
			if p.Name() == preferredProvider {
				return p, nil
			}
		}
	}

	// Fallback to default routing
	return (&DefaultRouter{}).Route(ctx, req, providers)
}
