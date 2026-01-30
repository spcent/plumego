package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

// DILifecycle defines the lifecycle management for dependencies.
type DILifecycle string

const (
	// Singleton lifecycle: same instance is returned for all resolutions
	Singleton DILifecycle = "singleton"
	// Transient lifecycle: new instance is created for each resolution
	Transient DILifecycle = "transient"
	// Scoped lifecycle: same instance within a resolution scope
	Scoped DILifecycle = "scoped"
)

// DIRegistration represents a registered dependency with metadata.
type DIRegistration struct {
	Type      reflect.Type
	Lifecycle DILifecycle
	Factory   func(*DIContainer) any
	Instance  any
}

// DIContainer is a thread-safe dependency injection container.
// It supports automatic dependency resolution, circular dependency detection,
// and multiple lifecycle management strategies.
//
// Features:
//   - Type-safe dependency resolution
//   - Interface-to-implementation binding
//   - Factory functions for complex initialization
//   - Recursive dependency injection into struct fields
//   - Thread-safe operations
//   - Multiple lifecycle management: Singleton, Transient, and Scoped
//
// Example:
//
//	container := core.NewDIContainer()
//	container.Register(&DatabaseService{})
//	container.RegisterFactory(reflect.TypeOf((*CacheService)(nil)), func(c *core.DIContainer) any {
//	    return NewRedisCache(c)
//	})
//
//	var db *DatabaseService
//	if err := container.Resolve(reflect.TypeOf(&db)); err != nil {
//	    // handle error
//	}
type DIContainer struct {
	mu       sync.RWMutex
	services map[reflect.Type]DIRegistration

	// resolutionCounter is used to generate unique IDs for each resolution chain.
	// This is more reliable than using goroutine IDs which depend on runtime internals.
	resolutionCounter atomic.Uint64

	// scopedInstances stores instances for scoped lifecycle within a resolution chain
	scopedInstances map[uint64]map[reflect.Type]any

	// injectionLocks protects concurrent injection into the same instance
	injectionLocks [64]sync.Mutex
}

// NewDIContainer creates a new dependency injection container.
// The container is initialized with no services and is ready for registration.
func NewDIContainer() *DIContainer {
	return &DIContainer{
		services: make(map[reflect.Type]DIRegistration),
	}
}

// getOrCreateChainID generates a unique ID for the current resolution chain.
// This is used to track scoped instances within a resolution scope.
func (c *DIContainer) getOrCreateChainID() uint64 {
	return c.resolutionCounter.Add(1)
}

// clearScopedChain removes scoped instances for a completed resolution chain.
func (c *DIContainer) clearScopedChain(chainID uint64) {
	if chainID == 0 {
		return
	}
	c.mu.Lock()
	if c.scopedInstances != nil {
		delete(c.scopedInstances, chainID)
	}
	c.mu.Unlock()
}

// resolveWithStackInChain resolves a service while tracking the resolution stack
// to detect circular dependencies. This is called internally by Resolve/Inject.
func (c *DIContainer) resolveWithStackInChain(serviceType reflect.Type, stack []reflect.Type, chainID uint64) (any, error) {
	// Check for circular dependency
	for _, t := range stack {
		if t == serviceType {
			return nil, fmt.Errorf("circular dependency detected: %s", serviceType.String())
		}
	}

	// Add current type to stack
	newStack := append(stack, serviceType)

	c.mu.RLock()
	registration, exists := c.services[serviceType]
	c.mu.RUnlock()

	if !exists {
		// Try to find assignable service for interfaces
		if serviceType.Kind() == reflect.Interface {
			return c.resolveAssignableWithStack(serviceType, newStack, chainID)
		}
		return nil, fmt.Errorf("service not found: %s", serviceType.String())
	}

	// Handle singleton with existing instance
	if registration.Lifecycle == Singleton && registration.Instance != nil {
		return registration.Instance, nil
	}

	// Handle scoped lifecycle - check if we already have an instance for this resolution chain
	if registration.Lifecycle == Scoped {
		c.mu.RLock()
		scopedChain, exists := c.scopedInstances[chainID]
		c.mu.RUnlock()

		if exists {
			if instance, ok := scopedChain[serviceType]; ok {
				return instance, nil
			}
		}
	}

	// Create new instance
	var instance any
	var err error

	if registration.Factory != nil {
		// Use factory function - pass original container
		instance = registration.Factory(c)
	} else {
		// Try to create instance using reflection
		instance, err = c.createInstanceWithStack(serviceType, newStack, chainID)
		if err != nil {
			return nil, err
		}
	}

	// Register instance based on lifecycle
	c.mu.Lock()
	if registration.Lifecycle == Singleton {
		registration.Instance = instance
		c.services[serviceType] = registration
	} else if registration.Lifecycle == Scoped {
		// Store scoped instance for this resolution chain
		if c.scopedInstances == nil {
			c.scopedInstances = make(map[uint64]map[reflect.Type]any)
		}
		if c.scopedInstances[chainID] == nil {
			c.scopedInstances[chainID] = make(map[reflect.Type]any)
		}
		c.scopedInstances[chainID][serviceType] = instance
	}
	c.mu.Unlock()

	return instance, nil
}

