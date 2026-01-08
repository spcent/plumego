package core

import (
	"errors"
	"reflect"
	"sync"
)

// DIContainer is a simple dependency injection container.
type DIContainer struct {
	mu        sync.RWMutex
	services  map[reflect.Type]interface{}
	factories map[reflect.Type]func(*DIContainer) interface{}
}

// NewDIContainer creates a new dependency injection container.
func NewDIContainer() *DIContainer {
	return &DIContainer{
		services:  make(map[reflect.Type]interface{}),
		factories: make(map[reflect.Type]func(*DIContainer) interface{}),
	}
}

// Register registers a service instance in the container.
func (c *DIContainer) Register(service interface{}) {
	if service == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Register service by its type
	serviceType := reflect.TypeOf(service)
	c.services[serviceType] = service

	// Also register by pointer type if it's a pointer
	if serviceType.Kind() == reflect.Ptr {
		originalType := serviceType.Elem()
		c.services[originalType] = service
	}
}

// RegisterFactory registers a factory function to create a service.
func (c *DIContainer) RegisterFactory(serviceType reflect.Type, factory func(*DIContainer) interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[serviceType] = factory
}

// Resolve resolves a service by its type.
func (c *DIContainer) Resolve(serviceType reflect.Type) (interface{}, error) {
	c.mu.RLock()
	// Check if service is already resolved
	if service, exists := c.services[serviceType]; exists {
		c.mu.RUnlock()
		return service, nil
	}

	// Check if there's a factory for this service
	factory, exists := c.factories[serviceType]
	c.mu.RUnlock()

	if exists {
		// Create service using factory
		service := factory(c)

		// Register the created service for future use
		c.Register(service)
		return service, nil
	}

	return nil, errors.New("service not found: " + serviceType.String())
}

// Inject injects dependencies into a struct.
// It looks for fields with the `inject` tag and tries to resolve them.
func (c *DIContainer) Inject(instance interface{}) error {
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

		// Check if field has inject tag
		if injectTag := field.Tag.Get("inject"); injectTag != "" {
			// Check if field is exportable
			if !fieldVal.CanSet() {
				continue
			}

			// Resolve the dependency
			var dep interface{}
			var err error

			if injectTag != "" {
				// Try to resolve by type name if tag is specified
				// This is a simplified implementation, in a real DI container we might
				// support named dependencies with more sophisticated tag parsing
				if depType, ok := parseTypeName(injectTag); ok {
					dep, err = c.Resolve(depType)
				} else {
					// Fallback to field type
					dep, err = c.Resolve(field.Type)
				}
			} else {
				// Resolve by field type
				dep, err = c.Resolve(field.Type)
			}

			if err != nil {
				return err
			}

			// Set the field value
			fieldVal.Set(reflect.ValueOf(dep))
		}
	}

	return nil
}

// Helper function to parse type name (simplified)
func parseTypeName(name string) (reflect.Type, bool) {
	// This is a simplified implementation. In a real DI container,
	// we would have a more sophisticated type name parsing mechanism.
	// For now, we just return false to fallback to field type.
	return nil, false
}

// Clear removes all services from the container.
func (c *DIContainer) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[reflect.Type]interface{})
	c.factories = make(map[reflect.Type]func(*DIContainer) interface{})
}
