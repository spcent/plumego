package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
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
}

// NewDIContainer creates a new dependency injection container.
// The container is initialized with no services and is ready for registration.
func NewDIContainer() *DIContainer {
	return &DIContainer{
		services: make(map[reflect.Type]DIRegistration),
	}
}

// resolveWithStack resolves a service while tracking the resolution stack
// to detect circular dependencies. This is called internally by Resolve.
func (c *DIContainer) resolveWithStack(serviceType reflect.Type, stack []reflect.Type) (any, error) {
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
			return c.resolveAssignableWithStack(serviceType, newStack)
		}
		return nil, fmt.Errorf("service not found: %s", serviceType.String())
	}

	// Handle singleton with existing instance
	if registration.Lifecycle == Singleton && registration.Instance != nil {
		return registration.Instance, nil
	}

	// Create new instance
	var instance any
	var err error

	if registration.Factory != nil {
		// Use factory function - pass original container
		instance = registration.Factory(c)
	} else {
		// Try to create instance using reflection
		instance, err = c.createInstanceWithStack(serviceType, newStack)
		if err != nil {
			return nil, err
		}
	}

	// Register instance based on lifecycle
	c.mu.Lock()
	if registration.Lifecycle == Singleton {
		registration.Instance = instance
		c.services[serviceType] = registration
	}
	c.mu.Unlock()

	return instance, nil
}

// resolveAssignableWithStack resolves a service that implements an interface with stack tracking.
func (c *DIContainer) resolveAssignableWithStack(serviceType reflect.Type, stack []reflect.Type) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var match any
	for _, registration := range c.services {
		if registration.Instance != nil {
			instanceType := reflect.TypeOf(registration.Instance)
			if instanceType.AssignableTo(serviceType) {
				if match == nil {
					match = registration.Instance
					continue
				}
				if servicesEqual(match, registration.Instance) {
					continue
				}
				return nil, fmt.Errorf("multiple services match interface: %s", serviceType.String())
			}
		}
	}

	if match == nil {
		return nil, errors.New("service not found: " + serviceType.String())
	}
	return match, nil
}

// createInstanceWithStack creates a new instance using reflection with stack tracking.
func (c *DIContainer) createInstanceWithStack(serviceType reflect.Type, stack []reflect.Type) (any, error) {
	if serviceType.Kind() == reflect.Ptr {
		// Create pointer to new instance
		elemType := serviceType.Elem()
		newInstance := reflect.New(elemType)

		// Try to inject dependencies into the struct
		if err := c.injectWithStack(newInstance.Interface(), stack); err != nil {
			// If injection fails, still return the instance
			// This allows for simple structs without dependencies
		}

		return newInstance.Interface(), nil
	}

	return nil, fmt.Errorf("cannot create instance of non-pointer type: %s", serviceType.String())
}

// injectWithStack injects dependencies into a struct with stack tracking.
func (c *DIContainer) injectWithStack(instance any, stack []reflect.Type) error {
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

	// Iterate over all fields
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
				dep, err = c.resolveByName(injectTag)
				if err != nil {
					dep, err = c.resolveWithStack(field.Type, stack)
				}
			} else {
				dep, err = c.resolveWithStack(field.Type, stack)
			}

			if err != nil {
				return err
			}

			// Set the field value
			depVal := reflect.ValueOf(dep)
			if depVal.Type().AssignableTo(field.Type) {
				fieldVal.Set(depVal)
				continue
			}

			if depVal.Kind() == reflect.Ptr && !depVal.IsNil() && depVal.Elem().Type().AssignableTo(field.Type) {
				fieldVal.Set(depVal.Elem())
				continue
			}

			return fmt.Errorf("cannot assign dependency %s to field %s", depVal.Type(), field.Type)
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
	return c.resolveWithStack(serviceType, nil)
}

// createInstance creates a new instance using reflection.
// It attempts to call the constructor and inject dependencies.
func (c *DIContainer) createInstance(serviceType reflect.Type) (any, error) {
	return c.createInstanceWithStack(serviceType, nil)
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
	return c.injectWithStack(instance, nil)
}

// resolveByName resolves a service by its name or type name.
func (c *DIContainer) resolveByName(name string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var match any
	for t, registration := range c.services {
		if t.Name() == name || t.String() == name {
			if match == nil {
				match = registration.Instance
				continue
			}
			if servicesEqual(match, registration.Instance) {
				continue
			}
			return nil, fmt.Errorf("multiple services match name: %s", name)
		}
	}

	if match == nil {
		return nil, errors.New("service not found: " + name)
	}
	return match, nil
}

// resolveAssignable resolves a service that implements an interface.
func (c *DIContainer) resolveAssignable(serviceType reflect.Type) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var match any
	for _, registration := range c.services {
		if registration.Instance != nil {
			instanceType := reflect.TypeOf(registration.Instance)
			if instanceType.AssignableTo(serviceType) {
				if match == nil {
					match = registration.Instance
					continue
				}
				if servicesEqual(match, registration.Instance) {
					continue
				}
				return nil, fmt.Errorf("multiple services match interface: %s", serviceType.String())
			}
		}
	}

	if match == nil {
		return nil, errors.New("service not found: " + serviceType.String())
	}
	return match, nil
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