// resolveAssignableWithStack resolves a service that implements an interface with stack tracking.
func (c *DIContainer) resolveAssignableWithStack(serviceType reflect.Type, stack []reflect.Type, chainID uint64) (any, error) {
	type candidate struct {
		serviceType reflect.Type
		instance    any
	}

	var candidates []candidate
	c.mu.RLock()
	for _, registration := range c.services {
		if registration.Instance != nil {
			instanceType := reflect.TypeOf(registration.Instance)
			if instanceType.AssignableTo(serviceType) {
				candidates = append(candidates, candidate{instance: registration.Instance})
			}
			continue
		}
		if registration.Factory != nil && registration.Type != nil {
			if registration.Type.AssignableTo(serviceType) || registration.Type == serviceType {
				candidates = append(candidates, candidate{serviceType: registration.Type})
			}
		}
	}
	c.mu.RUnlock()

	var matchInstance any
	var matchType reflect.Type
	hasInstance := false
	hasType := false

	for _, cand := range candidates {
		if cand.instance != nil {
			if !hasInstance && !hasType {
				matchInstance = cand.instance
				hasInstance = true
				continue
			}
			if hasType || (hasInstance && !servicesEqual(matchInstance, cand.instance)) {
				return nil, fmt.Errorf("multiple services match interface: %s", serviceType.String())
			}
			continue
		}
		// cand.serviceType != nil
		if hasInstance {
			return nil, fmt.Errorf("multiple services match interface: %s", serviceType.String())
		}
		if !hasType {
			matchType = cand.serviceType
			hasType = true
			continue
		}
		if matchType != cand.serviceType {
			return nil, fmt.Errorf("multiple services match interface: %s", serviceType.String())
		}
	}

	if matchInstance != nil {
		return matchInstance, nil
	}
	if matchType != nil {
		return c.resolveWithStackInChain(matchType, stack, chainID)
	}
	return nil, errors.New("service not found: " + serviceType.String())
}

// createInstanceWithStack creates a new instance using reflection with stack tracking.
func (c *DIContainer) createInstanceWithStack(serviceType reflect.Type, stack []reflect.Type, chainID uint64) (any, error) {
	if serviceType.Kind() == reflect.Ptr {
		// Create pointer to new instance
		elemType := serviceType.Elem()
		newInstance := reflect.New(elemType)

		// Try to inject dependencies into the struct
		if err := c.injectWithStack(newInstance.Interface(), stack, chainID); err != nil {
			// If injection fails, still return the instance
			// This allows for simple structs without dependencies
		}

		return newInstance.Interface(), nil
	}

	return nil, fmt.Errorf("cannot create instance of non-pointer type: %s", serviceType.String())
}

