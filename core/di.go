package core

import (
	"errors"
	"fmt"
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

	if serviceType.Kind() == reflect.Interface {
		if service, err := c.resolveAssignable(serviceType); err == nil {
			return service, nil
		} else {
			return nil, err
		}
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

		injectTag, ok := field.Tag.Lookup("inject")
		if ok {
			// Check if field is exportable
			if !fieldVal.CanSet() {
				continue
			}

			if injectTag == "-" {
				continue
			}

			// Resolve the dependency
			var dep interface{}
			var err error

			if injectTag != "" {
				dep, err = c.resolveByName(injectTag)
				if err != nil {
					dep, err = c.Resolve(field.Type)
				}
			} else {
				dep, err = c.Resolve(field.Type)
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

func (c *DIContainer) resolveByName(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var match interface{}
	for t, service := range c.services {
		if t.Name() == name || t.String() == name {
			if match == nil {
				match = service
				continue
			}
			if servicesEqual(match, service) {
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

func (c *DIContainer) resolveAssignable(serviceType reflect.Type) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var match interface{}
	for t, service := range c.services {
		if t.AssignableTo(serviceType) {
			if match == nil {
				match = service
				continue
			}
			if servicesEqual(match, service) {
				continue
			}
			return nil, fmt.Errorf("multiple services match interface: %s", serviceType.String())
		}
	}

	if match == nil {
		return nil, errors.New("service not found: " + serviceType.String())
	}
	return match, nil
}

func servicesEqual(a, b interface{}) bool {
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

	c.services = make(map[reflect.Type]interface{})
	c.factories = make(map[reflect.Type]func(*DIContainer) interface{})
}
