package tenant

import (
	"context"
	"errors"
)

var ErrPolicyDenied = errors.New("policy denied")

// PolicyConfig defines allow lists for tenant access control.
type PolicyConfig struct {
	AllowedModels []string
	AllowedTools  []string
}

// PolicyRequest describes policy evaluation inputs.
type PolicyRequest struct {
	Model  string
	Tool   string
	Method string
	Path   string
}

// PolicyResult describes policy evaluation output.
type PolicyResult struct {
	Allowed bool
	Reason  string
}

// PolicyEvaluator evaluates tenant policy requests.
type PolicyEvaluator interface {
	Evaluate(ctx context.Context, tenantID string, req PolicyRequest) (PolicyResult, error)
}

// ConfigPolicyEvaluator checks policy against a tenant configuration provider.
type ConfigPolicyEvaluator struct {
	Provider PolicyConfigProvider
}

// NewConfigPolicyEvaluator creates a policy evaluator from a config provider.
func NewConfigPolicyEvaluator(provider PolicyConfigProvider) *ConfigPolicyEvaluator {
	return &ConfigPolicyEvaluator{Provider: provider}
}

// Evaluate checks allow lists for models and tools. Empty lists allow all.
func (e *ConfigPolicyEvaluator) Evaluate(ctx context.Context, tenantID string, req PolicyRequest) (PolicyResult, error) {
	if e == nil || e.Provider == nil {
		return PolicyResult{Allowed: true}, nil
	}

	cfg, err := e.Provider.PolicyConfig(ctx, tenantID)
	if err != nil {
		return PolicyResult{Allowed: false, Reason: err.Error()}, err
	}

	if !allowedInList(cfg.AllowedModels, req.Model) {
		return PolicyResult{Allowed: false, Reason: "model not allowed"}, ErrPolicyDenied
	}
	if !allowedInList(cfg.AllowedTools, req.Tool) {
		return PolicyResult{Allowed: false, Reason: "tool not allowed"}, ErrPolicyDenied
	}

	return PolicyResult{Allowed: true}, nil
}

func allowedInList(list []string, value string) bool {
	if len(list) == 0 || value == "" {
		return true
	}
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