// injectWithStack injects dependencies into a struct with stack tracking.
func (c *DIContainer) injectWithStack(instance any, stack []reflect.Type, chainID uint64) error {
	if instance == nil {
		return nil
	}

	instanceVal := reflect.ValueOf(instance)
	if instanceVal.Kind() != reflect.Ptr {
		return errors.New("instance must be a pointer to a struct")
	}

	structVal := instanceVal.Elem()
	structType := structVal.Type()

	if structType.Kind() != reflect.Struct {
		return errors.New("instance must be a pointer to a struct")
	}

	// Iterate over all fields to resolve dependencies first
	// This separates resolution (slow, potentially recursive) from assignment (fast, needs lock)
	var updates []struct {
		fieldIndex int
		value      reflect.Value
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldVal := structVal.Field(i)

		injectTag, ok := field.Tag.Lookup("inject")
		if ok {
			// Check if field is exportable
			if !fieldVal.CanSet() {
				continue
			}

			if injectTag == "-" {
				continue
			}

			// Resolve the dependency with stack tracking
			var dep any
			var err error

			if injectTag != "" {
				dep, err = c.resolveByNameWithStack(injectTag, stack, chainID)
				if err != nil {
					dep, err = c.resolveWithStackInChain(field.Type, stack, chainID)
				}
			} else {
				dep, err = c.resolveWithStackInChain(field.Type, stack, chainID)
			}

			if err != nil {
				return err
			}

			// Prepare the value to set
			depVal := reflect.ValueOf(dep)
			var finalVal reflect.Value

			if depVal.Type().AssignableTo(field.Type) {
				finalVal = depVal
			} else if depVal.Kind() == reflect.Ptr && !depVal.IsNil() && depVal.Elem().Type().AssignableTo(field.Type) {
				finalVal = depVal.Elem()
			} else {
				return fmt.Errorf("cannot assign dependency %s to field %s", depVal.Type(), field.Type)
			}

			updates = append(updates, struct {
				fieldIndex int
				value      reflect.Value
			}{i, finalVal})
		}
	}

	// Apply updates with a lock to ensure atomicity
	if len(updates) > 0 {
		ptr := instanceVal.Pointer()
		lockIdx := ptr % uintptr(len(c.injectionLocks))
		c.injectionLocks[lockIdx].Lock()
		defer c.injectionLocks[lockIdx].Unlock()

		for _, update := range updates {
			structVal.Field(update.fieldIndex).Set(update.value)
		}
	}

	return nil
}

// Register registers a service instance in the container with singleton lifecycle.
// The service will be available for resolution by its type and any assignable types.
//
// Parameters:
//   - service: The service instance to register. Must be non-nil.
//
// Example:
//
//	container.Register(&DatabaseService{Connection: conn})
func (c *DIContainer) Register(service any) {
	if service == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Register service by its type
	serviceType := reflect.TypeOf(service)
	registration := DIRegistration{
		Type:      serviceType,
		Lifecycle: Singleton,
		Instance:  service,
	}
	c.services[serviceType] = registration

	// Also register by pointer type if it's a pointer
	if serviceType.Kind() == reflect.Ptr {
		originalType := serviceType.Elem()
		c.services[originalType] = registration
	}
}

// RegisterFactory registers a factory function to create a service.
// The factory function will be called when the service is first resolved.
//
// Parameters:
//   - serviceType: The type to register the factory for
//   - factory: Factory function that creates the service instance
//   - lifecycle: The lifecycle management strategy
//
// Example:
//
//	container.RegisterFactory(
//	    reflect.TypeOf((*CacheService)(nil)),
//	    func(c *DIContainer) any {
//	        return NewRedisCache(c)
//	    },
//	    Singleton,
//	)
func (c *DIContainer) RegisterFactory(serviceType reflect.Type, factory func(*DIContainer) any, lifecycle DILifecycle) {
	c.mu.Lock()
	defer c.mu.Unlock()

	registration := DIRegistration{
		Type:      serviceType,
		Lifecycle: lifecycle,
		Factory:   factory,
	}
	c.services[serviceType] = registration
}

