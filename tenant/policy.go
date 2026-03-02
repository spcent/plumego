package tenant

import (
	"context"
	"errors"
	"strings"
)

var ErrPolicyDenied = errors.New("policy denied")

// PolicyConfig defines allow lists for tenant access control.
// Empty lists mean "allow all" for that dimension.
type PolicyConfig struct {
	// AllowedModels restricts which AI model identifiers the tenant may use.
	AllowedModels []string
	// AllowedTools restricts which tool identifiers the tenant may use.
	AllowedTools []string
	// AllowedMethods restricts which HTTP methods the tenant may use.
	// Supports exact match (e.g. "GET") only.
	AllowedMethods []string
	// AllowedPaths restricts which URL paths the tenant may access.
	// Supports exact match or prefix match with a trailing "*"
	// (e.g. "/api/v1/*" matches any path starting with "/api/v1/").
	AllowedPaths []string
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

// Evaluate checks allow lists for models, tools, HTTP methods, and URL paths.
// Empty allow lists permit all values for that dimension.
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
	if !allowedInList(cfg.AllowedMethods, req.Method) {
		return PolicyResult{Allowed: false, Reason: "method not allowed"}, ErrPolicyDenied
	}
	if !pathAllowed(cfg.AllowedPaths, req.Path) {
		return PolicyResult{Allowed: false, Reason: "path not allowed"}, ErrPolicyDenied
	}

	return PolicyResult{Allowed: true}, nil
}

// allowedInList returns true when list is empty (allow all) or value is in the list.
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

// pathAllowed returns true when allowedPaths is empty (allow all), path is empty,
// or path matches an entry by exact match or prefix match (trailing "*").
func pathAllowed(allowedPaths []string, path string) bool {
	if len(allowedPaths) == 0 || path == "" {
		return true
	}
	for _, pattern := range allowedPaths {
		if strings.HasSuffix(pattern, "*") {
			prefix := pattern[:len(pattern)-1]
			if strings.HasPrefix(path, prefix) {
				return true
			}
		} else if pattern == path {
			return true
		}
	}
	return false
}
