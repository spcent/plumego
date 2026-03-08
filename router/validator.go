package router

import (
	"strings"
)

// ParamValidator defines the interface for route parameter validation.
// Implementations live in the routeparam package; this interface is defined
// here to keep the router package free of concrete validation logic.
type ParamValidator interface {
	Validate(name, value string) error
}

// RouteValidation holds parameter validation rules for a single route.
type RouteValidation struct {
	Params map[string]ParamValidator
}

// NewRouteValidation creates an empty RouteValidation ready for use.
func NewRouteValidation() *RouteValidation {
	return &RouteValidation{
		Params: make(map[string]ParamValidator),
	}
}

// AddParam adds a validator for the named path parameter.
// Returns the receiver for chaining.
func (rv *RouteValidation) AddParam(name string, validator ParamValidator) *RouteValidation {
	rv.Params[name] = validator
	return rv
}

// Validate runs all registered validators against the provided params map.
func (rv *RouteValidation) Validate(params map[string]string) error {
	for name, value := range params {
		if validator, exists := rv.Params[name]; exists {
			if err := validator.Validate(name, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// WithValidation is a RouterOption that pre-registers validation rules.
// The map key format is "METHOD /path", e.g. "GET /users/:id".
func WithValidation(validations map[string]*RouteValidation) RouterOption {
	return func(r *Router) {
		r.state.routeValidations = make(map[string]map[string]*RouteValidation)
		for key, validation := range validations {
			parts := strings.SplitN(key, " ", 2)
			if len(parts) != 2 {
				continue
			}
			r.setValidation(parts[0], parts[1], validation)
		}
	}
}

// WithValidationRule is a RouterOption that adds a single validation rule.
func WithValidationRule(method, path string, validation *RouteValidation) RouterOption {
	return func(r *Router) {
		r.setValidation(method, path, validation)
	}
}

// AddValidation adds a validation rule for a specific route.
func (r *Router) AddValidation(method, path string, validation *RouteValidation) {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	normalizedPath := strings.TrimRight(path, "/")
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	pattern := r.prefix + normalizedPath
	r.setValidation(method, pattern, validation)

	if routeNode := r.findRouteNodeLocked(method, pattern); routeNode != nil {
		routeNode.validation = validation
	}
}
