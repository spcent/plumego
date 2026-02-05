package routing

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/spcent/plumego/tenant"
)

var (
	ErrInvalidRoutePolicy = errors.New("route policy invalid")
	ErrNoProvider         = errors.New("route policy has no provider")
	ErrUnknownStrategy    = errors.New("route policy strategy unsupported")
)

// PolicyPayload defines the minimal JSON schema for routing decisions.
// It is intentionally small and can be extended by your service.
type PolicyPayload struct {
	DefaultProvider string           `json:"default_provider"`
	Providers       []ProviderWeight `json:"providers"`
}

// ProviderWeight describes a weighted provider option.
type ProviderWeight struct {
	Name   string `json:"provider"`
	Weight int    `json:"weight"`
}

// RandomSource allows deterministic tests.
type RandomSource interface {
	Intn(n int) int
}

// PolicyRouter selects a provider based on tenant route policy.
type PolicyRouter struct {
	Provider tenant.RoutePolicyProvider
	Random   RandomSource
}

// SelectProvider returns the selected provider and the policy used.
func (r *PolicyRouter) SelectProvider(ctx context.Context, tenantID string) (string, tenant.RoutePolicy, error) {
	if r == nil || r.Provider == nil {
		return "", tenant.RoutePolicy{}, ErrInvalidRoutePolicy
	}

	policy, err := r.Provider.RoutePolicy(ctx, tenantID)
	if err != nil {
		return "", tenant.RoutePolicy{}, err
	}

	payload, err := parsePayload(policy.Payload)
	if err != nil {
		return "", policy, err
	}

	strategy := policy.Strategy
	if strategy == "" {
		strategy = "weighted"
	}

	switch strategy {
	case "direct":
		if payload.DefaultProvider != "" {
			return payload.DefaultProvider, policy, nil
		}
		if len(payload.Providers) > 0 && payload.Providers[0].Name != "" {
			return payload.Providers[0].Name, policy, nil
		}
		return "", policy, ErrNoProvider
	case "weighted":
		provider, err := selectWeighted(payload, r.random())
		if err != nil {
			return "", policy, err
		}
		return provider, policy, nil
	default:
		return "", policy, ErrUnknownStrategy
	}
}

func (r *PolicyRouter) random() RandomSource {
	if r.Random != nil {
		return r.Random
	}
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func parsePayload(data []byte) (PolicyPayload, error) {
	if len(data) == 0 {
		return PolicyPayload{}, ErrInvalidRoutePolicy
	}
	var payload PolicyPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return PolicyPayload{}, ErrInvalidRoutePolicy
	}
	return payload, nil
}

func selectWeighted(payload PolicyPayload, rng RandomSource) (string, error) {
	if len(payload.Providers) == 0 {
		if payload.DefaultProvider != "" {
			return payload.DefaultProvider, nil
		}
		return "", ErrNoProvider
	}

	total := 0
	for _, provider := range payload.Providers {
		if provider.Weight > 0 && provider.Name != "" {
			total += provider.Weight
		}
	}
	if total <= 0 {
		return "", ErrNoProvider
	}

	roll := rng.Intn(total)
	running := 0
	for _, provider := range payload.Providers {
		if provider.Weight <= 0 || provider.Name == "" {
			continue
		}
		running += provider.Weight
		if roll < running {
			return provider.Name, nil
		}
	}

	return "", ErrNoProvider
}
