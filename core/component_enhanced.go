package core

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
)

// EnhancedComponentManager provides advanced component discovery and composition capabilities
type EnhancedComponentManager struct {
	mu           sync.RWMutex
	components   map[string]*EnhancedComponentInfo
	dependencies map[string][]string // component -> dependencies
}

// EnhancedComponentInfo contains rich metadata about a component
type EnhancedComponentInfo struct {
	Name         string
	Type         reflect.Type
	Category     string
	Description  string
	Dependencies []string
	Lifecycle    ComponentLifecycle
	Tags         []string
	Metadata     map[string]interface{}
	instance     interface{} // cached instance for singleton
}

// ComponentLifecycle represents the lifecycle strategy for a component
type ComponentLifecycle string

const (
	LifecycleSingleton ComponentLifecycle = "singleton"
	LifecycleTransient ComponentLifecycle = "transient"
)

// NewEnhancedComponentManager creates a new enhanced component manager
func NewEnhancedComponentManager() *EnhancedComponentManager {
	return &EnhancedComponentManager{
		components:   make(map[string]*EnhancedComponentInfo),
		dependencies: make(map[string][]string),
	}
}

// RegisterComponent registers a component with rich metadata
func (ecm *EnhancedComponentManager) RegisterComponent(
	name string,
	component interface{},
	category string,
	description string,
	dependencies []string,
	lifecycle ComponentLifecycle,
	tags []string,
	metadata map[string]interface{},
) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	componentType := reflect.TypeOf(component)
	ecm.components[name] = &EnhancedComponentInfo{
		Name:         name,
		Type:         componentType,
		Category:     category,
		Description:  description,
		Dependencies: dependencies,
		Lifecycle:    lifecycle,
		Tags:         tags,
		Metadata:     metadata,
		instance:     nil,
	}

	ecm.dependencies[name] = dependencies
	return nil
}

// RegisterComponentFactory registers a component factory with metadata
func (ecm *EnhancedComponentManager) RegisterComponentFactory(
	name string,
	factory func() (interface{}, error),
	category string,
	description string,
	dependencies []string,
	lifecycle ComponentLifecycle,
	tags []string,
) error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	ecm.components[name] = &EnhancedComponentInfo{
		Name:         name,
		Type:         nil, // Will be resolved on first instantiation
		Category:     category,
		Description:  description,
		Dependencies: dependencies,
		Lifecycle:    lifecycle,
		Tags:         tags,
		Metadata:     map[string]interface{}{"factory": true, "factory_func": factory},
		instance:     nil,
	}

	ecm.dependencies[name] = dependencies
	return nil
}

// ListComponents returns all registered components with optional filtering
func (ecm *EnhancedComponentManager) ListComponents(filter ...func(*EnhancedComponentInfo) bool) []*EnhancedComponentInfo {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	var result []*EnhancedComponentInfo
	for _, info := range ecm.components {
		match := true
		for _, f := range filter {
			if !f(info) {
				match = false
				break
			}
		}
		if match {
			// Return a copy to avoid external modification
			infoCopy := *info
			result = append(result, &infoCopy)
		}
	}

	// Sort by name for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// GetComponentInfo returns detailed information about a specific component
func (ecm *EnhancedComponentManager) GetComponentInfo(name string) (*EnhancedComponentInfo, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	info, exists := ecm.components[name]
	if !exists {
		return nil, fmt.Errorf("component %s not found", name)
	}

	// Return a copy
	infoCopy := *info
	return &infoCopy, nil
}

// GetDependencyGraph returns the dependency graph of all components
func (ecm *EnhancedComponentManager) GetDependencyGraph() *DependencyGraph {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Edges: make([]DependencyEdge, 0),
	}

	// Add all nodes
	for name, info := range ecm.components {
		graph.Nodes[name] = &DependencyNode{
			Name:     name,
			Category: info.Category,
			Status:   "registered",
		}
	}

	// Add all edges
	for name, deps := range ecm.dependencies {
		for _, dep := range deps {
			graph.Edges = append(graph.Edges, DependencyEdge{
				From: dep,
				To:   name,
			})
		}
	}

	return graph
}