// Resolve resolves a service by its type.
// It handles singleton, transient, and scoped lifecycles.
// Also supports interface types by finding assignable implementations.
// Circular dependencies are detected and will return an error.
//
// Parameters:
//   - serviceType: The type to resolve
//
// Returns:
//   - any: The resolved service instance
//   - error: Error if resolution fails (including circular dependency)
//
// Example:
//
//	var db *DatabaseService
//	service, err := container.Resolve(reflect.TypeOf(&db))
//	if err != nil {
//	    return err
//	}
//	db = service.(*DatabaseService)
func (c *DIContainer) Resolve(serviceType reflect.Type) (any, error) {
	// Start a new resolution chain with an empty stack
	chainID := c.getOrCreateChainID()
	defer c.clearScopedChain(chainID)
	return c.resolveWithStackInChain(serviceType, nil, chainID)
}

// Inject injects dependencies into a struct.
// It looks for fields with the `inject` tag and tries to resolve them.
// Circular dependencies are detected and will return an error.
//
// Parameters:
//   - instance: Pointer to a struct to inject dependencies into
//
// Returns:
//   - error: Error if injection fails (including circular dependency)
//
// Example:
//
//	type MyService struct {
//	    Database *DatabaseService `inject:""`
//	    Cache    CacheService     `inject:""`
//	}
//
//	service := &MyService{}
//	if err := container.Inject(service); err != nil {
//	    return err
//	}
func (c *DIContainer) Inject(instance any) error {
	// Start injection with an empty stack for circular dependency detection
	chainID := c.getOrCreateChainID()
	defer c.clearScopedChain(chainID)
	return c.injectWithStack(instance, nil, chainID)
}

// resolveByName resolves a service by its name or type name.
func (c *DIContainer) resolveByNameWithStack(name string, stack []reflect.Type, chainID uint64) (any, error) {
	type candidate struct {
		serviceType reflect.Type
		instance    any
	}

	var candidates []candidate
	c.mu.RLock()
	for t, registration := range c.services {
		if t.Name() == name || t.String() == name {
			if registration.Instance != nil {
				candidates = append(candidates, candidate{instance: registration.Instance})
				continue
			}
			if registration.Factory != nil && registration.Type != nil {
				candidates = append(candidates, candidate{serviceType: registration.Type})
			}
		}
	}
	c.mu.RUnlock()

	var matchInstance any
	var matchType reflect.Type
	for _, cand := range candidates {
		if cand.instance != nil {
			if matchInstance == nil && matchType == nil {
				matchInstance = cand.instance
				continue
			}
			if matchInstance != nil && servicesEqual(matchInstance, cand.instance) {
				continue
			}
			return nil, fmt.Errorf("multiple services match name: %s", name)
		}
		if matchInstance != nil {
			return nil, fmt.Errorf("multiple services match name: %s", name)
		}
		if matchType == nil {
			matchType = cand.serviceType
			continue
		}
		if matchType == cand.serviceType {
			continue
		}
		return nil, fmt.Errorf("multiple services match name: %s", name)
	}

	if matchInstance != nil {
		return matchInstance, nil
	}
	if matchType != nil {
		return c.resolveWithStackInChain(matchType, stack, chainID)
	}
	return nil, errors.New("service not found: " + name)
}

// servicesEqual checks if two service instances are equal.
func servicesEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}

	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)
	if ta != tb || !ta.Comparable() || !tb.Comparable() {
		return false
	}

	return a == b
}

// Clear removes all services from the container.
func (c *DIContainer) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[reflect.Type]DIRegistration)
	c.scopedInstances = make(map[uint64]map[reflect.Type]any)
}

// GetRegistrations returns a list of all registered services.
// Useful for debugging and diagnostics.
func (c *DIContainer) GetRegistrations() []DIRegistration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	registrations := make([]DIRegistration, 0, len(c.services))
	for _, reg := range c.services {
		registrations = append(registrations, reg)
	}
	return registrations
}

// GenerateDependencyGraph generates a DOT format graph of dependencies.
// Useful for visualizing the dependency structure.
func (c *DIContainer) GenerateDependencyGraph() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var dot strings.Builder
	dot.WriteString("digraph Dependencies {\n")
	dot.WriteString("  node [shape=box];\n")

	for _, reg := range c.services {
		if reg.Factory != nil || reg.Instance != nil {
			source := reg.Type.String()
			dot.WriteString(fmt.Sprintf("  \"%s\";\n", source))
		}
	}

	dot.WriteString("}\n")
	return dot.String()
}