// ValidateDependencies checks for missing or circular dependencies
func (ecm *EnhancedComponentManager) ValidateDependencies() []error {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	var errors []error

	// Check for missing dependencies
	for name, deps := range ecm.dependencies {
		for _, dep := range deps {
			if _, exists := ecm.components[dep]; !exists {
				errors = append(errors, fmt.Errorf("component %s missing dependency: %s", name, dep))
			}
		}
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var detectCycle func(string, []string) bool
	detectCycle = func(node string, path []string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range ecm.dependencies[node] {
			if !visited[dep] {
				if detectCycle(dep, append(path, dep)) {
					return true
				}
			} else if recStack[dep] {
				// Found cycle
				cyclePath := append(path, dep)
				errors = append(errors, fmt.Errorf("circular dependency detected: %s", strings.Join(cyclePath, " -> ")))
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for name := range ecm.components {
		if !visited[name] {
			detectCycle(name, []string{name})
		}
	}

	return errors
}

// GetComponentsByCategory returns all components in a specific category
func (ecm *EnhancedComponentManager) GetComponentsByCategory(category string) []*EnhancedComponentInfo {
	return ecm.ListComponents(func(info *EnhancedComponentInfo) bool {
		return info.Category == category
	})
}

// GetComponentsByTag returns all components with a specific tag
func (ecm *EnhancedComponentManager) GetComponentsByTag(tag string) []*EnhancedComponentInfo {
	return ecm.ListComponents(func(info *EnhancedComponentInfo) bool {
		for _, t := range info.Tags {
			if t == tag {
				return true
			}
		}
		return false
	})
}

// GetDependencyChain returns the full dependency chain for a component
func (ecm *EnhancedComponentManager) GetDependencyChain(name string) ([]string, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	visited := make(map[string]bool)
	var chain []string

	var buildChain func(string) error
	buildChain = func(current string) error {
		if visited[current] {
			return nil // Already in chain
		}
		visited[current] = true

		// Add dependencies first
		if deps, exists := ecm.dependencies[current]; exists {
			for _, dep := range deps {
				if err := buildChain(dep); err != nil {
					return err
				}
			}
		}

		// Then add current
		chain = append(chain, current)
		return nil
	}

	if err := buildChain(name); err != nil {
		return nil, err
	}

	return chain, nil
}

// ResolveAndInstantiate resolves dependencies and instantiates components
func (ecm *EnhancedComponentManager) ResolveAndInstantiate() error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	// Validate dependencies first
	errors := ecm.validateDependenciesUnlocked()
	if len(errors) > 0 {
		return fmt.Errorf("dependency validation failed: %v", errors)
	}

	// Build instantiation order (topological sort)
	instantiationOrder, err := ecm.getInstantiationOrderUnlocked()
	if err != nil {
		return err
	}

	// Instantiate components
	for _, name := range instantiationOrder {
		info := ecm.components[name]

		// Skip if already instantiated
		if info.instance != nil {
			continue
		}

		// Check if it's a factory
		if factory, ok := info.Metadata["factory_func"]; ok {
			factoryFunc := factory.(func() (interface{}, error))
			instance, err := factoryFunc()
			if err != nil {
				return fmt.Errorf("failed to instantiate component %s: %w", name, err)
			}
			info.instance = instance
			info.Type = reflect.TypeOf(instance)
		} else {
			// Direct instantiation for non-factory components
			if info.Type != nil && info.Type.Kind() == reflect.Ptr {
				// Create a new instance
				instance := reflect.New(info.Type.Elem()).Interface()
				info.instance = instance
			}
		}
	}

	return nil
}

// GetInstance returns the instantiated component
func (ecm *EnhancedComponentManager) GetInstance(name string) (interface{}, error) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	info, exists := ecm.components[name]
	if !exists {
		return nil, fmt.Errorf("component %s not found", name)
	}

	if info.instance == nil {
		return nil, fmt.Errorf("component %s not yet instantiated", name)
	}

	return info.instance, nil
}

// GenerateComponentReport generates a comprehensive report of all components
func (ecm *EnhancedComponentManager) GenerateComponentReport() string {
	var builder strings.Builder

	builder.WriteString("# Component Registry Report\n\n")

	// Summary
	components := ecm.ListComponents()
	builder.WriteString(fmt.Sprintf("## Summary\n\n"))
	builder.WriteString(fmt.Sprintf("- **Total Components**: %d\n", len(components)))

	categories := make(map[string]int)
	tags := make(map[string]int)
	for _, comp := range components {
		categories[comp.Category]++
		for _, tag := range comp.Tags {
			tags[tag]++
		}
	}

	builder.WriteString(fmt.Sprintf("- **Categories**: %d\n", len(categories)))
	builder.WriteString(fmt.Sprintf("- **Unique Tags**: %d\n", len(tags)))
	builder.WriteString("\n")

	// Categories
	if len(categories) > 0 {
		builder.WriteString("## Components by Category\n\n")
		for _, category := range sortedKeys(categories) {
			builder.WriteString(fmt.Sprintf("### %s (%d)\n\n", category, categories[category]))

			for _, comp := range ecm.GetComponentsByCategory(category) {
				builder.WriteString(fmt.Sprintf("#### `%s`\n", comp.Name))
				builder.WriteString(fmt.Sprintf("- **Description**: %s\n", comp.Description))
				builder.WriteString(fmt.Sprintf("- **Lifecycle**: %s\n", comp.Lifecycle))
				if len(comp.Dependencies) > 0 {
					builder.WriteString(fmt.Sprintf("- **Dependencies**: %s\n", strings.Join(comp.Dependencies, ", ")))
				}
				if len(comp.Tags) > 0 {
					builder.WriteString(fmt.Sprintf("- **Tags**: %s\n", strings.Join(comp.Tags, ", ")))
				}
				builder.WriteString("\n")
			}
		}
	}

	// Dependency Graph
	builder.WriteString("## Dependency Graph\n\n")
	graph := ecm.GetDependencyGraph()
	if len(graph.Edges) > 0 {
		builder.WriteString("```mermaid\n")
		builder.WriteString("graph TD\n")
		for _, edge := range graph.Edges {
			builder.WriteString(fmt.Sprintf("    %s --> %s\n", edge.From, edge.To))
		}
		builder.WriteString("```\n\n")
	} else {
		builder.WriteString("No explicit dependencies registered.\n\n")
	}

	// Validation Results
	builder.WriteString("## Validation Results\n\n")
	errors := ecm.ValidateDependencies()
	if len(errors) == 0 {
		builder.WriteString("✓ All dependencies are valid\n")
	} else {
		builder.WriteString(fmt.Sprintf("✗ Found %d validation errors:\n\n", len(errors)))
		for i, err := range errors {
			builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, err.Error()))
		}
	}

	return builder.String()
}

// Helper methods

func (ecm *EnhancedComponentManager) validateDependenciesUnlocked() []error {
	var errors []error

	// Check for missing dependencies
	for name, deps := range ecm.dependencies {
		for _, dep := range deps {
			if _, exists := ecm.components[dep]; !exists {
				errors = append(errors, fmt.Errorf("component %s missing dependency: %s", name, dep))
			}
		}
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var detectCycle func(string, []string) bool
	detectCycle = func(node string, path []string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range ecm.dependencies[node] {
			if !visited[dep] {
				if detectCycle(dep, append(path, dep)) {
					return true
				}
			} else if recStack[dep] {
				cyclePath := append(path, dep)
				errors = append(errors, fmt.Errorf("circular dependency detected: %s", strings.Join(cyclePath, " -> ")))
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for name := range ecm.components {
		if !visited[name] {
			detectCycle(name, []string{name})
		}
	}

	return errors
}

func (ecm *EnhancedComponentManager) getInstantiationOrderUnlocked() ([]string, error) {
	// Topological sort using Kahn's algorithm
	inDegree := make(map[string]int)
	graph := make(map[string][]string)

	// Build graph and calculate in-degrees
	for name, deps := range ecm.dependencies {
		for _, dep := range deps {
			graph[dep] = append(graph[dep], name)
			inDegree[name]++
		}
	}

	// Find all nodes with in-degree 0
	var queue []string
	for name := range ecm.components {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}

	var order []string
	for len(queue) > 0 {
		// Pop from queue
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		// Reduce in-degree of neighbors
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(order) != len(ecm.components) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return order, nil
}

// DependencyGraph represents the dependency relationships between components
type DependencyGraph struct {
	Nodes map[string]*DependencyNode
	Edges []DependencyEdge
}

// DependencyNode represents a single component in the dependency graph
type DependencyNode struct {
	Name     string
	Category string
	Status   string // registered, resolved, error
}

// DependencyEdge represents a dependency relationship
type DependencyEdge struct {
	From string
	To   string
}

// Helper function to sort map keys
func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Integration with Plumego's Component interface

// ComponentAdapter adapts enhanced components to Plumego's Component interface
type ComponentAdapter struct {
	*BaseComponent
	name     string
	manager  *EnhancedComponentManager
	router   *router.Router
	registry *middleware.Registry
}

// NewComponentAdapter creates a new component adapter
func NewComponentAdapter(name string, manager *EnhancedComponentManager) *ComponentAdapter {
	return &ComponentAdapter{
		BaseComponent: &BaseComponent{},
		name:          name,
		manager:       manager,
	}
}

// RegisterRoutes implements Component.RegisterRoutes
func (ca *ComponentAdapter) RegisterRoutes(r *router.Router) {
	ca.router = r
	// Components can register their routes here
}

// RegisterMiddleware implements Component.RegisterMiddleware
func (ca *ComponentAdapter) RegisterMiddleware(m *middleware.Registry) {
	ca.registry = m
	// Components can register their middleware here
}

// Start implements Component.Start
func (ca *ComponentAdapter) Start(ctx context.Context) error {
	// Resolve and instantiate all components
	return ca.manager.ResolveAndInstantiate()
}

// Stop implements Component.Stop
func (ca *ComponentAdapter) Stop(ctx context.Context) error {
	// Cleanup logic can be added here
	return nil
}

// Health implements Component.Health
func (ca *ComponentAdapter) Health() (name string, status health.HealthStatus) {
	// Check if component is healthy
	errors := ca.manager.ValidateDependencies()
	if len(errors) > 0 {
		return ca.name, health.HealthStatus{
			Status:  health.StatusUnhealthy,
			Message: fmt.Sprintf("validation errors: %v", errors),
			Details: map[string]any{
				"error_count": len(errors),
				"errors":      errors,
			},
		}
	}
	return ca.name, health.HealthStatus{
		Status:  health.StatusHealthy,
		Message: "component is healthy",
		Details: map[string]any{
			"component_count": len(ca.manager.ListComponents()),
		},
	}
}
